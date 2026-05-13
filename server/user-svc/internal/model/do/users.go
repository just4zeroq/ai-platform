// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Users is the golang structure of table users for DAO operations like Where/Data.
type Users struct {
	g.Meta      `orm:"table:users, do:true"`
	Id          any         //
	Username    any         //
	Password    any         //
	Email       any         //
	Phone       any         //
	DisplayName any         //
	Avatar      any         //
	Source      any         //
	Role        any         //
	Status      any         //
	GroupName   any         //
	Remark      any         //
	TenantId    any         //
	CreatedAt   *gtime.Time //
	UpdatedAt   *gtime.Time //
}
