// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// ApiKeys is the golang structure of table api_keys for DAO operations like Where/Data.
type ApiKeys struct {
	g.Meta             `orm:"table:api_keys, do:true"`
	Id                 any         //
	UserId             any         //
	Key                any         //
	Status             any         //
	ExpireTime         *gtime.Time //
	ModelLimits        any         //
	ModelLimitsEnabled any         //
	GroupName          any         //
	CreatedAt          *gtime.Time //
	UpdatedAt          *gtime.Time //
	DeletedAt          *gtime.Time //
}
