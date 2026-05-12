package asset

import (
	"context"

	assetv1 "api/asset/v1"
	"asset-svc/internal/service"
)

type Controller struct {
	assetv1.UnimplementedAssetServiceServer
	svc *service.AssetService
}

func New(svc *service.AssetService) *Controller {
	return &Controller{svc: svc}
}

func (c *Controller) GetBalance(ctx context.Context, req *assetv1.GetBalanceReq) (*assetv1.GetBalanceRes, error) {
	return c.svc.GetBalance(ctx, req)
}

func (c *Controller) ListTransactions(ctx context.Context, req *assetv1.ListTransactionsReq) (*assetv1.ListTransactionsRes, error) {
	return c.svc.ListTransactions(ctx, req)
}

func (c *Controller) ReportUsage(ctx context.Context, req *assetv1.ReportUsageReq) (*assetv1.ReportUsageRes, error) {
	return c.svc.ReportUsage(ctx, req)
}

func (c *Controller) CreateOrder(ctx context.Context, req *assetv1.CreateOrderReq) (*assetv1.CreateOrderRes, error) {
	return c.svc.CreateOrder(ctx, req)
}
