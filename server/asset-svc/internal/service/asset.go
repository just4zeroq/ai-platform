package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"

	assetv1 "api/asset/v1"
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

func (s *AssetService) CompleteOrder(ctx context.Context, req *assetv1.CompleteOrderReq) (*assetv1.CompleteOrderRes, error) {
	if req.OrderId <= 0 {
		return nil, errors.New("order_id is required")
	}

	var order *assetv1.Order
	var balanceAfter float64

	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		o, err := tx.Model("orders").Ctx(ctx).Where("id", req.OrderId).One()
		if err != nil {
			return err
		}
		if o == nil {
			return errors.New("order not found")
		}
		if o["status"].String() != "pending" {
			return errors.New("order already completed or cancelled")
		}

		userId := o["user_id"].Int64()
		amount := o["total_amount"].Float64()

		// Mark order as completed
		_, err = tx.Model("orders").Ctx(ctx).
			Where("id", req.OrderId).
			Data(g.Map{
				"status":     "completed",
				"paid_at":    gdb.Raw("NOW()"),
				"updated_at": gdb.Raw("NOW()"),
			}).Update()
		if err != nil {
			return err
		}

		// Ensure balance exists
		if err := s.EnsureBalance(ctx, userId); err != nil {
			return err
		}

		// Add balance with optimistic locking
		amountStr := strconv.FormatFloat(amount, 'f', -1, 64)
		for attempt := 0; attempt < 3; attempt++ {
			record, err := tx.Model("balances").Ctx(ctx).Where("user_id", userId).One()
			if err != nil {
				return err
			}
			if record == nil {
				return errors.New("balance not found")
			}

			version := record["version"].Int()
			currentBalance := record["balance"].Float64()
			balanceAfter = currentBalance + amount

			result, err := tx.Model("balances").Ctx(ctx).
				Where("user_id", userId).
				Where("version", version).
				Data(g.Map{
					"balance":         balanceAfter,
					"total_recharged": gdb.Raw("total_recharged + " + amountStr),
					"version":         gdb.Raw("version + 1"),
					"updated_at":      gdb.Raw("NOW()"),
				}).Update()
			if err != nil {
				return err
			}
			rows, _ := result.RowsAffected()
			if rows > 0 {
				break
			}
			if attempt == 2 {
				return errors.New("concurrent update conflict")
			}
		}

		// Record transaction
		_, err = tx.Model("transactions").Ctx(ctx).Data(g.Map{
			"user_id":        userId,
			"type":           "recharge",
			"amount":         amount,
			"balance_before": balanceAfter - amount,
			"balance_after":  balanceAfter,
			"reference_type": "order",
			"reference_id":   req.OrderId,
			"remark":         "order completed: " + o["order_no"].String(),
		}).Insert()
		if err != nil {
			return err
		}

		order = &assetv1.Order{
			Id: o["id"].Int64(), OrderNo: o["order_no"].String(),
			UserId: userId, Type: o["type"].String(), Status: "completed",
			TotalAmount: amount, PaymentMethod: o["payment_method"].String(),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &assetv1.CompleteOrderRes{Order: order, BalanceAfter: balanceAfter}, nil
}

func (s *AssetService) ListOrders(ctx context.Context, req *assetv1.ListOrdersReq) (*assetv1.ListOrdersRes, error) {
	m := g.DB().Model("orders").Ctx(ctx)
	if req.UserId > 0 {
		m = m.Where("user_id", req.UserId)
	}
	if req.Status != "" {
		m = m.Where("status", req.Status)
	}

	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	var rows []gdb.Record
	err = m.Order("id DESC").Limit(pageSize).Offset((page-1)*pageSize).Scan(&rows)
	if err != nil {
		return nil, err
	}

	orders := make([]*assetv1.Order, 0, len(rows))
	for _, o := range rows {
		orders = append(orders, &assetv1.Order{
			Id: o["id"].Int64(), OrderNo: o["order_no"].String(),
			UserId: o["user_id"].Int64(), Type: o["type"].String(),
			Status: o["status"].String(), TotalAmount: o["total_amount"].Float64(),
			PaymentMethod: o["payment_method"].String(),
			PaymentTradeNo: o["payment_trade_no"].String(),
			PaidAt: o["paid_at"].String(),
		})
	}
	return &assetv1.ListOrdersRes{Orders: orders, Total: int32(total)}, nil
}

func (s *AssetService) RechargeBalance(ctx context.Context, req *assetv1.RechargeBalanceReq) (*assetv1.RechargeBalanceRes, error) {
	if req.UserId <= 0 {
		return nil, errors.New("user_id is required")
	}
	if req.Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	if err := s.EnsureBalance(ctx, req.UserId); err != nil {
		return nil, err
	}

	var balanceAfter float64
	amountStr := strconv.FormatFloat(req.Amount, 'f', -1, 64)

	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		for attempt := 0; attempt < 3; attempt++ {
			record, err := tx.Model("balances").Ctx(ctx).Where("user_id", req.UserId).One()
			if err != nil {
				return err
			}
			if record == nil {
				return errors.New("balance not found")
			}

			version := record["version"].Int()
			currentBalance := record["balance"].Float64()
			balanceAfter = currentBalance + req.Amount

			result, err := tx.Model("balances").Ctx(ctx).
				Where("user_id", req.UserId).
				Where("version", version).
				Data(g.Map{
					"balance":         balanceAfter,
					"total_recharged": gdb.Raw("total_recharged + " + amountStr),
					"version":         gdb.Raw("version + 1"),
					"updated_at":      gdb.Raw("NOW()"),
				}).Update()
			if err != nil {
				return err
			}
			rows, _ := result.RowsAffected()
			if rows > 0 {
				break
			}
			if attempt == 2 {
				return errors.New("concurrent update conflict")
			}
		}

		// Record transaction
		_, err := tx.Model("transactions").Ctx(ctx).Data(g.Map{
			"user_id":        req.UserId,
			"type":           "recharge",
			"amount":         req.Amount,
			"balance_before": balanceAfter - req.Amount,
			"balance_after":  balanceAfter,
			"reference_type": "admin",
			"remark":         req.Remark,
		}).Insert()
		return err
	})
	if err != nil {
		return nil, err
	}

	return &assetv1.RechargeBalanceRes{
		Balance: &assetv1.Balance{
			UserId: req.UserId, Balance: balanceAfter,
		},
	}, nil
}

