package asset

import (
	assetv1 "api-gateway/api/asset/v1"
	assetpb "api/asset/v1"
	"api-gateway/internal/grpcclient"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type Controller struct{}

func (c *Controller) GetBalance(r *ghttp.Request) {
	userId := r.GetCtxVar("user_id").Int64()
	if userId == 0 {
		r.Response.WriteJson(g.Map{"code": gcode.CodeNotAuthorized.Code(), "message": "unauthorized"})
		return
	}
	pbRes, err := grpcclient.AssetSvc.GetBalance(r.Context(), &assetpb.GetBalanceReq{UserId: userId})
	if err != nil {
		r.Response.WriteJson(g.Map{"code": gcode.CodeInvalidRequest.Code(), "message": err.Error()})
		return
	}
	r.Response.WriteJson(assetv1.GetBalanceRes{Balance: pbRes.Balance.GetBalance()})
}

func (c *Controller) ListTransactions(r *ghttp.Request) {
	userId := r.GetCtxVar("user_id").Int64()
	if userId == 0 {
		r.Response.WriteJson(g.Map{"code": gcode.CodeNotAuthorized.Code(), "message": "unauthorized"})
		return
	}
	page := int32(r.Get("page", 1).Int())
	pageSize := int32(r.Get("pageSize", 20).Int())
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	pbRes, err := grpcclient.AssetSvc.ListTransactions(r.Context(), &assetpb.ListTransactionsReq{
		UserId: userId, Page: page, PageSize: pageSize,
	})
	if err != nil {
		r.Response.WriteJson(g.Map{"code": gcode.CodeInvalidRequest.Code(), "message": err.Error()})
		return
	}
	items := make([]assetv1.TransactionItem, 0)
	for _, tx := range pbRes.Transactions {
		items = append(items, assetv1.TransactionItem{
			Id: tx.Id, Type: tx.Type, Amount: tx.Amount,
			BalanceAfter: tx.BalanceAfter, Remark: tx.Remark,
		})
	}
	r.Response.WriteJson(assetv1.ListTransactionsRes{List: items, Total: int(pbRes.Total)})
}
