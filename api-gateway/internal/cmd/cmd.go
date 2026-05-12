package cmd

import (
	"context"

	"api-gateway/internal/controller/asset"
	"api-gateway/internal/controller/user"

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
			s.Group("/api/v1", func(group *ghttp.RouterGroup) {
				group.Bind(
					&user.Controller{},
					&asset.Controller{},
				)
			})
			s.Run()
			return nil
		},
	}
)
