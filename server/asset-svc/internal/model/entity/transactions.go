// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Transactions is the golang structure for table transactions.
type Transactions struct {
	Id            int64       `json:"id"            orm:"id"             description:""` //
	UserId        int64       `json:"userId"        orm:"user_id"        description:""` //
	Type          string      `json:"type"          orm:"type"           description:""` //
	Amount        float64     `json:"amount"        orm:"amount"         description:""` //
	BalanceBefore float64     `json:"balanceBefore" orm:"balance_before" description:""` //
	BalanceAfter  float64     `json:"balanceAfter"  orm:"balance_after"  description:""` //
	ReferenceType string      `json:"referenceType" orm:"reference_type" description:""` //
	ReferenceId   int64       `json:"referenceId"   orm:"reference_id"   description:""` //
	Remark        string      `json:"remark"        orm:"remark"         description:""` //
	CreatedAt     *gtime.Time `json:"createdAt"     orm:"created_at"     description:""` //
}
