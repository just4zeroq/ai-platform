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

// OpenAIAdaptor implements Adaptor for OpenAI-compatible API providers.
type OpenAIAdaptor struct {
	channelType int
}

func (a *OpenAIAdaptor) Init(info *RelayInfo) {
	a.channelType = info.ChannelType
}

func (a *OpenAIAdaptor) GetRequestURL(info *RelayInfo) (string, error) {
	baseUrl := info.BaseUrl
	if baseUrl == "" {
		baseUrl = "https://api.openai.com"
	}
	return strings.TrimRight(baseUrl, "/") + "/v1/chat/completions", nil
}

func (a *OpenAIAdaptor) SetupRequestHeader(req *http.Header, info *RelayInfo) error {
	req.Set("Content-Type", "application/json")
	req.Set("Authorization", "Bearer "+info.ApiKey)
	return nil
}

func (a *OpenAIAdaptor) ConvertRequest(req *model.ChatRequest) (any, error) {
	return req, nil
}

func (a *OpenAIAdaptor) DoRequest(ctx context.Context, info *RelayInfo, requestBody io.Reader) (*http.Response, error) {
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

// DoResponse handles both streaming and non-streaming responses.
func (a *OpenAIAdaptor) DoResponse(r *ghttp.Request, resp *http.Response, info *RelayInfo) (*model.Usage, error) {
	if info.IsStream {
		return a.handleStreamingResponse(r.Context(), r, resp)
	}
	return a.handleNonStreamingResponse(r, resp)
}

func (a *OpenAIAdaptor) handleNonStreamingResponse(r *ghttp.Request, resp *http.Response) (*model.Usage, error) {
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		r.Response.WriteHeader(resp.StatusCode)
		r.Response.Writeln(string(respBody))
		return nil, fmt.Errorf("upstream error (status %d)", resp.StatusCode)
	}

	var upstreamResp model.ChatCompletionResponse
	if err := json.Unmarshal(respBody, &upstreamResp); err != nil {
		r.Response.Header().Set("Content-Type", "application/json")
		r.Response.Writeln(string(respBody))
		return nil, fmt.Errorf("parse response: %w", err)
	}

	// Forward response to client
	r.Response.Header().Set("Content-Type", "application/json")
	r.Response.WriteJson(upstreamResp)

	return &upstreamResp.Usage, nil
}

func (a *OpenAIAdaptor) handleStreamingResponse(ctx context.Context, r *ghttp.Request, resp *http.Response) (*model.Usage, error) {
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
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := line[6:]
		if data == "[DONE]" {
			break
		}

		var chunk model.ChatCompletionChunk
		if err := json.Unmarshal([]byte(data), &chunk); err == nil {
			if chunk.Usage != nil {
				totalPromptTokens = chunk.Usage.PromptTokens
				totalCompletionTokens = chunk.Usage.CompletionTokens
			}
		}

		r.Response.Writeln(line)
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

func (a *OpenAIAdaptor) GetModelList() []string {
	return []string{
		"gpt-4", "gpt-4-32k", "gpt-4-1106-preview", "gpt-4-turbo",
		"gpt-4o", "gpt-4o-mini", "gpt-4o-mini-2024-07-18",
		"gpt-3.5-turbo", "gpt-3.5-turbo-1106",
	}
}

func (a *OpenAIAdaptor) GetChannelName() string {
	return "OpenAI"
}
