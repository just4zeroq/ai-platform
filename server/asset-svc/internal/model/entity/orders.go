// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Orders is the golang structure for table orders.
type Orders struct {
	Id            int64       `json:"id"            orm:"id"             description:""` //
	OrderNo       string      `json:"orderNo"       orm:"order_no"       description:""` //
	UserId        int64       `json:"userId"        orm:"user_id"        description:""` //
	Type          string      `json:"type"          orm:"type"           description:""` //
	Status        string      `json:"status"        orm:"status"         description:""` //
	TotalAmount   float64     `json:"totalAmount"   orm:"total_amount"   description:""` //
	PaymentMethod string      `json:"paymentMethod" orm:"payment_method" description:""` //
	CreatedAt     *gtime.Time `json:"createdAt"     orm:"created_at"     description:""` //
	UpdatedAt     *gtime.Time `json:"updatedAt"     orm:"updated_at"     description:""` //
}
