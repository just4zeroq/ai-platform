package cmd

import (
	"context"
	"time"

	assetpb "api/asset/v1"
	userpb "api/user/v1"

	"ai-gateway/internal/grpcclient"
	"ai-gateway/internal/router"
	"ai-gateway/internal/service"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			initGrpcClients(ctx)
			service.LogQ.Start(ctx, 50, 100*time.Millisecond)

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
	// Init user-svc
	cfg := g.Cfg().MustGet(ctx, "grpc.user-svc").Map()
	address := cfg["address"].(string)
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		g.Log().Fatalf(ctx, "failed to connect to user-svc: %v", err)
	}
	grpcclient.UserSvc = userpb.NewUserServiceClient(conn)
	g.Log().Print(ctx, "connected to user-svc gRPC at", address)

	// Init asset-svc
	assetCfg := g.Cfg().MustGet(ctx, "grpc.asset-svc").Map()
	assetAddr := assetCfg["address"].(string)
	assetConn, err := grpc.Dial(assetAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		g.Log().Fatalf(ctx, "failed to connect to asset-svc: %v", err)
	}
	grpcclient.AssetSvc = assetpb.NewAssetServiceClient(assetConn)
	g.Log().Print(ctx, "connected to asset-svc gRPC at", assetAddr)
}
