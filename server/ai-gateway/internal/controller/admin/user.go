package admin

import (
	assetv1 "api/asset/v1"
	userv1 "api/user/v1"
	"ai-gateway/internal/grpcclient"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func ListUsers(r *ghttp.Request) {
	page := r.Get("page").Int()
	pageSize := r.Get("page_size").Int()
	status := r.Get("status").Int()
	group := r.Get("group").String()
	keyword := r.Get("keyword").String()

	res, err := grpcclient.UserSvc.ListUsers(r.Context(), &userv1.ListUsersReq{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Status:   int32(status),
		Group:    group,
		Keyword:  keyword,
	})
	if err != nil {
		r.Response.WriteJson(g.Map{"error": err.Error()})
		return
	}
	r.Response.WriteJson(g.Map{
		"users": res.Users,
		"total": res.Total,
		"page":  page,
	})
}

func UpdateUser(r *ghttp.Request) {
	userId := r.Get("id").Int64()
	if userId <= 0 {
		r.Response.WriteJson(g.Map{"error": "invalid user id"})
		return
	}

	var req struct {
		Status int32  `json:"status"`
		Group  string `json:"group"`
		Role   int32  `json:"role"`
		Remark string `json:"remark"`
	}
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{"error": "parse failed: " + err.Error()})
		return
	}

	res, err := grpcclient.UserSvc.UpdateUser(r.Context(), &userv1.UpdateUserReq{
		UserId: userId,
		Status: req.Status,
		Group:  req.Group,
		Role:   req.Role,
		Remark: req.Remark,
	})
	if err != nil {
		r.Response.WriteJson(g.Map{"error": err.Error()})
		return
	}
	r.Response.WriteJson(g.Map{"user": res.User})
}

func ListOrders(r *ghttp.Request) {
	userId := r.Get("user_id").Int64()
	page := r.Get("page").Int()
	pageSize := r.Get("page_size").Int()
	status := r.Get("status").String()

	res, err := grpcclient.AssetSvc.ListOrders(r.Context(), &assetv1.ListOrdersReq{
		UserId:   userId,
		Page:     int32(page),
		PageSize: int32(pageSize),
		Status:   status,
	})
	if err != nil {
		r.Response.WriteJson(g.Map{"error": err.Error()})
		return
	}
	r.Response.WriteJson(g.Map{
		"orders":    res.Orders,
		"total":     res.Total,
		"page":      page,
		"page_size": pageSize,
	})
}

func RechargeBalance(r *ghttp.Request) {
	userId := r.Get("id").Int64()
	if userId <= 0 {
		r.Response.WriteJson(g.Map{"error": "invalid user id"})
		return
	}

	var req struct {
		Amount float64 `json:"amount"`
		Remark string  `json:"remark"`
	}
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{"error": "parse failed: " + err.Error()})
		return
	}
	if req.Amount <= 0 {
		r.Response.WriteJson(g.Map{"error": "amount must be positive"})
		return
	}

	res, err := grpcclient.AssetSvc.RechargeBalance(r.Context(), &assetv1.RechargeBalanceReq{
		UserId: userId,
		Amount: req.Amount,
		Remark: req.Remark,
	})
	if err != nil {
		r.Response.WriteJson(g.Map{"error": err.Error()})
		return
	}
	r.Response.WriteJson(g.Map{"balance": res.Balance})
}

func UpdateApiKey(r *ghttp.Request) {
	id := r.Get("id").Int64()
	if id <= 0 {
		r.Response.WriteJson(g.Map{"error": "invalid api key id"})
		return
	}

	var req struct {
		Status             int    `json:"status"`
		Group              string `json:"group"`
		ModelLimits        string `json:"model_limits"`
		ModelLimitsEnabled bool   `json:"model_limits_enabled"`
	}
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{"error": "parse failed: " + err.Error()})
		return
	}

	res, err := grpcclient.UserSvc.UpdateApiKey(r.Context(), &userv1.UpdateApiKeyReq{
		Id:                 id,
		Status:             int32(req.Status),
		Group:              req.Group,
		ModelLimits:        req.ModelLimits,
		ModelLimitsEnabled: req.ModelLimitsEnabled,
	})
	if err != nil {
		r.Response.WriteJson(g.Map{"error": err.Error()})
		return
	}
	r.Response.WriteJson(g.Map{"api_key": res.ApiKey})
}
