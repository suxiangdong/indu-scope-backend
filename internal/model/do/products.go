// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Products is the golang structure of table products for DAO operations like Where/Data.
type Products struct {
	g.Meta          `orm:"table:products, do:true"`
	Id              any //
	SupplierUserId  any //
	CategoryId      any //
	Name            any //
	DescriptionHtml any //
	ConditionCode   any //
	StockQuantity   any //
	StockUnitCode   any //
	StockUpdatedAt  any //
	StockUpdatedBy  any //
	Status          any //
	CreatedAt       any //
	UpdatedAt       any //
}
