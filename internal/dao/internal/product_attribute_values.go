// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ProductAttributeValuesDao is the data access object for the table product_attribute_values.
type ProductAttributeValuesDao struct {
	table    string                        // table is the underlying table name of the DAO.
	group    string                        // group is the database configuration group name of the current DAO.
	columns  ProductAttributeValuesColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler            // handlers for customized model modification.
}

// ProductAttributeValuesColumns defines and stores column names for the table product_attribute_values.
type ProductAttributeValuesColumns struct {
	Id            string //
	ProductId     string //
	AttributeId   string //
	ValueNumber   string //
	ValueText     string //
	ValueBool     string //
	ValueOptionId string //
	ValueJson     string //
	CreatedAt     string //
	UpdatedAt     string //
}

// productAttributeValuesColumns holds the columns for the table product_attribute_values.
var productAttributeValuesColumns = ProductAttributeValuesColumns{
	Id:            "id",
	ProductId:     "product_id",
	AttributeId:   "attribute_id",
	ValueNumber:   "value_number",
	ValueText:     "value_text",
	ValueBool:     "value_bool",
	ValueOptionId: "value_option_id",
	ValueJson:     "value_json",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
}

// NewProductAttributeValuesDao creates and returns a new DAO object for table data access.
func NewProductAttributeValuesDao(handlers ...gdb.ModelHandler) *ProductAttributeValuesDao {
	return &ProductAttributeValuesDao{
		group:    "default",
		table:    "product_attribute_values",
		columns:  productAttributeValuesColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ProductAttributeValuesDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ProductAttributeValuesDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ProductAttributeValuesDao) Columns() ProductAttributeValuesColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ProductAttributeValuesDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ProductAttributeValuesDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ProductAttributeValuesDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