func (s *AssetService) ListUsageRecords(ctx context.Context, req *assetv1.ListUsageRecordsReq) (*assetv1.ListUsageRecordsRes, error) {
	m := g.DB().Model("usage_records").Ctx(ctx)
	if req.UserId > 0 {
		m = m.Where("user_id", req.UserId)
	}

	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	pageSize := int(req.PageSize)
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	page := int(req.Page)
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	rows, err := m.Order("id DESC").Limit(pageSize).Offset(offset).All()
	if err != nil {
		return nil, err
	}

	records := make([]*assetv1.UsageRecord, 0, len(rows))
	for _, r := range rows {
		records = append(records, &assetv1.UsageRecord{
			Id:               r["id"].Int64(),
			UserId:           r["user_id"].Int64(),
			ApiKeyId:         r["api_key_id"].Int64(),
			ModelName:        r["model_name"].String(),
			PromptTokens:     int32(r["prompt_tokens"].Int()),
			CompletionTokens: int32(r["completion_tokens"].Int()),
			Quota:            r["quota"].Float64(),
			RequestId:        r["request_id"].String(),
			ChannelId:        r["channel_id"].Int64(),
			ChannelName:      r["channel_name"].String(),
			CreatedAt:        r["created_at"].String(),
		})
	}

	return &assetv1.ListUsageRecordsRes{Records: records, Total: int32(total)}, nil
}
