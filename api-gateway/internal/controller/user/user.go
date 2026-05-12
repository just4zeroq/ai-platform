package user

import (
	"context"

	userv1 "api-gateway/api/user/v1"
	userpb "api-gateway/api/userpb/v1"
	"api-gateway/internal/grpcclient"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

type Controller struct{}

func (c *Controller) Register(ctx context.Context, req *userv1.RegisterReq) (res *userv1.RegisterRes, err error) {
	pbRes, err := grpcclient.UserSvc.Register(ctx, &userpb.RegisterReq{
		Username: req.Username, Password: req.Password, Email: req.Email,
	})
	if err != nil {
		return nil, err
	}
	return &userv1.RegisterRes{
		UserId: pbRes.User.Id, Username: pbRes.User.Username, Token: pbRes.AccessToken,
	}, nil
}

func (c *Controller) Login(ctx context.Context, req *userv1.LoginReq) (res *userv1.LoginRes, err error) {
	pbRes, err := grpcclient.UserSvc.Login(ctx, &userpb.LoginReq{
		Username: req.Username, Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &userv1.LoginRes{
		UserId: pbRes.User.Id, Username: pbRes.User.Username, Token: pbRes.AccessToken,
	}, nil
}

func (c *Controller) GetProfile(ctx context.Context, req *userv1.GetProfileReq) (res *userv1.GetProfileRes, err error) {
	r := g.RequestFromCtx(ctx)
	userId := r.GetCtxVar("user_id").Int64()
	if userId == 0 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized)
	}
	pbRes, err := grpcclient.UserSvc.GetUser(ctx, &userpb.GetUserReq{UserId: userId})
	if err != nil {
		return nil, err
	}
	return &userv1.GetProfileRes{
		Id: pbRes.User.Id, Username: pbRes.User.Username,
		Email: pbRes.User.Email, DisplayName: pbRes.User.DisplayName,
	}, nil
}
