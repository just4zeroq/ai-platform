// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Users is the golang structure for table users.
type Users struct {
	Id          int64       `json:"id"          orm:"id"           description:""` //
	Username    string      `json:"username"    orm:"username"     description:""` //
	Password    string      `json:"password"    orm:"password"     description:""` //
	Email       string      `json:"email"       orm:"email"        description:""` //
	Phone       string      `json:"phone"       orm:"phone"        description:""` //
	DisplayName string      `json:"displayName" orm:"display_name" description:""` //
	Avatar      string      `json:"avatar"      orm:"avatar"       description:""` //
	Source      string      `json:"source"      orm:"source"       description:""` //
	Role        int         `json:"role"        orm:"role"         description:""` //
	Status      int         `json:"status"      orm:"status"       description:""` //
	GroupName   string      `json:"groupName"   orm:"group_name"   description:""` //
	Remark      string      `json:"remark"      orm:"remark"       description:""` //
	TenantId    int64       `json:"tenantId"    orm:"tenant_id"    description:""` //
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"   description:""` //
	UpdatedAt   *gtime.Time `json:"updatedAt"   orm:"updated_at"   description:""` //
}
