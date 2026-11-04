// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// AttributeOptions is the golang structure for table attribute_options.
type AttributeOptions struct {
	Id          uint64    `json:"id"          orm:"id"           description:""` //
	AttributeId uint64    `json:"attributeId" orm:"attribute_id" description:""` //
	Code        string    `json:"code"        orm:"code"         description:""` //
	Name        string    `json:"name"        orm:"name"         description:""` //
	SortOrder   uint      `json:"sortOrder"   orm:"sort_order"   description:""` //
	Status      uint      `json:"status"      orm:"status"       description:""` //
	CreatedAt   time.Time `json:"createdAt"   orm:"created_at"   description:""` //
	UpdatedAt   time.Time `json:"updatedAt"   orm:"updated_at"   description:""` //
}
