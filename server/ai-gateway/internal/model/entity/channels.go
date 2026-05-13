// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Channels is the golang structure for table channels.
type Channels struct {
	Id           int         `json:"id"           orm:"id"            description:""` //
	Type         int         `json:"type"         orm:"type"          description:""` //
	Key          string      `json:"key"          orm:"key"           description:""` //
	Name         string      `json:"name"         orm:"name"          description:""` //
	Models       string      `json:"models"       orm:"models"        description:""` //
	GroupName    string      `json:"groupName"    orm:"group_name"    description:""` //
	Status       int         `json:"status"       orm:"status"        description:""` //
	Priority     int64       `json:"priority"     orm:"priority"      description:""` //
	Weight       int         `json:"weight"       orm:"weight"        description:""` //
	ModelMapping string      `json:"modelMapping" orm:"model_mapping" description:""` //
	BaseUrl      string      `json:"baseUrl"      orm:"base_url"      description:""` //
	CreatedAt    int64       `json:"createdAt"    orm:"created_at"    description:""` //
	UpdatedAt    int64       `json:"updatedAt"    orm:"updated_at"    description:""` //
	DeletedAt    *gtime.Time `json:"deletedAt"    orm:"deleted_at"    description:""` //
}
