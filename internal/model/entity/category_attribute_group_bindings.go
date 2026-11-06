package entity

import "time"

// CategoryAttributeGroupBindings links reusable groups to leaf categories.
type CategoryAttributeGroupBindings struct {
	Id         uint64    `json:"id" orm:"id"`
	CategoryId uint64    `json:"categoryId" orm:"category_id"`
	GroupId    uint64    `json:"groupId" orm:"group_id"`
	SortOrder  uint      `json:"sortOrder" orm:"sort_order"`
	Status     uint      `json:"status" orm:"status"`
	CreatedAt  time.Time `json:"createdAt" orm:"created_at"`
	UpdatedAt  time.Time `json:"updatedAt" orm:"updated_at"`
}
