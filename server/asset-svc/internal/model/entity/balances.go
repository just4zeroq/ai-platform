// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Balances is the golang structure for table balances.
type Balances struct {
	Id             int64       `json:"id"             orm:"id"              description:""` //
	UserId         int64       `json:"userId"         orm:"user_id"         description:""` //
	Balance        float64     `json:"balance"        orm:"balance"         description:""` //
	TotalRecharged float64     `json:"totalRecharged" orm:"total_recharged" description:""` //
	TotalConsumed  float64     `json:"totalConsumed"  orm:"total_consumed"  description:""` //
	Version        int         `json:"version"        orm:"version"         description:""` //
	CreatedAt      *gtime.Time `json:"createdAt"      orm:"created_at"      description:""` //
	UpdatedAt      *gtime.Time `json:"updatedAt"      orm:"updated_at"      description:""` //
}
