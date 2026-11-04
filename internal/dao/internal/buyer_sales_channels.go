// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// BuyerSalesChannelsDao is the data access object for the table buyer_sales_channels.
type BuyerSalesChannelsDao struct {
	table    string                    // table is the underlying table name of the DAO.
	group    string                    // group is the database configuration group name of the current DAO.
	columns  BuyerSalesChannelsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler        // handlers for customized model modification.
}

// BuyerSalesChannelsColumns defines and stores column names for the table buyer_sales_channels.
type BuyerSalesChannelsColumns struct {
	Id           string //
	UserId       string //
	ChannelType  string //
	PlatformCode string //
	ChannelName  string //
	ChannelUrl   string //
	ReviewStatus string //
	ReviewReason string //
	ReviewedBy   string //
	ReviewedAt   string //
	SortOrder    string //
	CreatedAt    string //
	UpdatedAt    string //
}

// buyerSalesChannelsColumns holds the columns for the table buyer_sales_channels.
var buyerSalesChannelsColumns = BuyerSalesChannelsColumns{
	Id:           "id",
	UserId:       "user_id",
	ChannelType:  "channel_type",
	PlatformCode: "platform_code",
	ChannelName:  "channel_name",
	ChannelUrl:   "channel_url",
	ReviewStatus: "review_status",
	ReviewReason: "review_reason",
	ReviewedBy:   "reviewed_by",
	ReviewedAt:   "reviewed_at",
	SortOrder:    "sort_order",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
}

// NewBuyerSalesChannelsDao creates and returns a new DAO object for table data access.
func NewBuyerSalesChannelsDao(handlers ...gdb.ModelHandler) *BuyerSalesChannelsDao {
	return &BuyerSalesChannelsDao{
		group:    "default",
		table:    "buyer_sales_channels",
		columns:  buyerSalesChannelsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *BuyerSalesChannelsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *BuyerSalesChannelsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *BuyerSalesChannelsDao) Columns() BuyerSalesChannelsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *BuyerSalesChannelsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *BuyerSalesChannelsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *BuyerSalesChannelsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
