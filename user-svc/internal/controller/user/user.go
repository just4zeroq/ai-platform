package user

import (
	"context"

	userv1 "user-svc/api/user/v1"
)

type Controller struct {
	userv1.UnimplementedUserServiceServer
}

func (c *Controller) Register(ctx context.Context, req *userv1.RegisterReq) (*userv1.RegisterRes, error) {
	return nil, nil
}

func (c *Controller) Login(ctx context.Context, req *userv1.LoginReq) (*userv1.LoginRes, error) {
	return nil, nil
}

func (c *Controller) ValidateToken(ctx context.Context, req *userv1.ValidateTokenReq) (*userv1.ValidateTokenRes, error) {
	return nil, nil
}

func (c *Controller) CreateApiKey(ctx context.Context, req *userv1.CreateApiKeyReq) (*userv1.CreateApiKeyRes, error) {
	return nil, nil
}

func (c *Controller) ListApiKeys(ctx context.Context, req *userv1.ListApiKeysReq) (*userv1.ListApiKeysRes, error) {
	return nil, nil
}

func (c *Controller) DeleteApiKey(ctx context.Context, req *userv1.DeleteApiKeyReq) (*userv1.DeleteApiKeyRes, error) {
	return nil, nil
}

func (c *Controller) GetUser(ctx context.Context, req *userv1.GetUserReq) (*userv1.GetUserRes, error) {
	return nil, nil
}
