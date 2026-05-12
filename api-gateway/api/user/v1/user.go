package user

import (
	"github.com/gogf/gf/v2/frame/g"
)

type RegisterReq struct {
	g.Meta   `path:"/user/register" method:"POST" summary:"Register" tags:"User"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type RegisterRes struct {
	UserId   int64  `json:"user_id"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

type LoginReq struct {
	g.Meta   `path:"/user/login" method:"POST" summary:"Login" tags:"User"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRes struct {
	UserId   int64  `json:"user_id"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

type GetProfileReq struct {
	g.Meta `path:"/user/profile" method:"GET" summary:"Get Profile" tags:"User"`
}

type GetProfileRes struct {
	Id          int64  `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

type ListKeysReq struct {
	g.Meta `path:"/user/keys" method:"GET" summary:"List API Keys" tags:"User"`
}

type ListKeysRes struct {
	ApiKeys []KeyItem `json:"api_keys"`
	Total   int       `json:"total"`
}

type KeyItem struct {
	Id        int64  `json:"id"`
	Name      string `json:"name"`
	Key       string `json:"key"`
	Status    int32  `json:"status"`
	CreatedAt string `json:"created_at"`
}

type CreateKeyReq struct {
	g.Meta `path:"/user/keys/create" method:"POST" summary:"Create API Key" tags:"User"`
	Name   string `json:"name"`
}

type CreateKeyRes struct {
	ApiKey KeyItem `json:"api_key"`
	RawKey string  `json:"raw_key"`
}

type DeleteKeyReq struct {
	g.Meta `path:"/user/keys/delete" method:"POST" summary:"Delete API Key" tags:"User"`
	Id     int64 `json:"id"`
}

type DeleteKeyRes struct{}
