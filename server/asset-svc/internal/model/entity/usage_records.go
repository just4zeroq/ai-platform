// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// UsageRecords is the golang structure for table usage_records.
type UsageRecords struct {
	Id               int64       `json:"id"               orm:"id"                description:""` //
	UserId           int64       `json:"userId"           orm:"user_id"           description:""` //
	ApiKeyId         int64       `json:"apiKeyId"         orm:"api_key_id"        description:""` //
	ModelName        string      `json:"modelName"        orm:"model_name"        description:""` //
	PromptTokens     int         `json:"promptTokens"     orm:"prompt_tokens"     description:""` //
	CompletionTokens int         `json:"completionTokens" orm:"completion_tokens" description:""` //
	Quota            float64     `json:"quota"            orm:"quota"             description:""` //
	RequestId        string      `json:"requestId"        orm:"request_id"        description:""` //
	ChannelId        int64       `json:"channelId"        orm:"channel_id"        description:""` //
	ChannelName      string      `json:"channelName"      orm:"channel_name"      description:""` //
	CreatedAt        *gtime.Time `json:"createdAt"        orm:"created_at"        description:""` //
}
