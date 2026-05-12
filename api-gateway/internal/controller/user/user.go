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

func (c *Controller) ListKeys(ctx context.Context, req *userv1.ListKeysReq) (res *userv1.ListKeysRes, err error) {
	r := g.RequestFromCtx(ctx)
	userId := r.GetCtxVar("user_id").Int64()
	if userId == 0 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized)
	}
	pbRes, err := grpcclient.UserSvc.ListApiKeys(ctx, &userpb.ListApiKeysReq{
		UserId: userId, Page: 1, PageSize: 100,
	})
	if err != nil {
		return nil, err
	}
	items := make([]userv1.KeyItem, 0)
	for _, k := range pbRes.ApiKeys {
		items = append(items, userv1.KeyItem{
			Id: k.Id, Name: k.Name, Key: k.Key,
			Status: k.Status, CreatedAt: k.CreatedAt,
		})
	}
	return &userv1.ListKeysRes{ApiKeys: items, Total: int(pbRes.Total)}, nil
}

func (c *Controller) CreateKey(ctx context.Context, req *userv1.CreateKeyReq) (res *userv1.CreateKeyRes, err error) {
	r := g.RequestFromCtx(ctx)
	userId := r.GetCtxVar("user_id").Int64()
	if userId == 0 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized)
	}
	pbRes, err := grpcclient.UserSvc.CreateApiKey(ctx, &userpb.CreateApiKeyReq{
		UserId: userId, Name: req.Name,
	})
	if err != nil {
		return nil, err
	}
	return &userv1.CreateKeyRes{
		ApiKey: userv1.KeyItem{
			Id: pbRes.ApiKey.Id, Name: pbRes.ApiKey.Name,
			Key: pbRes.ApiKey.Key, Status: pbRes.ApiKey.Status,
			CreatedAt: pbRes.ApiKey.CreatedAt,
		},
		RawKey: pbRes.RawKey,
	}, nil
}

func (c *Controller) DeleteKey(ctx context.Context, req *userv1.DeleteKeyReq) (res *userv1.DeleteKeyRes, err error) {
	r := g.RequestFromCtx(ctx)
	userId := r.GetCtxVar("user_id").Int64()
	if userId == 0 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized)
	}
	_, err = grpcclient.UserSvc.DeleteApiKey(ctx, &userpb.DeleteApiKeyReq{
		Id: req.Id, UserId: userId,
	})
	if err != nil {
		return nil, err
	}
	return &userv1.DeleteKeyRes{}, nil
}
