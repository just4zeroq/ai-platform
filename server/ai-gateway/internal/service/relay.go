package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"ai-gateway/internal/grpcclient"
	"ai-gateway/internal/model"

	assetpb "api/asset/v1"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

// --- Channel Selection ---

// getAbilities returns enabled abilities for model+group, ordered by priority.
func getAbilities(ctx context.Context, modelName, group string) (gdb.Result, error) {
	abilities, err := g.DB().Model("abilities").Ctx(ctx).
		Where("model", modelName).
		Where("group_name", group).
		Where("enabled", true).
		Order("priority DESC, weight DESC").
		All()
	if err != nil {
		return nil, err
	}
	if abilities.Len() > 0 {
		return abilities, nil
	}

	// No exact match — try LIKE
	return g.DB().Model("abilities m").Ctx(ctx).
		Where("group_name", group).
		Where("enabled", true).
		Where("? LIKE '%' || m.model || '%'", modelName).
		Order("priority DESC, weight DESC").
		All()
}

// getChannel fetches a single channel by id if enabled.
func getChannel(ctx context.Context, channelId int) (*model.Channel, error) {
	ch, err := g.DB().Model("channels").Ctx(ctx).
		Where("id", channelId).
		Where("status", 1).
		Where("deleted_at IS NULL").
		One()
	if err != nil || ch == nil {
		return nil, err
	}
	return &model.Channel{
		Id:           ch["id"].Int(),
		Type:         ch["type"].Int(),
		Key:          ch["key"].String(),
		Name:         ch["name"].String(),
		Models:       ch["models"].String(),
		Group:        ch["group_name"].String(),
		Status:       ch["status"].Int(),
		Priority:     ch["priority"].Int64(),
		Weight:       uint(ch["weight"].Uint()),
		ModelMapping: ch["model_mapping"].String(),
		BaseUrl:      ch["base_url"].String(),
	}, nil
}

// getWeightedChannel picks a channel using weighted random from abilities.
func getWeightedChannel(ctx context.Context, abilities gdb.Result) (*model.Channel, error) {
	type candidate struct {
		channelId int
		weight    int
	}

	var candidates []candidate
	totalWeight := 0

	for _, a := range abilities {
		channelId := a["channel_id"].Int()
		ch, err := getChannel(ctx, channelId)
		if err != nil || ch == nil {
			continue
		}
		w := a["weight"].Int()
		if w <= 0 {
			w = 1
		}
		candidates = append(candidates, candidate{channelId, w})
		totalWeight += w
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no enabled channels available")
	}

	roll := rand.Intn(totalWeight)
	for _, c := range candidates {
		roll -= c.weight
		if roll < 0 {
			return getChannel(ctx, c.channelId)
		}
	}

	return getChannel(ctx, candidates[0].channelId)
}

// SelectChannel picks a channel for a model in the user's group using weighted random.
func SelectChannel(ctx context.Context, modelName, group string) (*model.Channel, error) {
	if group == "" {
		group = "default"
	}
	abilities, err := getAbilities(ctx, modelName, group)
	if err != nil {
		return nil, fmt.Errorf("query abilities: %w", err)
	}
	if abilities.Len() == 0 {
		return nil, fmt.Errorf("no available channel for model '%s' in group '%s'", modelName, group)
	}
	return getWeightedChannel(ctx, abilities)
}

// getAllChannels returns all enabled channels for a model (used for retry fallback).
func getAllChannels(ctx context.Context, modelName, group string) ([]*model.Channel, error) {
	if group == "" {
		group = "default"
	}
	abilities, err := getAbilities(ctx, modelName, group)
	if err != nil {
		return nil, err
	}
	var result []*model.Channel
	for _, a := range abilities {
		ch, err := getChannel(ctx, a["channel_id"].Int())
		if err != nil || ch == nil {
			continue
		}
		result = append(result, ch)
	}
	return result, nil
}

// --- Billing ---

// PreDeductQuota reserves quota from user balance.
func PreDeductQuota(ctx context.Context, userId int64, quota int64) (int64, error) {
	if quota <= 0 {
		return 0, nil
	}
	balanceRes, err := grpcclient.AssetSvc.GetBalance(ctx, &assetpb.GetBalanceReq{UserId: userId})
	if err != nil {
		return 0, fmt.Errorf("get balance: %w", err)
	}
	balance := balanceRes.Balance.GetBalance()
	if balance < float64(quota)/1000 {
		return 0, fmt.Errorf("insufficient balance: have %.4f, need %.4f", balance, float64(quota)/1000)
	}
	return quota, nil
}

