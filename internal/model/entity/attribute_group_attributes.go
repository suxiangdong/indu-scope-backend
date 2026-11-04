package entity

import "time"

// AttributeGroupAttributes links reusable groups to attributes.
type AttributeGroupAttributes struct {
	GroupId     uint64    `json:"groupId" orm:"group_id"`
	AttributeId uint64    `json:"attributeId" orm:"attribute_id"`
	SortOrder   uint      `json:"sortOrder" orm:"sort_order"`
	CreatedAt   time.Time `json:"createdAt" orm:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" orm:"updated_at"`
}
