package market

import (
	"context"

	marketv1 "market-svc/api/market/v1"
)

type Controller struct {
	marketv1.UnimplementedMarketServiceServer
}

func (c *Controller) CreateListing(ctx context.Context, req *marketv1.CreateListingReq) (*marketv1.CreateListingRes, error) {
	return nil, nil
}

func (c *Controller) ListListings(ctx context.Context, req *marketv1.ListListingsReq) (*marketv1.ListListingsRes, error) {
	return nil, nil
}

func (c *Controller) BuyProduct(ctx context.Context, req *marketv1.BuyProductReq) (*marketv1.BuyProductRes, error) {
	return nil, nil
}

func (c *Controller) ListTrades(ctx context.Context, req *marketv1.ListTradesReq) (*marketv1.ListTradesRes, error) {
	return nil, nil
}
