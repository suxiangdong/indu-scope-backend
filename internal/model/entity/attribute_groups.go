package entity

import "time"

// AttributeGroups is the reusable global attribute group entity.
type AttributeGroups struct {
	Id          uint64    `json:"id" orm:"id"`
	Code        string    `json:"code" orm:"code"`
	Name        string    `json:"name" orm:"name"`
	DisplayName string    `json:"displayName" orm:"display_name"`
	SortOrder   uint      `json:"sortOrder" orm:"sort_order"`
	Status      uint      `json:"status" orm:"status"`
	CreatedAt   time.Time `json:"createdAt" orm:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" orm:"updated_at"`
}