// ReportUsage reports actual usage and deducts exact quota.
func ReportUsage(ctx context.Context, userId, apiKeyId int64, modelName string, promptTokens, completionTokens int, quota int64, requestId string, channelId int, channelName string) error {
	_, err := grpcclient.AssetSvc.ReportUsage(ctx, &assetpb.ReportUsageReq{
		UserId:           userId,
		ApiKeyId:         apiKeyId,
		ModelName:        modelName,
		PromptTokens:     int32(promptTokens),
		CompletionTokens: int32(completionTokens),
		Quota:            float64(quota) / 1000,
		RequestId:        requestId,
		ChannelId:        int64(channelId),
		ChannelName:      channelName,
	})
	return err
}

// --- relayError for classification ---

type relayError struct {
	errType ErrorType
	message string
}

func (e *relayError) Error() string { return e.message }

// --- Relay Pipeline ---

func doRelay(r *ghttp.Request, chatReq *model.ChatRequest, ch *model.Channel, preQuota int64) (*model.Usage, error) {
	ctx := r.Context()
	userId := r.GetCtxVar("user_id").Int64()
	apiKeyId := r.GetCtxVar("api_key_id").Int64()
	modelName := chatReq.Model

	info := &RelayInfo{
		UserId:            userId,
		ApiKeyId:          apiKeyId,
		Group:             r.GetCtxVar("user_group").String(),
		ChannelId:         ch.Id,
		ChannelType:       ch.Type,
		ChannelName:       ch.Name,
		ApiKey:            ch.Key,
		BaseUrl:           ch.BaseUrl,
		ModelMapping:      ch.ModelMapping,
		UpstreamModelName: modelName,
		IsStream:          chatReq.Stream,
		RequestId:         fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano()),
	}

	adaptor := GetAdaptor(ch.Type)
	adaptor.Init(info)

	adaptedReq, err := adaptor.ConvertRequest(chatReq)
	if err != nil {
		return nil, fmt.Errorf("convert request: %w", err)
	}

	upstreamBody, err := json.Marshal(adaptedReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	upstreamRes, err := adaptor.DoRequest(ctx, info, bytes.NewReader(upstreamBody))
	if err != nil {
		return nil, fmt.Errorf("upstream request: %w", err)
	}
	defer upstreamRes.Body.Close()

	if upstreamRes.StatusCode != http.StatusOK {
		errType := ClassifyError(upstreamRes.StatusCode)
		respBody, _ := io.ReadAll(upstreamRes.Body)

		r.Response.WriteHeader(upstreamRes.StatusCode)
		r.Response.Writeln(string(respBody))

		return nil, &relayError{errType: errType, message: fmt.Sprintf("upstream error (status %d)", upstreamRes.StatusCode)}
	}

	usage, err := adaptor.DoResponse(r, upstreamRes, info)
	if err != nil {
		return nil, fmt.Errorf("response handling: %w", err)
	}

	if usage != nil && (usage.PromptTokens > 0 || usage.CompletionTokens > 0) {
		actualQuota := calcQuota(modelName, usage.PromptTokens, usage.CompletionTokens)
		if actualQuota > 0 {
			if err := ReportUsage(ctx, userId, apiKeyId, modelName, usage.PromptTokens, usage.CompletionTokens, actualQuota, info.RequestId, ch.Id, ch.Name); err != nil {
				glog.Errorf(ctx, "report usage failed: %v", err)
			}
		}
		LogQ.Push(LogEntry{
			UserId:           userId,
			ApiKeyId:         apiKeyId,
			ModelName:        modelName,
			PromptTokens:     usage.PromptTokens,
			CompletionTokens: usage.CompletionTokens,
			Quota:            float64(actualQuota) / 1000,
			RequestId:        info.RequestId,
			ChannelId:        int64(ch.Id),
			ChannelName:      ch.Name,
		})
	}

	return usage, nil
}

const maxRetries = 3

