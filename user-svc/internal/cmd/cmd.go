package cmd

import (
	"context"
	"fmt"
	"net"

	assetpb "user-svc/api/assetpb/v1"
	userv1 "user-svc/api/user/v1"
	"user-svc/internal/controller/user"
	"user-svc/internal/grpcclient"
	"user-svc/internal/service"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start gRPC server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			initDB(ctx)
			initGrpcClients(ctx)
			lis, err := net.Listen("tcp", ":8100")
			if err != nil {
				glog.Fatalf(ctx, "failed to listen: %v", err)
			}

			s := grpc.NewServer()

			// Register services
			userv1.RegisterUserServiceServer(s, user.New(&service.UserService{}))

			// Register reflection service on gRPC server
			reflection.Register(s)

			glog.Printf(ctx, "user-svc gRPC server listening at %v", lis.Addr())
			fmt.Printf("user-svc gRPC server listening at %v\n", lis.Addr())

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

func initGrpcClients(ctx context.Context) {
	cfg := g.Cfg().MustGet(ctx, "grpc.asset-svc").Map()
	address := cfg["address"].(string)
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		glog.Fatalf(ctx, "failed to connect to asset-svc: %v", err)
	}
	grpcclient.AssetSvc = assetpb.NewAssetServiceClient(conn)
	glog.Printf(ctx, "connected to asset-svc gRPC at %s", address)
}
