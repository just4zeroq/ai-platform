package cmd

import (
	"context"
	"fmt"
	"net"

	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	userv1 "user-svc/api/user/v1"
	"user-svc/internal/controller/user"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start gRPC server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			lis, err := net.Listen("tcp", ":8100")
			if err != nil {
				glog.Fatalf(ctx, "failed to listen: %v", err)
			}

			s := grpc.NewServer()

			// Register services
			userv1.RegisterUserServiceServer(s, &user.Controller{})

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
