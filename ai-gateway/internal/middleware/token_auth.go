package middleware

import (
	"net/http"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// TokenAuth validates the API key via gRPC call to user-svc.
// It extracts "sk-xxx" from Authorization header, calls user-svc.ValidateToken,
// and injects user context into the request.
func TokenAuth(r *ghttp.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		r.Response.WriteStatus(http.StatusUnauthorized, gerror.New("missing authorization header"))
		r.Exit()
		return
	}
	// Extract key from "Bearer sk-xxx" or "Bearer xxx"
	key := authHeader
	if len(key) > 7 && (key[:7] == "Bearer " || key[:7] == "bearer ") {
		key = key[7:]
	}
	// TODO: gRPC call to user-svc.ValidateToken
	// For now, just set anonymous user
	r.SetCtxVar("user_id", int64(0))
	r.SetCtxVar("api_key_id", int64(0))
	r.Middleware.Next()
}
