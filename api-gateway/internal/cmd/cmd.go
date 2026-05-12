package cmd

import (
	"context"

	"api-gateway/internal/controller/asset"
	"api-gateway/internal/controller/user"
	"api-gateway/internal/grpcclient"
	"api-gateway/internal/middleware"

	assetpb "api-gateway/api/assetpb/v1"
	userpb "api-gateway/api/userpb/v1"

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

			s := g.Server()

			userCtrl := &user.Controller{}
			assetCtrl := &asset.Controller{}

			// Public routes
			s.Group("/api/v1", func(group *ghttp.RouterGroup) {
				group.POST("/user/register", userCtrl.Register)
				group.POST("/user/login", userCtrl.Login)
			})

			// Protected routes (JWT required)
			s.Group("/api/v1", func(group *ghttp.RouterGroup) {
				group.Middleware(middleware.Auth)
				group.GET("/user/profile", userCtrl.GetProfile)
				group.GET("/asset/balance", assetCtrl.GetBalance)
				group.GET("/asset/transactions", assetCtrl.ListTransactions)
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
	g.Log().Info(ctx, "connected to user-svc gRPC at", address)

	// Init asset-svc
	assetCfg := g.Cfg().MustGet(ctx, "grpc.asset-svc").Map()
	assetConn, err := grpc.Dial(assetCfg["address"].(string), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		g.Log().Fatalf(ctx, "failed to connect to asset-svc: %v", err)
	}
	grpcclient.AssetSvc = assetpb.NewAssetServiceClient(assetConn)
	g.Log().Info(ctx, "connected to asset-svc gRPC at", assetCfg["address"].(string))
}
