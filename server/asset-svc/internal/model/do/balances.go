// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Balances is the golang structure of table balances for DAO operations like Where/Data.
type Balances struct {
	g.Meta         `orm:"table:balances, do:true"`
	Id             any         //
	UserId         any         //
	Balance        any         //
	TotalRecharged any         //
	TotalConsumed  any         //
	Version        any         //
	CreatedAt      *gtime.Time //
	UpdatedAt      *gtime.Time //
}
