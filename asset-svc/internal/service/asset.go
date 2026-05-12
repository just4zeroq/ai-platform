package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"

	assetv1 "asset-svc/api/asset/v1"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

type AssetService struct{}

// EnsureBalance creates a balance record if one doesn't exist for the user.
func (s *AssetService) EnsureBalance(ctx context.Context, userId int64) error {
	count, err := g.DB().Model("balances").Ctx(ctx).Where("user_id", userId).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		_, err = g.DB().Model("balances").Ctx(ctx).Data(g.Map{
			"user_id":         userId,
			"balance":         0,
			"total_recharged": 0,
			"total_consumed":  0,
			"version":         0,
		}).Insert()
	}
	return err
}

func (s *AssetService) GetBalance(ctx context.Context, req *assetv1.GetBalanceReq) (*assetv1.GetBalanceRes, error) {
	if err := s.EnsureBalance(ctx, req.UserId); err != nil {
		return nil, err
	}
	record, err := g.DB().Model("balances").Ctx(ctx).Where("user_id", req.UserId).One()
	if err != nil {
		return nil, err
	}
	if record == nil {
		return &assetv1.GetBalanceRes{
			Balance: &assetv1.Balance{
				UserId:         req.UserId,
				Balance:        0,
				TotalRecharged: 0,
				TotalConsumed:  0,
			},
		}, nil
	}
	return &assetv1.GetBalanceRes{
		Balance: &assetv1.Balance{
			UserId:         record["user_id"].Int64(),
			Balance:        record["balance"].Float64(),
			TotalRecharged: record["total_recharged"].Float64(),
			TotalConsumed:  record["total_consumed"].Float64(),
		},
	}, nil
}

func (s *AssetService) ListTransactions(ctx context.Context, req *assetv1.ListTransactionsReq) (*assetv1.ListTransactionsRes, error) {
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 20
	}
	total, err := g.DB().Model("transactions").Ctx(ctx).
		Where("user_id", req.UserId).Count()
	if err != nil {
		return nil, err
	}
	var records []gdb.Record
	err = g.DB().Model("transactions").Ctx(ctx).
		Where("user_id", req.UserId).
		Order("id DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Scan(&records)
	if err != nil {
		return nil, err
	}
	txs := make([]*assetv1.Transaction, 0)
	for _, r := range records {
		txs = append(txs, &assetv1.Transaction{
			Id:            r["id"].Int64(),
			UserId:        r["user_id"].Int64(),
			Type:          r["type"].String(),
			Amount:        r["amount"].Float64(),
			BalanceBefore: r["balance_before"].Float64(),
			BalanceAfter:  r["balance_after"].Float64(),
			ReferenceType: r["reference_type"].String(),
			ReferenceId:   r["reference_id"].Int64(),
			Remark:        r["remark"].String(),
		})
	}
	return &assetv1.ListTransactionsRes{Transactions: txs, Total: int32(total)}, nil
}

func (s *AssetService) ReportUsage(ctx context.Context, req *assetv1.ReportUsageReq) (*assetv1.ReportUsageRes, error) {
	if req.UserId <= 0 {
		return nil, errors.New("user_id is required")
	}
	if req.Quota <= 0 {
		return nil, errors.New("quota must be positive")
	}
	if math.IsNaN(req.Quota) || math.IsInf(req.Quota, 0) {
		return nil, errors.New("invalid quota value")
	}

	if err := s.EnsureBalance(ctx, req.UserId); err != nil {
		return nil, err
	}

	// Optimistic locking: retry up to 3 times on version conflict
	var balanceAfter float64
	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		for attempt := 0; attempt < 3; attempt++ {
			// Read current balance with version
			record, err := tx.Model("balances").Ctx(ctx).Where("user_id", req.UserId).One()
			if err != nil {
				return err
			}
			if record == nil {
				return errors.New("balance not found")
			}

			currentBalance := record["balance"].Float64()
			version := record["version"].Int()

			if currentBalance < req.Quota {
				return errors.New("insufficient balance")
			}

			balanceAfter = math.Max(currentBalance-req.Quota, 0)
			quotaStr := strconv.FormatFloat(req.Quota, 'f', -1, 64)

			// Atomic update with version check
			result, err := tx.Model("balances").Ctx(ctx).
				Where("user_id", req.UserId).
				Where("version", version).
				Data(g.Map{
					"balance":         balanceAfter,
					"total_consumed":  gdb.Raw("total_consumed + " + quotaStr),
					"version":         gdb.Raw("version + 1"),
					"updated_at":      gdb.Raw("NOW()"),
				}).Update()
			if err != nil {
				return err
			}
			rows, err := result.RowsAffected()
			if err != nil {
				return err
			}
			if rows > 0 {
				break // success
			}
			if attempt == 2 {
				return errors.New("concurrent update conflict, please retry")
			}
		}

		// Record transaction
		balanceBefore := balanceAfter + req.Quota
		_, err := tx.Model("transactions").Ctx(ctx).Data(g.Map{
			"user_id":        req.UserId,
			"type":           "consume",
			"amount":         req.Quota,
			"balance_before": balanceBefore,
			"balance_after":  balanceAfter,
			"reference_type": "usage",
			"remark":         "usage: " + req.ModelName,
		}).Insert()
		if err != nil {
			return err
		}

		// Record usage
		_, err = tx.Model("usage_records").Ctx(ctx).Data(g.Map{
			"user_id":           req.UserId,
			"api_key_id":        req.ApiKeyId,
			"model_name":        req.ModelName,
			"prompt_tokens":     req.PromptTokens,
			"completion_tokens": req.CompletionTokens,
			"quota":             req.Quota,
			"request_id":        req.RequestId,
			"channel_id":        req.ChannelId,
			"channel_name":      req.ChannelName,
		}).Insert()
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &assetv1.ReportUsageRes{BalanceAfter: balanceAfter}, nil
}

func (s *AssetService) CreateOrder(ctx context.Context, req *assetv1.CreateOrderReq) (*assetv1.CreateOrderRes, error) {
	if req.UserId == 0 {
		return nil, errors.New("user_id is required")
	}
	if req.Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	orderNo := fmt.Sprintf("ORD%s%04d", time.Now().Format("20060102150405"), rand.Intn(10000))
	result, err := g.DB().Model("orders").Ctx(ctx).Data(g.Map{
		"order_no":       orderNo,
		"user_id":        req.UserId,
		"type":           req.Type,
		"status":         "pending",
		"total_amount":   req.Amount,
		"payment_method": req.PaymentMethod,
	}).Insert()
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return &assetv1.CreateOrderRes{
		Order: &assetv1.Order{
			Id: id, OrderNo: orderNo, UserId: req.UserId,
			Type: req.Type, Status: "pending", TotalAmount: req.Amount,
				PaymentMethod: req.PaymentMethod,
		},
	}, nil
}
