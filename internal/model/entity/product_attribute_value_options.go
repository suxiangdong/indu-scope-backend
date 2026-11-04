// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// ProductAttributeValueOptions is the golang structure for table product_attribute_value_options.
type ProductAttributeValueOptions struct {
	ValueId   uint64    `json:"valueId"   orm:"value_id"   description:""` //
	OptionId  uint64    `json:"optionId"  orm:"option_id"  description:""` //
	CreatedAt time.Time `json:"createdAt" orm:"created_at" description:""` //
	UpdatedAt time.Time `json:"updatedAt" orm:"updated_at" description:""` //
}
