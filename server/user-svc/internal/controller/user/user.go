package user

import (
	"context"

	userv1 "api/user/v1"
	"user-svc/internal/service"
)

type Controller struct {
	userv1.UnimplementedUserServiceServer
	svc *service.UserService
}

func New(svc *service.UserService) *Controller {
	return &Controller{svc: svc}
}

func (c *Controller) Register(ctx context.Context, req *userv1.RegisterReq) (*userv1.RegisterRes, error) {
	return c.svc.Register(ctx, req)
}
func (c *Controller) Login(ctx context.Context, req *userv1.LoginReq) (*userv1.LoginRes, error) {
	return c.svc.Login(ctx, req)
}
func (c *Controller) ValidateToken(ctx context.Context, req *userv1.ValidateTokenReq) (*userv1.ValidateTokenRes, error) {
	return c.svc.ValidateToken(ctx, req.Token)
}
func (c *Controller) CreateApiKey(ctx context.Context, req *userv1.CreateApiKeyReq) (*userv1.CreateApiKeyRes, error) {
	return c.svc.CreateApiKey(ctx, req)
}
func (c *Controller) ListApiKeys(ctx context.Context, req *userv1.ListApiKeysReq) (*userv1.ListApiKeysRes, error) {
	return c.svc.ListApiKeys(ctx, req)
}
func (c *Controller) DeleteApiKey(ctx context.Context, req *userv1.DeleteApiKeyReq) (*userv1.DeleteApiKeyRes, error) {
	return c.svc.DeleteApiKey(ctx, req)
}
func (c *Controller) GetUser(ctx context.Context, req *userv1.GetUserReq) (*userv1.GetUserRes, error) {
	user, err := c.svc.GetUser(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	return &userv1.GetUserRes{User: user}, nil
}

func (c *Controller) ListUsers(ctx context.Context, req *userv1.ListUsersReq) (*userv1.ListUsersRes, error) {
	return c.svc.ListUsers(ctx, req)
}

func (c *Controller) UpdateUser(ctx context.Context, req *userv1.UpdateUserReq) (*userv1.UpdateUserRes, error) {
	return c.svc.UpdateUser(ctx, req)
}

func (c *Controller) UpdateApiKey(ctx context.Context, req *userv1.UpdateApiKeyReq) (*userv1.UpdateApiKeyRes, error) {
	return c.svc.UpdateApiKey(ctx, req)
}
