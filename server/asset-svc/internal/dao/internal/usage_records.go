// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UsageRecordsDao is the data access object for the table usage_records.
type UsageRecordsDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  UsageRecordsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// UsageRecordsColumns defines and stores column names for the table usage_records.
type UsageRecordsColumns struct {
	Id               string //
	UserId           string //
	ApiKeyId         string //
	ModelName        string //
	PromptTokens     string //
	CompletionTokens string //
	Quota            string //
	RequestId        string //
	ChannelId        string //
	ChannelName      string //
	CreatedAt        string //
}

// usageRecordsColumns holds the columns for the table usage_records.
var usageRecordsColumns = UsageRecordsColumns{
	Id:               "id",
	UserId:           "user_id",
	ApiKeyId:         "api_key_id",
	ModelName:        "model_name",
	PromptTokens:     "prompt_tokens",
	CompletionTokens: "completion_tokens",
	Quota:            "quota",
	RequestId:        "request_id",
	ChannelId:        "channel_id",
	ChannelName:      "channel_name",
	CreatedAt:        "created_at",
}

// NewUsageRecordsDao creates and returns a new DAO object for table data access.
func NewUsageRecordsDao(handlers ...gdb.ModelHandler) *UsageRecordsDao {
	return &UsageRecordsDao{
		group:    "default",
		table:    "usage_records",
		columns:  usageRecordsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *UsageRecordsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *UsageRecordsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *UsageRecordsDao) Columns() UsageRecordsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *UsageRecordsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *UsageRecordsDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *UsageRecordsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
