// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// CategoryAttributes is the golang structure for table category_attributes.
type CategoryAttributes struct {
	CategoryId   uint64    `json:"categoryId"   orm:"category_id"   description:""` //
	AttributeId  uint64    `json:"attributeId"  orm:"attribute_id"  description:""` //
	GroupId      uint64    `json:"groupId"      orm:"group_id"      description:""` //
	IsRequired   uint      `json:"isRequired"   orm:"is_required"   description:""` //
	IsFilterable uint      `json:"isFilterable" orm:"is_filterable" description:""` //
	IsSearchable uint      `json:"isSearchable" orm:"is_searchable" description:""` //
	SortOrder    uint      `json:"sortOrder"    orm:"sort_order"    description:""` //
	GroupName    string    `json:"groupName"    orm:"group_name"    description:""` //
	CreatedAt    time.Time `json:"createdAt"    orm:"created_at"    description:""` //
	UpdatedAt    time.Time `json:"updatedAt"    orm:"updated_at"    description:""` //
}
