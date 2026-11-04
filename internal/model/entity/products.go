// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Products is the golang structure for table products.
type Products struct {
	Id              uint64    `json:"id"              orm:"id"               description:""` //
	SupplierUserId  uint64    `json:"supplierUserId"  orm:"supplier_user_id" description:""` //
	CategoryId      uint64    `json:"categoryId"      orm:"category_id"      description:""` //
	Name            string    `json:"name"            orm:"name"             description:""` //
	DescriptionHtml string    `json:"descriptionHtml" orm:"description_html" description:""` //
	ConditionCode   string    `json:"conditionCode"   orm:"condition_code"   description:""` //
	StockQuantity   float64   `json:"stockQuantity"   orm:"stock_quantity"   description:""` //
	StockUnitCode   string    `json:"stockUnitCode"   orm:"stock_unit_code"  description:""` //
	StockUpdatedAt  time.Time `json:"stockUpdatedAt"  orm:"stock_updated_at" description:""` //
	StockUpdatedBy  uint64    `json:"stockUpdatedBy"  orm:"stock_updated_by" description:""` //
	Status          uint      `json:"status"          orm:"status"           description:""` //
	CreatedAt       time.Time `json:"createdAt"       orm:"created_at"       description:""` //
	UpdatedAt       time.Time `json:"updatedAt"       orm:"updated_at"       description:""` //
}
