// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// BuyerSalesChannels is the golang structure for table buyer_sales_channels.
type BuyerSalesChannels struct {
	Id           uint64    `json:"id"           orm:"id"            description:""` //
	UserId       uint64    `json:"userId"       orm:"user_id"       description:""` //
	ChannelType  uint      `json:"channelType"  orm:"channel_type"  description:""` //
	PlatformCode string    `json:"platformCode" orm:"platform_code" description:""` //
	ChannelName  string    `json:"channelName"  orm:"channel_name"  description:""` //
	ChannelUrl   string    `json:"channelUrl"   orm:"channel_url"   description:""` //
	ReviewStatus uint      `json:"reviewStatus" orm:"review_status" description:""` //
	ReviewReason string    `json:"reviewReason" orm:"review_reason" description:""` //
	ReviewedBy   uint64    `json:"reviewedBy"   orm:"reviewed_by"   description:""` //
	ReviewedAt   time.Time `json:"reviewedAt"   orm:"reviewed_at"   description:""` //
	SortOrder    uint      `json:"sortOrder"    orm:"sort_order"    description:""` //
	CreatedAt    time.Time `json:"createdAt"    orm:"created_at"    description:""` //
	UpdatedAt    time.Time `json:"updatedAt"    orm:"updated_at"    description:""` //
}
