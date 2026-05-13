// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// UsageRecords is the golang structure of table usage_records for DAO operations like Where/Data.
type UsageRecords struct {
	g.Meta           `orm:"table:usage_records, do:true"`
	Id               any         //
	UserId           any         //
	ApiKeyId         any         //
	ModelName        any         //
	PromptTokens     any         //
	CompletionTokens any         //
	Quota            any         //
	RequestId        any         //
	ChannelId        any         //
	ChannelName      any         //
	CreatedAt        *gtime.Time //
}
