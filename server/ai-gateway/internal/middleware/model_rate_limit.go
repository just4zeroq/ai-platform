package middleware

import "github.com/gogf/gf/v2/net/ghttp"

// ModelRateLimit limits requests per model.
// Skeleton for now — will be implemented with Redis counter.
func ModelRateLimit(r *ghttp.Request) {
	// TODO: implement per-model rate limiting
	r.Middleware.Next()
}
