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

// AzureAdaptor implements Adaptor for Azure OpenAI.
// Uses api-key header and deployment-based endpoint.
type AzureAdaptor struct {
	channelType int
}

func (a *AzureAdaptor) Init(info *RelayInfo) {
	a.channelType = info.ChannelType
}

func (a *AzureAdaptor) GetRequestURL(info *RelayInfo) (string, error) {
	baseUrl := info.BaseUrl
	if baseUrl == "" {
		return "", fmt.Errorf("base_url required for Azure channel")
	}
	baseUrl = strings.TrimRight(baseUrl, "/")

	// Azure endpoint format:
	// https://{resource}.openai.azure.com/openai/deployments/{deployment}/chat/completions?api-version=2024-02-15-preview
	// Deployment name is stored as part of model mapping or UpstreamModelName
	return fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=2024-02-15-preview",
		baseUrl, info.UpstreamModelName), nil
}

func (a *AzureAdaptor) SetupRequestHeader(req *http.Header, info *RelayInfo) error {
	req.Set("Content-Type", "application/json")
	req.Set("api-key", info.ApiKey)
	return nil
}

func (a *AzureAdaptor) ConvertRequest(req *model.ChatRequest) (any, error) {
	return req, nil
}

func (a *AzureAdaptor) DoRequest(ctx context.Context, info *RelayInfo, requestBody io.Reader) (*http.Response, error) {
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

func (a *AzureAdaptor) DoResponse(r *ghttp.Request, resp *http.Response, info *RelayInfo) (*model.Usage, error) {
	if info.IsStream {
		return a.handleStreamingResponse(r.Context(), r, resp)
	}
	return a.handleNonStreamingResponse(r, resp)
}

func (a *AzureAdaptor) handleNonStreamingResponse(r *ghttp.Request, resp *http.Response) (*model.Usage, error) {
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		r.Response.WriteHeader(resp.StatusCode)
		r.Response.Writeln(string(respBody))
		return nil, fmt.Errorf("upstream error (status %d)", resp.StatusCode)
	}

	r.Response.Header().Set("Content-Type", "application/json")
	r.Response.Writeln(string(respBody))

	// Parse usage from response (same OpenAI format)
	var upstreamResp model.ChatCompletionResponse
	if err := json.Unmarshal(respBody, &upstreamResp); err == nil {
		return &upstreamResp.Usage, nil
	}

	return nil, nil
}

func (a *AzureAdaptor) handleStreamingResponse(ctx context.Context, r *ghttp.Request, resp *http.Response) (*model.Usage, error) {
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	r.Response.Header().Set("Content-Type", "text/event-stream")
	r.Response.Header().Set("Cache-Control", "no-cache")
	r.Response.Header().Set("Connection", "keep-alive")
	r.Response.WriteHeader(http.StatusOK)

	var totalPromptTokens, totalCompletionTokens int

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

func (a *AzureAdaptor) GetModelList() []string {
	return []string{
		"gpt-4", "gpt-4-32k", "gpt-4-turbo", "gpt-4o", "gpt-4o-mini",
		"gpt-35-turbo", "gpt-35-turbo-16k",
	}
}

func (a *AzureAdaptor) GetChannelName() string {
	return "Azure OpenAI"
}
