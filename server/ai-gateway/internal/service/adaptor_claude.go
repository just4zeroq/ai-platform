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

// ClaudeAdaptor implements Adaptor for Anthropic Claude.
type ClaudeAdaptor struct {
	channelType int
}

func (a *ClaudeAdaptor) Init(info *RelayInfo) {
	a.channelType = info.ChannelType
}

func (a *ClaudeAdaptor) GetRequestURL(info *RelayInfo) (string, error) {
	baseUrl := info.BaseUrl
	if baseUrl == "" {
		baseUrl = "https://api.anthropic.com"
	}
	return strings.TrimRight(baseUrl, "/") + "/v1/messages", nil
}

func (a *ClaudeAdaptor) SetupRequestHeader(req *http.Header, info *RelayInfo) error {
	req.Set("Content-Type", "application/json")
	req.Set("x-api-key", info.ApiKey)
	req.Set("anthropic-version", "2023-06-01")
	return nil
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []claudeMessage `json:"messages"`
	Stream    bool            `json:"stream,omitempty"`
}

func (a *ClaudeAdaptor) ConvertRequest(req *model.ChatRequest) (any, error) {
	claudeMsgs := make([]claudeMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		role := m.Role
		if role == "system" {
			role = "user"
		}
		claudeMsgs = append(claudeMsgs, claudeMessage{Role: role, Content: m.Content})
	}

	maxTokens := 4096
	if req.MaxTokens != nil && *req.MaxTokens > 0 {
		maxTokens = *req.MaxTokens
	}

	return &claudeRequest{
		Model:     req.Model,
		MaxTokens: maxTokens,
		Messages:  claudeMsgs,
		Stream:    req.Stream,
	}, nil
}

func (a *ClaudeAdaptor) DoRequest(ctx context.Context, info *RelayInfo, requestBody io.Reader) (*http.Response, error) {
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

func (a *ClaudeAdaptor) DoResponse(r *ghttp.Request, resp *http.Response, info *RelayInfo) (*model.Usage, error) {
	if info.IsStream {
		return a.handleStreamingResponse(r.Context(), r, resp, info)
	}
	return a.handleNonStreamingResponse(r, resp, info)
}

type claudeContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type claudeResponse struct {
	Id      string              `json:"id"`
	Type    string              `json:"type"`
	Role    string              `json:"role"`
	Content []claudeContentBlock `json:"content"`
	Usage   struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	StopReason string `json:"stop_reason"`
}

func (a *ClaudeAdaptor) handleNonStreamingResponse(r *ghttp.Request, resp *http.Response, info *RelayInfo) (*model.Usage, error) {
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		r.Response.WriteHeader(resp.StatusCode)
		r.Response.Writeln(string(respBody))
		return nil, fmt.Errorf("upstream error (status %d)", resp.StatusCode)
	}

	var claudeResp claudeResponse
	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		r.Response.Header().Set("Content-Type", "application/json")
		r.Response.Writeln(string(respBody))
		return nil, fmt.Errorf("parse response: %w", err)
	}

	// Convert Claude → OpenAI format
	content := ""
	for _, block := range claudeResp.Content {
		if block.Type == "text" {
			content += block.Text
		}
	}

	finishReason := claudeResp.StopReason
	if finishReason == "end_turn" {
		finishReason = "stop"
	}

	openAIResp := model.ChatCompletionResponse{
		Id:      claudeResp.Id,
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
			PromptTokens:     claudeResp.Usage.InputTokens,
			CompletionTokens: claudeResp.Usage.OutputTokens,
			TotalTokens:      claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens,
		},
	}

	r.Response.Header().Set("Content-Type", "application/json")
	r.Response.WriteJson(openAIResp)

	return &openAIResp.Usage, nil
}

func (a *ClaudeAdaptor) handleStreamingResponse(ctx context.Context, r *ghttp.Request, resp *http.Response, info *RelayInfo) (*model.Usage, error) {
	var totalPromptTokens, totalCompletionTokens int
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	r.Response.Header().Set("Content-Type", "text/event-stream")
	r.Response.Header().Set("Cache-Control", "no-cache")
	r.Response.Header().Set("Connection", "keep-alive")
	r.Response.WriteHeader(http.StatusOK)

	var contentBuffer string

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "event: ") && !strings.HasPrefix(line, "data: ") {
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			data := line[6:]

			// Parse Claude SSE data
			var event struct {
				Type string `json:"type"`
				Delta struct {
					Text string `json:"text"`
				} `json:"delta,omitempty"`
				ContentBlock *struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content_block,omitempty"`
				Message *struct {
					Usage struct {
						InputTokens  int `json:"input_tokens"`
						OutputTokens int `json:"output_tokens"`
					} `json:"usage,omitempty"`
				} `json:"message,omitempty"`
			}

			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}

			switch event.Type {
			case "content_block_delta":
				if event.Delta.Text != "" {
					contentBuffer += event.Delta.Text
					// Convert to OpenAI chunk
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
								Index: 0,
								Delta: model.ChatMessage{Role: "", Content: event.Delta.Text},
							},
						},
					}
					dataBytes, _ := json.Marshal(chunk)
					r.Response.Writeln("data: " + string(dataBytes))
					r.Response.Flush()
				}

			case "message_delta":
				if event.Message != nil {
					totalPromptTokens = event.Message.Usage.InputTokens
					totalCompletionTokens = event.Message.Usage.OutputTokens
				}

				// Final chunk
				finishReason := "stop"
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
							Delta:        model.ChatMessage{},
							FinishReason: &finishReason,
						},
					},
				}
				dataBytes, _ := json.Marshal(chunk)
				r.Response.Writeln("data: " + string(dataBytes))
				r.Response.Flush()
			}
		}
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

func (a *ClaudeAdaptor) GetModelList() []string {
	return []string{
		"claude-3-opus-20240229", "claude-3-sonnet-20240229",
		"claude-3-haiku-20240307", "claude-3-5-sonnet-20241022",
		"claude-3-7-sonnet-20250219",
	}
}

func (a *ClaudeAdaptor) GetChannelName() string {
	return "Claude"
}
