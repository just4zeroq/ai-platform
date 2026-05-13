package model

// Ability maps a model+group pair to a channel for routing.
type Ability struct {
	Group     string `json:"group"`
	Model     string `json:"model"`
	ChannelId int    `json:"channel_id"`
	Enabled   bool   `json:"enabled"`
	Priority  int64  `json:"priority"`
	Weight    uint   `json:"weight"`
	Tag       string `json:"tag"`
}
