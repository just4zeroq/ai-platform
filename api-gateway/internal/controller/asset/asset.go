package asset

import (
	"context"

	assetv1 "api-gateway/api/asset/v1"
	assetpb "api-gateway/api/assetpb/v1"
	"api-gateway/internal/grpcclient"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

type Controller struct{}

func (c *Controller) GetBalance(ctx context.Context, req *assetv1.GetBalanceReq) (res *assetv1.GetBalanceRes, err error) {
	r := g.RequestFromCtx(ctx)
	userId := r.GetCtxVar("user_id").Int64()
	if userId == 0 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized)
	}
	pbRes, err := grpcclient.AssetSvc.GetBalance(ctx, &assetpb.GetBalanceReq{UserId: userId})
	if err != nil {
		return nil, err
	}
	return &assetv1.GetBalanceRes{Balance: pbRes.Balance.GetBalance()}, nil
}

func (c *Controller) ListTransactions(ctx context.Context, req *assetv1.ListTransactionsReq) (res *assetv1.ListTransactionsRes, err error) {
	r := g.RequestFromCtx(ctx)
	userId := r.GetCtxVar("user_id").Int64()
	if userId == 0 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized)
	}
	page := int32(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int32(req.PageSize)
	if pageSize < 1 {
		pageSize = 20
	}
	pbRes, err := grpcclient.AssetSvc.ListTransactions(ctx, &assetpb.ListTransactionsReq{
		UserId: userId, Page: page, PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	items := make([]assetv1.TransactionItem, 0)
	for _, tx := range pbRes.Transactions {
		items = append(items, assetv1.TransactionItem{
			Id: tx.Id, Type: tx.Type, Amount: tx.Amount,
			BalanceAfter: tx.BalanceAfter, Remark: tx.Remark,
		})
	}
	return &assetv1.ListTransactionsRes{List: items, Total: int(pbRes.Total)}, nil
}
