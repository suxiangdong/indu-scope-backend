package entity

import "time"

// CategoryAttributeGroups is the golang structure for table category_attribute_groups.
type CategoryAttributeGroups struct {
	Id         uint64    `json:"id" orm:"id" description:""`
	CategoryId uint64    `json:"categoryId" orm:"category_id" description:""`
	Code       string    `json:"code" orm:"code" description:""`
	Name       string    `json:"name" orm:"name" description:""`
	SortOrder  uint      `json:"sortOrder" orm:"sort_order" description:""`
	Status     uint      `json:"status" orm:"status" description:""`
	CreatedAt  time.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt  time.Time `json:"updatedAt" orm:"updated_at" description:""`
}
