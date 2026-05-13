// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Abilities is the golang structure for table abilities.
type Abilities struct {
	GroupName string `json:"groupName" orm:"group_name" description:""` //
	Model     string `json:"model"     orm:"model"      description:""` //
	ChannelId int    `json:"channelId" orm:"channel_id" description:""` //
	Enabled   bool   `json:"enabled"   orm:"enabled"    description:""` //
	Priority  int64  `json:"priority"  orm:"priority"   description:""` //
	Weight    int    `json:"weight"    orm:"weight"     description:""` //
	Tag       string `json:"tag"       orm:"tag"        description:""` //
}
