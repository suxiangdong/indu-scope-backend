// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// BuyerSalesChannels is the golang structure of table buyer_sales_channels for DAO operations like Where/Data.
type BuyerSalesChannels struct {
	g.Meta       `orm:"table:buyer_sales_channels, do:true"`
	Id           any //
	UserId       any //
	ChannelType  any //
	PlatformCode any //
	ChannelName  any //
	ChannelUrl   any //
	ReviewStatus any //
	ReviewReason any //
	ReviewedBy   any //
	ReviewedAt   any //
	SortOrder    any //
	CreatedAt    any //
	UpdatedAt    any //
}
