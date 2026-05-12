package asset

import (
	"github.com/gogf/gf/v2/frame/g"
)

type GetBalanceReq struct {
	g.Meta `path:"/asset/balance" method:"GET" summary:"Get Balance" tags:"Asset"`
}

type GetBalanceRes struct {
	Balance float64 `json:"balance"`
}

type ListTransactionsReq struct {
	g.Meta   `path:"/asset/transactions" method:"GET" summary:"List Transactions" tags:"Asset"`
	Page     int `json:"page" dc:"Page number"`
	PageSize int `json:"page_size" dc:"Page size"`
}

type ListTransactionsRes struct {
	List  []TransactionItem `json:"list"`
	Total int               `json:"total"`
}

type TransactionItem struct {
	Id            int64   `json:"id"`
	Type          string  `json:"type"`
	Amount        float64 `json:"amount"`
	BalanceAfter  float64 `json:"balance_after"`
	Remark        string  `json:"remark"`
	CreatedAt     string  `json:"created_at"`
}
