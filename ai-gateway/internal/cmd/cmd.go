package cmd

import (
	"context"

	"ai-gateway/internal/router"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()

			// Relay routes (no /api prefix — matches OpenAI API format)
			s.Group("/", func(group *ghttp.RouterGroup) {
				router.RelayRouter(group)
			})

			s.Run()
			return nil
		},
	}
)
