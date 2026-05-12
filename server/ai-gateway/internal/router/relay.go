package router

import (
	"ai-gateway/internal/middleware"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func RelayRouter(group *ghttp.RouterGroup) {
	group.Middleware(middleware.CORS)
	group.Middleware(middleware.TokenAuth)
	group.Middleware(middleware.ModelRateLimit)

	// OpenAI-compatible endpoints
	group.POST("/v1/chat/completions", func(r *ghttp.Request) {
		// TODO: relay to upstream provider
		r.Response.WriteJson(g.Map{
			"error": "not implemented",
		})
	})

	group.POST("/v1/completions", func(r *ghttp.Request) {
		r.Response.WriteJson(g.Map{
			"error": "not implemented",
		})
	})

	group.POST("/v1/embeddings", func(r *ghttp.Request) {
		r.Response.WriteJson(g.Map{
			"error": "not implemented",
		})
	})

	group.GET("/v1/models", func(r *ghttp.Request) {
		r.Response.WriteJson(g.Map{
			"data": []interface{}{},
		})
	})
}
