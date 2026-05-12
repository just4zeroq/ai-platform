package grpcclient

import (
	assetpb "api-gateway/api/assetpb/v1"
	userpb "api-gateway/api/userpb/v1"
)

var (
	UserSvc  userpb.UserServiceClient
	AssetSvc assetpb.AssetServiceClient
)
