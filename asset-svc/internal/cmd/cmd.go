package cmd

import (
	"context"
	"fmt"
	"net"

	assetv1 "asset-svc/api/asset/v1"
	"asset-svc/internal/controller/asset"
	"asset-svc/internal/service"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start gRPC server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			initDB(ctx)
			lis, err := net.Listen("tcp", ":8101")
			if err != nil {
				glog.Fatalf(ctx, "failed to listen: %v", err)
			}
			s := grpc.NewServer()
			assetv1.RegisterAssetServiceServer(s, asset.New(&service.AssetService{}))
			reflection.Register(s)
			glog.Printf(ctx, "asset-svc gRPC server listening at %v", lis.Addr())
			fmt.Printf("asset-svc gRPC server listening at %v\n", lis.Addr())
			if err := s.Serve(lis); err != nil {
				glog.Fatalf(ctx, "failed to serve: %v", err)
			}
			return nil
		},
	}
)

func initDB(ctx context.Context) {
	if err := g.DB().PingMaster(); err != nil {
		glog.Fatalf(ctx, "database connection failed: %v", err)
	}
	glog.Printf(ctx, "database connected successfully")
}
