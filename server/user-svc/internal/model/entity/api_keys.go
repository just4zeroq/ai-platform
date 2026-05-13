// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ApiKeys is the golang structure for table api_keys.
type ApiKeys struct {
	Id                 int64       `json:"id"                 orm:"id"                   description:""` //
	UserId             int64       `json:"userId"             orm:"user_id"              description:""` //
	Key                string      `json:"key"                orm:"key"                  description:""` //
	Status             int         `json:"status"             orm:"status"               description:""` //
	ExpireTime         *gtime.Time `json:"expireTime"         orm:"expire_time"          description:""` //
	ModelLimits        string      `json:"modelLimits"        orm:"model_limits"         description:""` //
	ModelLimitsEnabled bool        `json:"modelLimitsEnabled" orm:"model_limits_enabled" description:""` //
	GroupName          string      `json:"groupName"          orm:"group_name"           description:""` //
	CreatedAt          *gtime.Time `json:"createdAt"          orm:"created_at"           description:""` //
	UpdatedAt          *gtime.Time `json:"updatedAt"          orm:"updated_at"           description:""` //
	DeletedAt          *gtime.Time `json:"deletedAt"          orm:"deleted_at"           description:""` //
}
