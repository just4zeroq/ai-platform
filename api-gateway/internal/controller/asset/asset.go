package asset

import (
	"context"

	assetv1 "api-gateway/api/asset/v1"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

type Controller struct{}

func (c *Controller) GetBalance(ctx context.Context, req *assetv1.GetBalanceReq) (res *assetv1.GetBalanceRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented, "not implemented")
}

func (c *Controller) ListTransactions(ctx context.Context, req *assetv1.ListTransactionsReq) (res *assetv1.ListTransactionsRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented, "not implemented")
}
