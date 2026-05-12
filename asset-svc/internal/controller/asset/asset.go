package asset

import (
	"context"

	assetv1 "asset-svc/api/asset/v1"
)

type Controller struct {
	assetv1.UnimplementedAssetServiceServer
}

func (c *Controller) CreateOrder(ctx context.Context, req *assetv1.CreateOrderReq) (*assetv1.CreateOrderRes, error) {
	return nil, nil
}

func (c *Controller) GetBalance(ctx context.Context, req *assetv1.GetBalanceReq) (*assetv1.GetBalanceRes, error) {
	return nil, nil
}

func (c *Controller) ReportUsage(ctx context.Context, req *assetv1.ReportUsageReq) (*assetv1.ReportUsageRes, error) {
	return nil, nil
}

func (c *Controller) ListTransactions(ctx context.Context, req *assetv1.ListTransactionsReq) (*assetv1.ListTransactionsRes, error) {
	return nil, nil
}
