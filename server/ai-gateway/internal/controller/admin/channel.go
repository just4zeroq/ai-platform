package admin

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	assetv1 "api/asset/v1"
	"ai-gateway/internal/grpcclient"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func ListChannels(r *ghttp.Request) {
	rows, err := g.DB().Model("channels").Ctx(r.Context()).
		Where("deleted_at IS NULL").
		Order("id ASC").
		All()
	if err != nil {
		r.Response.WriteJson(g.Map{"error": err.Error()})
		return
	}
	r.Response.WriteJson(rows)
}

func CreateChannel(r *ghttp.Request) {
	var data map[string]interface{}
	if err := r.Parse(&data); err != nil {
		r.Response.WriteJson(g.Map{"error": "parse failed: " + err.Error()})
		return
	}
	data["created_at"] = time.Now().Unix()
	data["updated_at"] = time.Now().Unix()

	result, err := g.DB().Model("channels").Ctx(r.Context()).Insert(data)
	if err != nil {
		r.Response.WriteJson(g.Map{"error": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	r.Response.WriteJson(g.Map{"id": id, "message": "channel created"})
}

func UpdateChannel(r *ghttp.Request) {
	id := r.Get("id").Int()
	if id <= 0 {
		r.Response.WriteJson(g.Map{"error": "invalid id"})
		return
	}

	var data map[string]interface{}
	if err := r.Parse(&data); err != nil {
		r.Response.WriteJson(g.Map{"error": "parse failed: " + err.Error()})
		return
	}
	delete(data, "id")
	delete(data, "created_at")
	data["updated_at"] = time.Now().Unix()

	_, err := g.DB().Model("channels").Ctx(r.Context()).
		Where("id", id).
		Where("deleted_at IS NULL").
		Update(data)
	if err != nil {
		r.Response.WriteJson(g.Map{"error": err.Error()})
		return
	}
	r.Response.WriteJson(g.Map{"message": "channel updated"})
}

func DeleteChannel(r *ghttp.Request) {
	id := r.Get("id").Int()
	if id <= 0 {
		r.Response.WriteJson(g.Map{"error": "invalid id"})
		return
	}

	_, err := g.DB().Model("channels").Ctx(r.Context()).
		Where("id", id).
		Where("deleted_at IS NULL").
		Update(g.Map{
			"deleted_at": time.Now().Unix(),
			"updated_at": time.Now().Unix(),
		})
	if err != nil {
		r.Response.WriteJson(g.Map{"error": err.Error()})
		return
	}
	r.Response.WriteJson(g.Map{"message": "channel deleted"})
}

func TestChannel(r *ghttp.Request) {
	id := r.Get("id").Int()
	if id <= 0 {
		r.Response.WriteJson(g.Map{"error": "invalid id"})
		return
	}

	ctx := r.Context()
	ch, err := g.DB().Model("channels").Ctx(ctx).
		Where("id", id).
		Where("deleted_at IS NULL").
		One()
	if err != nil || ch == nil {
		r.Response.WriteJson(g.Map{"error": "channel not found"})
		return
	}

	baseUrl := ch["base_url"].String()
	if baseUrl == "" {
		baseUrl = "https://api.openai.com"
	}
	url := strings.TrimRight(baseUrl, "/") + "/v1/chat/completions"
	apiKey := ch["key"].String()

	// Use the first model from the channel's model list
	models := strings.Split(ch["models"].String(), ",")
	testModel := strings.TrimSpace(models[0])
	if testModel == "" {
		testModel = "gpt-3.5-turbo"
	}

	body := fmt.Sprintf(`{"model":"%s","messages":[{"role":"user","content":"hi"}],"max_tokens":1}`, testModel)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(body))
	if err != nil {
		r.Response.WriteJson(g.Map{"error": "create request failed: " + err.Error()})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		r.Response.WriteJson(g.Map{"error": "request failed: " + err.Error(), "channel_id": id})
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		r.Response.WriteJson(g.Map{
			"error":      fmt.Sprintf("upstream returned status %d", resp.StatusCode),
			"response":   string(respBody),
			"channel_id": id,
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"message":    "channel test passed",
		"status":     resp.StatusCode,
		"channel_id": id,
	})
}

func ListAbilities(r *ghttp.Request) {
	rows, err := g.DB().Model("abilities").Ctx(r.Context()).
		Order("group_name ASC, model ASC").
		All()
	if err != nil {
		r.Response.WriteJson(g.Map{"error": err.Error()})
		return
	}
	r.Response.WriteJson(rows)
}

func CreateAbility(r *ghttp.Request) {
	var data map[string]interface{}
	if err := r.Parse(&data); err != nil {
		r.Response.WriteJson(g.Map{"error": "parse failed: " + err.Error()})
		return
	}

	_, err := g.DB().Model("abilities").Ctx(r.Context()).Insert(data)
	if err != nil {
		r.Response.WriteJson(g.Map{"error": err.Error()})
		return
	}
	r.Response.WriteJson(g.Map{"message": "ability created"})
}

func UpdateAbility(r *ghttp.Request) {
	groupName := r.Get("group_name").String()
	model := r.Get("model").String()
	channelId := r.Get("channel_id").Int()
	if groupName == "" || model == "" || channelId <= 0 {
		r.Response.WriteJson(g.Map{"error": "group_name, model, and channel_id required"})
		return
	}

	var data map[string]interface{}
	if err := r.Parse(&data); err != nil {
		r.Response.WriteJson(g.Map{"error": "parse failed: " + err.Error()})
		return
	}
	delete(data, "group_name")
	delete(data, "model")
	delete(data, "channel_id")

	_, err := g.DB().Model("abilities").Ctx(r.Context()).
		Where("group_name", groupName).
		Where("model", model).
		Where("channel_id", channelId).
		Update(data)
	if err != nil {
		r.Response.WriteJson(g.Map{"error": err.Error()})
		return
	}
	r.Response.WriteJson(g.Map{"message": "ability updated"})
}

func DeleteAbility(r *ghttp.Request) {
	groupName := r.Get("group_name").String()
	model := r.Get("model").String()
	channelId := r.Get("channel_id").Int()
	if groupName == "" || model == "" || channelId <= 0 {
		r.Response.WriteJson(g.Map{"error": "group_name, model, and channel_id required"})
		return
	}

	_, err := g.DB().Model("abilities").Ctx(r.Context()).
		Where("group_name", groupName).
		Where("model", model).
		Where("channel_id", channelId).
		Delete()
	if err != nil {
		r.Response.WriteJson(g.Map{"error": err.Error()})
		return
	}
	r.Response.WriteJson(g.Map{"message": "ability deleted"})
}

func ListUsageRecords(r *ghttp.Request) {
	ctx := r.Context()
	userId := r.Get("user_id").Int64()
	page := r.Get("page").Int()
	pageSize := r.Get("page_size").Int()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	res, err := grpcclient.AssetSvc.ListUsageRecords(ctx, &assetv1.ListUsageRecordsReq{
		UserId:   userId,
		Page:     int32(page),
		PageSize: int32(pageSize),
	})
	if err != nil {
		r.Response.WriteJson(g.Map{"error": err.Error()})
		return
	}

	r.Response.WriteJson(g.Map{
		"records": res.Records,
		"total":   res.Total,
		"page":    page,
		"page_size": pageSize,
	})
}
