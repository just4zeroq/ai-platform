package middleware

import (
	"net/http"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func AdminAuth(r *ghttp.Request) {
	role := r.GetCtxVar("user_role").Int()
	if role < 1 {
		r.Response.WriteStatus(http.StatusForbidden, g.Map{"error": "admin access required"})
		r.Exit()
		return
	}
	r.Middleware.Next()
}
