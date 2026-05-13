package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ai-gateway/internal/model"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

// GeminiAdaptor implements Adaptor for Google Gemini.
type GeminiAdaptor struct {
	channelType int
}

func (a *GeminiAdaptor) Init(info *RelayInfo) {
	a.channelType = info.ChannelType
}

func (a *GeminiAdaptor) GetRequestURL(info *RelayInfo) (string, error) {
	baseUrl := info.BaseUrl
	if baseUrl == "" {
		baseUrl = "https://generativelanguage.googleapis.com"
	}
	baseUrl = strings.TrimRight(baseUrl, "/")

	// Map upstream model name to Gemini model ID
	modelName := info.UpstreamModelName

	// Remove 'gemini-' prefix if present (we store as 'gemini-pro' in DB)
	geminiModel := strings.TrimPrefix(modelName, "gemini-")
	geminiModel = "gemini-" + geminiModel

	if info.IsStream {
		return fmt.Sprintf("%s/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s", baseUrl, geminiModel, info.ApiKey), nil
	}
	return fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", baseUrl, geminiModel, info.ApiKey), nil
}

func (a *GeminiAdaptor) SetupRequestHeader(req *http.Header, info *RelayInfo) error {
	req.Set("Content-Type", "application/json")
	return nil
}

type geminiPart struct {
	Text string `json:"text,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiRequest struct {
	Contents              []geminiContent       `json:"contents"`
	SystemInstruction     *geminiContent         `json:"system_instruction,omitempty"`
	GenerationConfig      *geminiGenerationConfig `json:"generation_config,omitempty"`
}

type geminiGenerationConfig struct {
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	Temperature     float64 `json:"temperature,omitempty"`
}

func (a *GeminiAdaptor) ConvertRequest(req *model.ChatRequest) (any, error) {
	var contents []geminiContent
	var systemContent string

	for _, msg := range req.Messages {
		switch msg.Role {
		case "system":
			systemContent = msg.Content
		case "assistant":
			contents = append(contents, geminiContent{
				Role:  "model",
				Parts: []geminiPart{{Text: msg.Content}},
			})
		default:
			contents = append(contents, geminiContent{
				Role:  "user",
				Parts: []geminiPart{{Text: msg.Content}},
			})
		}
	}

	if len(contents) == 0 {
		contents = append(contents, geminiContent{
			Role:  "user",
			Parts: []geminiPart{{Text: "Hello"}},
		})
	}

	gReq := &geminiRequest{Contents: contents}

	if systemContent != "" {
		gReq.SystemInstruction = &geminiContent{
			Parts: []geminiPart{{Text: systemContent}},
		}
	}

	if req.MaxTokens != nil || req.Temperature != nil {
		cfg := &geminiGenerationConfig{}
		if req.MaxTokens != nil && *req.MaxTokens > 0 {
			cfg.MaxOutputTokens = *req.MaxTokens
		}
		if req.Temperature != nil {
			cfg.Temperature = *req.Temperature
		}
		gReq.GenerationConfig = cfg
	}

	return gReq, nil
}

func (a *GeminiAdaptor) DoRequest(ctx context.Context, info *RelayInfo, requestBody io.Reader) (*http.Response, error) {
	url, err := a.GetRequestURL(info)
	if err != nil {
		return nil, err
	}
	upstreamReq, err := http.NewRequestWithContext(ctx, "POST", url, requestBody)
	if err != nil {
		return nil, fmt.Errorf("create upstream request: %w", err)
	}
	a.SetupRequestHeader(&upstreamReq.Header, info)
	client := &http.Client{Timeout: 300 * time.Second}
	return client.Do(upstreamReq)
}

func (a *GeminiAdaptor) DoResponse(r *ghttp.Request, resp *http.Response, info *RelayInfo) (*model.Usage, error) {
	if info.IsStream {
		return a.handleStreamingResponse(r.Context(), r, resp, info)
	}
	return a.handleNonStreamingResponse(r, resp, info)
}

type geminiCandidate struct {
	Content       geminiContent `json:"content"`
	FinishReason  string        `json:"finishReason"`
}

type geminiUsage struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

type geminiResponse struct {
	Candidates    []geminiCandidate `json:"candidates"`
	UsageMetadata geminiUsage       `json:"usageMetadata"`
}

func (a *GeminiAdaptor) handleNonStreamingResponse(r *ghttp.Request, resp *http.Response, info *RelayInfo) (*model.Usage, error) {
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		r.Response.WriteHeader(resp.StatusCode)
		r.Response.Writeln(string(respBody))
		return nil, fmt.Errorf("upstream error (status %d)", resp.StatusCode)
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		r.Response.Header().Set("Content-Type", "application/json")
		r.Response.Writeln(string(respBody))
		return nil, fmt.Errorf("parse response: %w", err)
	}

	// Extract text from candidates
	content := ""
	if len(geminiResp.Candidates) > 0 {
		for _, part := range geminiResp.Candidates[0].Content.Parts {
			content += part.Text
		}
	}

	finishReason := geminiResp.Candidates[0].FinishReason
	if finishReason == "STOP" {
		finishReason = "stop"
	} else if finishReason != "" {
		finishReason = strings.ToLower(finishReason)
	}

	openAIResp := model.ChatCompletionResponse{
		Id:      fmt.Sprintf("gemini-%d", time.Now().UnixNano()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   info.UpstreamModelName,
		Choices: []struct {
			Index        int         `json:"index"`
			Message      model.ChatMessage `json:"message"`
			FinishReason string      `json:"finish_reason"`
		}{
			{
				Index:        0,
				Message:      model.ChatMessage{Role: "assistant", Content: content},
				FinishReason: finishReason,
			},
		},
		Usage: model.Usage{
			PromptTokens:     geminiResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: geminiResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      geminiResp.UsageMetadata.TotalTokenCount,
		},
	}

	r.Response.Header().Set("Content-Type", "application/json")
	r.Response.WriteJson(openAIResp)

	return &openAIResp.Usage, nil
}

func (a *GeminiAdaptor) handleStreamingResponse(ctx context.Context, r *ghttp.Request, resp *http.Response, info *RelayInfo) (*model.Usage, error) {
	var totalPromptTokens, totalCompletionTokens int
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	r.Response.Header().Set("Content-Type", "text/event-stream")
	r.Response.Header().Set("Cache-Control", "no-cache")
	r.Response.Header().Set("Connection", "keep-alive")
	r.Response.WriteHeader(http.StatusOK)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Gemini SSE: "data: {...}"
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := line[6:]

		if data == "[DONE]" {
			break
		}

		var geminiResp geminiResponse
		if err := json.Unmarshal([]byte(data), &geminiResp); err != nil {
			continue
		}

		if len(geminiResp.Candidates) == 0 {
			continue
		}

		content := ""
		for _, part := range geminiResp.Candidates[0].Content.Parts {
			content += part.Text
		}

		if geminiResp.UsageMetadata.PromptTokenCount > 0 {
			totalPromptTokens = geminiResp.UsageMetadata.PromptTokenCount
			totalCompletionTokens = geminiResp.UsageMetadata.CandidatesTokenCount
		}

		var finishReason *string
		fr := geminiResp.Candidates[0].FinishReason
		if fr == "STOP" {
			s := "stop"
			finishReason = &s
		} else if fr != "" {
			s := strings.ToLower(fr)
			finishReason = &s
		}

		deltaRole := ""
		if content != "" && finishReason == nil {
			deltaRole = ""
		}

		chunk := model.ChatCompletionChunk{
			Id:      info.RequestId,
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   info.UpstreamModelName,
			Choices: []struct {
				Index        int         `json:"index"`
				Delta        model.ChatMessage `json:"delta"`
				FinishReason *string     `json:"finish_reason,omitempty"`
			}{
				{
					Index:        0,
					Delta:        model.ChatMessage{Role: deltaRole, Content: content},
					FinishReason: finishReason,
				},
			},
		}
		dataBytes, _ := json.Marshal(chunk)
		r.Response.Writeln("data: " + string(dataBytes))
		r.Response.Flush()
	}

	if err := scanner.Err(); err != nil {
		glog.Errorf(ctx, "stream read error: %v", err)
	}

	r.Response.Writeln("data: [DONE]")
	r.Response.Flush()

	return &model.Usage{
		PromptTokens:     totalPromptTokens,
		CompletionTokens: totalCompletionTokens,
		TotalTokens:      totalPromptTokens + totalCompletionTokens,
	}, nil
}

func (a *GeminiAdaptor) GetModelList() []string {
	return []string{
		"gemini-pro", "gemini-2.0-flash", "gemini-2.0-flash-lite",
		"gemini-1.5-pro", "gemini-1.5-flash",
	}
}

func (a *GeminiAdaptor) GetChannelName() string {
	return "Gemini"
}
