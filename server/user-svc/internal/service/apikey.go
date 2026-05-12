package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	userv1 "api/user/v1"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

func (s *UserService) CreateApiKey(ctx context.Context, req *userv1.CreateApiKeyReq) (*userv1.CreateApiKeyRes, error) {
	if req.UserId == 0 {
		return nil, errors.New("user_id is required")
	}
	keyBytes := make([]byte, 24)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, err
	}
	rawKey := "sk-" + hex.EncodeToString(keyBytes)
	name := req.Name
	if name == "" {
		name = "Default Key"
	}
	dbData := g.Map{
		"user_id":              req.UserId,
		"key":                  rawKey,
		"name":                 name,
		"status":               1,
		"model_limits_enabled": req.ModelLimitsEnabled,
		"model_limits":         req.ModelLimits,
		"allow_ips":            req.AllowIps,
		"group_name":           "default",
	}
	if req.ExpireTime > 0 {
		dbData["expire_time"] = gtime.NewFromTimeStamp(req.ExpireTime)
	}
	if req.Group != "" {
		dbData["group_name"] = req.Group
	}
	result, err := g.DB().Model("api_keys").Ctx(ctx).Data(dbData).Insert()
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return &userv1.CreateApiKeyRes{
		ApiKey: &userv1.ApiKey{
			Id: id, UserId: req.UserId, Key: rawKey, Name: name, Status: 1,
			ModelLimitsEnabled: req.ModelLimitsEnabled, ModelLimits: req.ModelLimits,
			ExpireTime: req.ExpireTime, AllowIps: req.AllowIps, Group: req.Group,
		},
		RawKey: rawKey,
	}, nil
}

func (s *UserService) ListApiKeys(ctx context.Context, req *userv1.ListApiKeysReq) (*userv1.ListApiKeysRes, error) {
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 20
	}
	total, err := g.DB().Model("api_keys").Ctx(ctx).
		Where("user_id", req.UserId).Where("deleted_at IS NULL").Count()
	if err != nil {
		return nil, err
	}
	var dbKeys []gdb.Record
	err = g.DB().Model("api_keys").Ctx(ctx).
		Where("user_id", req.UserId).Where("deleted_at IS NULL").
		Order("id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Scan(&dbKeys)
	if err != nil {
		return nil, err
	}
	apiKeys := make([]*userv1.ApiKey, 0)
	for _, k := range dbKeys {
		apiKeys = append(apiKeys, &userv1.ApiKey{
			Id: k["id"].Int64(), UserId: k["user_id"].Int64(), Key: k["key"].String(),
			Name: k["name"].String(), Status: int32(k["status"].Int()),
			ModelLimitsEnabled: k["model_limits_enabled"].Bool(),
			ModelLimits: k["model_limits"].String(), AllowIps: k["allow_ips"].String(),
			Group: k["group_name"].String(), CreatedAt: k["created_at"].String(),
		})
	}
	return &userv1.ListApiKeysRes{ApiKeys: apiKeys, Total: int32(total)}, nil
}

func (s *UserService) DeleteApiKey(ctx context.Context, req *userv1.DeleteApiKeyReq) (*userv1.DeleteApiKeyRes, error) {
	_, err := g.DB().Model("api_keys").Ctx(ctx).
		Where("id", req.Id).Where("user_id", req.UserId).Delete()
	if err != nil {
		return nil, err
	}
	return &userv1.DeleteApiKeyRes{}, nil
}