// ChatCompletions handles /v1/chat/completions relay with retry.
func ChatCompletions(r *ghttp.Request) {
	ctx := r.Context()
	userId := r.GetCtxVar("user_id").Int64()
	apiKeyId := r.GetCtxVar("api_key_id").Int64()
	group := r.GetCtxVar("user_group").String()
	modelLimits := r.GetCtxVar("model_limits").Strings()
	modelLimitsEnabled := r.GetCtxVar("model_limits_enabled").Bool()

	body, err := io.ReadAll(r.Request.Body)
	if err != nil {
		r.Response.WriteJson(g.Map{"error": "read body: " + err.Error()})
		return
	}
	r.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	var chatReq model.ChatRequest
	if err := json.Unmarshal(body, &chatReq); err != nil {
		r.Response.WriteJson(g.Map{"error": "parse body: " + err.Error()})
		return
	}

	modelName := chatReq.Model
	if modelName == "" {
		r.Response.WriteJson(g.Map{"error": "model is required"})
		return
	}

	if modelLimitsEnabled && len(modelLimits) > 0 {
		allowed := false
		for _, m := range modelLimits {
			if strings.EqualFold(m, modelName) || strings.Contains(modelName, m) {
				allowed = true
				break
			}
		}
		if !allowed {
			r.Response.WriteJson(g.Map{"error": fmt.Sprintf("model '%s' not allowed by API key limits", modelName)})
			return
		}
	}

	var promptTokens int
	for _, msg := range chatReq.Messages {
		promptTokens += estimateTokens(msg.Content)
	}
	preQuota := calcQuota(modelName, promptTokens, 0) + calcQuota(modelName, 0, 256)

	if preQuota > 0 {
		if _, err := PreDeductQuota(ctx, userId, preQuota); err != nil {
			glog.Errorf(ctx, "pre-deduct failed: %v", err)
			r.Response.WriteJson(g.Map{"error": "insufficient balance"})
			return
		}
	}

	channels, err := getAllChannels(ctx, modelName, group)
	if err != nil || len(channels) == 0 {
		glog.Errorf(ctx, "channel selection failed: %v", err)
		r.Response.WriteJson(g.Map{"error": "no available channel for model '" + modelName + "'"})
		return
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries && attempt < len(channels); attempt++ {
		ch := channels[attempt]

		_, err := doRelay(r, &chatReq, ch, preQuota)
		if err == nil {
			ChannelFailureTrack.RecordSuccess(ch.Id)
			return
		}

		glog.Warningf(ctx, "relay attempt %d failed via channel %s (id=%d): %v", attempt+1, ch.Name, ch.Id, err)
		lastErr = err

		if rErr, ok := err.(*relayError); ok {
			shouldDisable := ChannelFailureTrack.RecordFailure(ch.Id, rErr.errType)
			if shouldDisable {
				glog.Errorf(ctx, "auto-disabling channel %s (id=%d) after failures", ch.Name, ch.Id)
				g.DB().Model("channels").Ctx(ctx).
					Where("id", ch.Id).
					Update(g.Map{"status": 0, "updated_at": time.Now().Unix()})
			}
			if !ShouldRetry(rErr.errType) {
				break
			}
		} else {
			break
		}
	}

	glog.Errorf(ctx, "all relay attempts failed: %v", lastErr)
	if preQuota > 0 {
		ReportUsage(ctx, userId, apiKeyId, modelName, 0, 0, 0, fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano()), 0, "")
	}
}

// ListModels returns available models from enabled channels.
func ListModels(ctx context.Context) ([]map[string]interface{}, error) {
	channels, err := g.DB().Model("channels").Ctx(ctx).
		Where("status", 1).
		Where("deleted_at IS NULL").
		All()
	if err != nil {
		return nil, err
	}

	modelSet := make(map[string]bool)
	var result []map[string]interface{}
	for _, ch := range channels {
		models := strings.Split(ch["models"].String(), ",")
		for _, m := range models {
			m = strings.TrimSpace(m)
			if m == "" || modelSet[m] {
				continue
			}
			modelSet[m] = true
			result = append(result, map[string]interface{}{
				"id":       m,
				"object":   "model",
				"created":  time.Now().Unix(),
				"owned_by": "ai-platform",
			})
		}
	}
	return result, nil
}
