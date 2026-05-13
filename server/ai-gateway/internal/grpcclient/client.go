package grpcclient

import (
	assetpb "api/asset/v1"
	userpb "api/user/v1"
)

var (
	UserSvc  userpb.UserServiceClient
	AssetSvc assetpb.AssetServiceClient
)
