package service

import (
	"context"
	"io"
	"net/http"

	"ai-gateway/internal/model"

	"github.com/gogf/gf/v2/net/ghttp"
)

// RelayInfo holds per-request relay metadata set by middleware and channel selection.
type RelayInfo struct {
	UserId     int64
	ApiKeyId   int64
	Group      string

	ChannelId    int
	ChannelType  int
	ChannelName  string
	ApiKey       string
	BaseUrl      string
	ModelMapping string

	UpstreamModelName string
	IsStream          bool
	PreConsumedQuota  int64
	RequestId         string
}

// Adaptor interface — simplified for chat completions only.
// Each LLM provider channel type implements this interface.
type Adaptor interface {
	// Init per-request setup (called once per relay).
	Init(info *RelayInfo)
	// GetRequestURL builds upstream URL for this provider.
	GetRequestURL(info *RelayInfo) (string, error)
	// SetupRequestHeader sets auth and content-type headers.
	SetupRequestHeader(req *http.Header, info *RelayInfo) error
	// ConvertRequest transforms an OpenAI-format request to provider format.
	ConvertRequest(req *model.ChatRequest) (any, error)
	// DoRequest sends HTTP request to the upstream provider.
	DoRequest(ctx context.Context, info *RelayInfo, requestBody io.Reader) (*http.Response, error)
	// DoResponse reads upstream response, writes to client, returns token usage.
	DoResponse(r *ghttp.Request, resp *http.Response, info *RelayInfo) (usage *model.Usage, err error)
	// GetModelList returns model IDs this adaptor can handle.
	GetModelList() []string
	// GetChannelName returns display name for this channel type.
	GetChannelName() string
}

// GetAdaptor returns the Adaptor for the given channel type.
// Types: 1=OpenAI, 2=Azure, 3=Custom, 4=Claude, 5=DeepSeek, 6=Gemini, ...
func GetAdaptor(channelType int) Adaptor {
	switch channelType {
	case 2:
		return &AzureAdaptor{}
	case 4:
		return &ClaudeAdaptor{}
	case 6:
		return &GeminiAdaptor{}
	default:
		return &OpenAIAdaptor{}
	}
}
