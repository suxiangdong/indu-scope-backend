// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// CategoryAttributesDao is the data access object for the table category_attributes.
type CategoryAttributesDao struct {
	table    string                    // table is the underlying table name of the DAO.
	group    string                    // group is the database configuration group name of the current DAO.
	columns  CategoryAttributesColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler        // handlers for customized model modification.
}

// CategoryAttributesColumns defines and stores column names for the table category_attributes.
type CategoryAttributesColumns struct {
	CategoryId   string //
	AttributeId  string //
	GroupId      string //
	IsRequired   string //
	IsFilterable string //
	IsSearchable string //
	SortOrder    string //
	GroupName    string //
	CreatedAt    string //
	UpdatedAt    string //
}

// categoryAttributesColumns holds the columns for the table category_attributes.
var categoryAttributesColumns = CategoryAttributesColumns{
	CategoryId:   "category_id",
	AttributeId:  "attribute_id",
	GroupId:      "group_id",
	IsRequired:   "is_required",
	IsFilterable: "is_filterable",
	IsSearchable: "is_searchable",
	SortOrder:    "sort_order",
	GroupName:    "group_name",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
}

// NewCategoryAttributesDao creates and returns a new DAO object for table data access.
func NewCategoryAttributesDao(handlers ...gdb.ModelHandler) *CategoryAttributesDao {
	return &CategoryAttributesDao{
		group:    "default",
		table:    "category_attributes",
		columns:  categoryAttributesColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *CategoryAttributesDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *CategoryAttributesDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *CategoryAttributesDao) Columns() CategoryAttributesColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *CategoryAttributesDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *CategoryAttributesDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *CategoryAttributesDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
