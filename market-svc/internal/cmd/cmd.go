package cmd

import (
	"context"
	"fmt"
	"net"

	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	marketv1 "api/market/v1"
	"market-svc/internal/controller/market"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start gRPC server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			lis, err := net.Listen("tcp", ":8102")
			if err != nil {
				glog.Fatalf(ctx, "failed to listen: %v", err)
			}

			s := grpc.NewServer()

			// Register services
			marketv1.RegisterMarketServiceServer(s, &market.Controller{})

			// Register reflection service on gRPC server
			reflection.Register(s)

			glog.Printf(ctx, "market-svc gRPC server listening at %v", lis.Addr())
			fmt.Printf("market-svc gRPC server listening at %v\n", lis.Addr())

			if err := s.Serve(lis); err != nil {
				glog.Fatalf(ctx, "failed to serve: %v", err)
			}

			return nil
		},
	}
)
