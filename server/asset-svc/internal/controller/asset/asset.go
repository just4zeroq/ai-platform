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

func (c *Controller) ListUsageRecords(ctx context.Context, req *assetv1.ListUsageRecordsReq) (*assetv1.ListUsageRecordsRes, error) {
	return c.svc.ListUsageRecords(ctx, req)
}

func (c *Controller) CompleteOrder(ctx context.Context, req *assetv1.CompleteOrderReq) (*assetv1.CompleteOrderRes, error) {
	return c.svc.CompleteOrder(ctx, req)
}

func (c *Controller) ListOrders(ctx context.Context, req *assetv1.ListOrdersReq) (*assetv1.ListOrdersRes, error) {
	return c.svc.ListOrders(ctx, req)
}

func (c *Controller) RechargeBalance(ctx context.Context, req *assetv1.RechargeBalanceReq) (*assetv1.RechargeBalanceRes, error) {
	return c.svc.RechargeBalance(ctx, req)
}
