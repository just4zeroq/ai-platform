package middleware

import (
	"net/http"
	"strings"

	"ai-gateway/internal/grpcclient"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	userpb "api/user/v1"
)

// TokenAuth validates the API key via gRPC call to user-svc.
// It extracts "sk-xxx" from Authorization header, calls user-svc.ValidateToken,
// and injects user context into the request.
func TokenAuth(r *ghttp.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		r.Response.WriteStatus(http.StatusUnauthorized, g.Map{"error": "missing authorization header"})
		r.Exit()
		return
	}
	key := authHeader
	if strings.HasPrefix(key, "Bearer ") || strings.HasPrefix(key, "bearer ") {
		key = key[7:]
	}

	// Call user-svc via gRPC to validate the API key
	ctx := r.Context()
	res, err := grpcclient.UserSvc.ValidateToken(ctx, &userpb.ValidateTokenReq{Token: key})
	if err != nil {
		g.Log().Errorf(ctx, "token validation failed: %v", err)
		r.Response.WriteStatus(http.StatusUnauthorized, g.Map{"error": "invalid token"})
		r.Exit()
		return
	}
	if !res.HasToken {
		r.Response.WriteStatus(http.StatusUnauthorized, g.Map{"error": "API key required"})
		r.Exit()
		return
	}

	// Inject validated user info into context for downstream handlers
	r.SetCtxVar("user_id", res.UserId)
	r.SetCtxVar("api_key_id", res.ApiKeyId)
	r.SetCtxVar("user_role", res.UserRole)
	r.SetCtxVar("user_group", res.KeyGroup)
	r.SetCtxVar("model_limits_enabled", res.ModelLimitsEnabled)
	r.SetCtxVar("model_limits", res.ModelLimits)
	r.SetCtxVar("user_status", res.UserStatus)

	r.Middleware.Next()
}
