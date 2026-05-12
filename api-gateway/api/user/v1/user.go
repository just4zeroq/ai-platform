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
