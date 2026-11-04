// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Categories is the golang structure for table categories.
type Categories struct {
	Id        uint64    `json:"id"        orm:"id"         description:""` //
	ParentId  uint64    `json:"parentId"  orm:"parent_id"  description:""` //
	Name      string    `json:"name"      orm:"name"       description:""` //
	SortOrder uint      `json:"sortOrder" orm:"sort_order" description:""` //
	Status    uint      `json:"status"    orm:"status"     description:""` //
	CreatedAt time.Time `json:"createdAt" orm:"created_at" description:""` //
	UpdatedAt time.Time `json:"updatedAt" orm:"updated_at" description:""` //
}
