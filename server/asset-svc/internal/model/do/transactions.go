// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Transactions is the golang structure of table transactions for DAO operations like Where/Data.
type Transactions struct {
	g.Meta        `orm:"table:transactions, do:true"`
	Id            any         //
	UserId        any         //
	Type          any         //
	Amount        any         //
	BalanceBefore any         //
	BalanceAfter  any         //
	ReferenceType any         //
	ReferenceId   any         //
	Remark        any         //
	CreatedAt     *gtime.Time //
}
