package cmd

import (
	"context"

	"ai-gateway/internal/grpcclient"
	"ai-gateway/internal/router"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	userpb "ai-gateway/api/userpb/v1"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			initGrpcClients(ctx)

			s := g.Server()
			s.Group("/", func(group *ghttp.RouterGroup) {
				router.RelayRouter(group)
			})
			s.Run()
			return nil
		},
	}
)

func initGrpcClients(ctx context.Context) {
	cfg := g.Cfg().MustGet(ctx, "grpc.user-svc").Map()
	address := cfg["address"].(string)
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		g.Log().Fatalf(ctx, "failed to connect to user-svc: %v", err)
	}
	grpcclient.UserSvc = userpb.NewUserServiceClient(conn)
	g.Log().Print(ctx, "connected to user-svc gRPC at", address)
}
