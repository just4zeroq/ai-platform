package router

import (
	"ai-gateway/internal/controller/admin"
	"ai-gateway/internal/middleware"
	"ai-gateway/internal/service"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func RelayRouter(group *ghttp.RouterGroup) {
	group.Middleware(middleware.CORS)
	group.Middleware(middleware.TokenAuth)
	group.Middleware(middleware.ModelRateLimit)

	// OpenAI-compatible endpoints
	group.POST("/v1/chat/completions", service.ChatCompletions)
	group.POST("/v1/completions", service.ChatCompletions)

	group.GET("/v1/models", func(r *ghttp.Request) {
		models, err := service.ListModels(r.Context())
		if err != nil {
			r.Response.WriteJson(g.Map{"data": []interface{}{}})
			return
		}
		r.Response.WriteJson(g.Map{
			"object": "list",
			"data":   models,
		})
	})

	// Admin routes (require admin role)
	group.Group("/api/v1/admin", func(adminGroup *ghttp.RouterGroup) {
		adminGroup.Middleware(middleware.AdminAuth)

		// Channel management
		adminGroup.GET("/channels", admin.ListChannels)
		adminGroup.POST("/channels", admin.CreateChannel)
		adminGroup.PUT("/channels/:id", admin.UpdateChannel)
		adminGroup.DELETE("/channels/:id", admin.DeleteChannel)
		adminGroup.POST("/channels/:id/test", admin.TestChannel)

		// Ability management (composite PK: group_name + model + channel_id)
		adminGroup.GET("/abilities", admin.ListAbilities)
		adminGroup.POST("/abilities", admin.CreateAbility)
		adminGroup.PUT("/abilities", admin.UpdateAbility)
		adminGroup.DELETE("/abilities", admin.DeleteAbility)

		// Usage logs
		adminGroup.GET("/usage-records", admin.ListUsageRecords)

		// User management (via user-svc gRPC)
		adminGroup.GET("/users", admin.ListUsers)
		adminGroup.PUT("/users/:id", admin.UpdateUser)
		adminGroup.PUT("/api-keys/:id", admin.UpdateApiKey)

		// Order management (via asset-svc gRPC)
		adminGroup.GET("/orders", admin.ListOrders)
		adminGroup.POST("/users/:id/recharge", admin.RechargeBalance)
	})
}
