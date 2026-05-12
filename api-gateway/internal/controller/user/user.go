package user

import (
	"context"

	userv1 "api-gateway/api/user/v1"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

type Controller struct{}

func (c *Controller) Register(ctx context.Context, req *userv1.RegisterReq) (res *userv1.RegisterRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented, "not implemented")
}

func (c *Controller) Login(ctx context.Context, req *userv1.LoginReq) (res *userv1.LoginRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented, "not implemented")
}

func (c *Controller) GetProfile(ctx context.Context, req *userv1.GetProfileReq) (res *userv1.GetProfileRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented, "not implemented")
}
