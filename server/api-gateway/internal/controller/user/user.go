package user

import (
	userv1 "api-gateway/api/user/v1"
	userpb "api/user/v1"
	"api-gateway/internal/grpcclient"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type Controller struct{}

func (c *Controller) Register(r *ghttp.Request) {
	var req userv1.RegisterReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{"code": gcode.CodeInvalidParameter.Code(), "message": err.Error()})
		return
	}
	g.Log().Info(r.Context(), "Register called", req.Username)
	pbRes, err := grpcclient.UserSvc.Register(r.Context(), &userpb.RegisterReq{
		Username: req.Username, Password: req.Password, Email: req.Email,
	})
	if err != nil {
		g.Log().Error(r.Context(), "Register gRPC error", err)
		r.Response.WriteJson(g.Map{"code": gcode.CodeInvalidRequest.Code(), "message": err.Error()})
		return
	}
	g.Log().Info(r.Context(), "Register gRPC success", pbRes.User.Id, pbRes.AccessToken)
	r.Response.WriteJson(g.Map{
		"code":    0,
		"message": "ok",
		"data": userv1.RegisterRes{
			UserId: pbRes.User.Id, Username: pbRes.User.Username, Token: pbRes.AccessToken,
		},
	})
	r.Exit()
}

func (c *Controller) Login(r *ghttp.Request) {
	var req userv1.LoginReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{"code": gcode.CodeInvalidParameter.Code(), "message": err.Error()})
		return
	}
	pbRes, err := grpcclient.UserSvc.Login(r.Context(), &userpb.LoginReq{
		Username: req.Username, Password: req.Password,
	})
	if err != nil {
		r.Response.WriteJson(g.Map{"code": gcode.CodeInvalidRequest.Code(), "message": err.Error()})
		return
	}
	r.Response.WriteJson(g.Map{
		"code":    0,
		"message": "ok",
		"data": userv1.LoginRes{
			UserId: pbRes.User.Id, Username: pbRes.User.Username, Token: pbRes.AccessToken,
		},
	})
	r.Exit()
}

func (c *Controller) GetProfile(r *ghttp.Request) {
	userId := r.GetCtxVar("user_id").Int64()
	if userId == 0 {
		r.Response.WriteJson(g.Map{"code": gcode.CodeNotAuthorized.Code(), "message": "unauthorized"})
		return
	}
	pbRes, err := grpcclient.UserSvc.GetUser(r.Context(), &userpb.GetUserReq{UserId: userId})
	if err != nil {
		r.Response.WriteJson(g.Map{"code": gcode.CodeInvalidRequest.Code(), "message": err.Error()})
		return
	}
	r.Response.WriteJson(userv1.GetProfileRes{
		Id: pbRes.User.Id, Username: pbRes.User.Username,
		Email: pbRes.User.Email, DisplayName: pbRes.User.DisplayName,
	})
}

func (c *Controller) ListKeys(r *ghttp.Request) {
	userId := r.GetCtxVar("user_id").Int64()
	if userId == 0 {
		r.Response.WriteJson(g.Map{"code": gcode.CodeNotAuthorized.Code(), "message": "unauthorized"})
		return
	}
	pbRes, err := grpcclient.UserSvc.ListApiKeys(r.Context(), &userpb.ListApiKeysReq{
		UserId: userId, Page: 1, PageSize: 100,
	})
	if err != nil {
		r.Response.WriteJson(g.Map{"code": gcode.CodeInvalidRequest.Code(), "message": err.Error()})
		return
	}
	items := make([]userv1.KeyItem, 0)
	for _, k := range pbRes.ApiKeys {
		items = append(items, userv1.KeyItem{
			Id: k.Id, Name: k.Name, Key: k.Key,
			Status: k.Status, CreatedAt: k.CreatedAt,
		})
	}
	r.Response.WriteJson(userv1.ListKeysRes{ApiKeys: items, Total: int(pbRes.Total)})
}

func (c *Controller) CreateKey(r *ghttp.Request) {
	var req userv1.CreateKeyReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{"code": gcode.CodeInvalidParameter.Code(), "message": err.Error()})
		return
	}
	userId := r.GetCtxVar("user_id").Int64()
	if userId == 0 {
		r.Response.WriteJson(g.Map{"code": gcode.CodeNotAuthorized.Code(), "message": "unauthorized"})
		return
	}
	pbRes, err := grpcclient.UserSvc.CreateApiKey(r.Context(), &userpb.CreateApiKeyReq{
		UserId: userId, Name: req.Name,
	})
	if err != nil {
		r.Response.WriteJson(g.Map{"code": gcode.CodeInvalidRequest.Code(), "message": err.Error()})
		return
	}
	r.Response.WriteJson(userv1.CreateKeyRes{
		ApiKey: userv1.KeyItem{
			Id: pbRes.ApiKey.Id, Name: pbRes.ApiKey.Name,
			Key: pbRes.ApiKey.Key, Status: pbRes.ApiKey.Status,
			CreatedAt: pbRes.ApiKey.CreatedAt,
		},
		RawKey: pbRes.RawKey,
	})
}

func (c *Controller) DeleteKey(r *ghttp.Request) {
	var req userv1.DeleteKeyReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{"code": gcode.CodeInvalidParameter.Code(), "message": err.Error()})
		return
	}
	userId := r.GetCtxVar("user_id").Int64()
	if userId == 0 {
		r.Response.WriteJson(g.Map{"code": gcode.CodeNotAuthorized.Code(), "message": "unauthorized"})
		return
	}
	_, err := grpcclient.UserSvc.DeleteApiKey(r.Context(), &userpb.DeleteApiKeyReq{
		Id: req.Id, UserId: userId,
	})
	if err != nil {
		r.Response.WriteJson(g.Map{"code": gcode.CodeInvalidRequest.Code(), "message": err.Error()})
		return
	}
	r.Response.WriteJson(userv1.DeleteKeyRes{})
}
