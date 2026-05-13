// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Channels is the golang structure of table channels for DAO operations like Where/Data.
type Channels struct {
	g.Meta       `orm:"table:channels, do:true"`
	Id           any         //
	Type         any         //
	Key          any         //
	Name         any         //
	Models       any         //
	GroupName    any         //
	Status       any         //
	Priority     any         //
	Weight       any         //
	ModelMapping any         //
	BaseUrl      any         //
	CreatedAt    any         //
	UpdatedAt    any         //
	DeletedAt    *gtime.Time //
}
