// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// ProductAttributeValues is the golang structure for table product_attribute_values.
type ProductAttributeValues struct {
	Id            uint64    `json:"id"            orm:"id"              description:""` //
	ProductId     uint64    `json:"productId"     orm:"product_id"      description:""` //
	AttributeId   uint64    `json:"attributeId"   orm:"attribute_id"    description:""` //
	ValueNumber   float64   `json:"valueNumber"   orm:"value_number"    description:""` //
	ValueText     string    `json:"valueText"     orm:"value_text"      description:""` //
	ValueBool     uint      `json:"valueBool"     orm:"value_bool"      description:""` //
	ValueOptionId uint64    `json:"valueOptionId" orm:"value_option_id" description:""` //
	ValueJson     string    `json:"valueJson"     orm:"value_json"      description:""` //
	CreatedAt     time.Time `json:"createdAt"     orm:"created_at"      description:""` //
	UpdatedAt     time.Time `json:"updatedAt"     orm:"updated_at"      description:""` //
}
