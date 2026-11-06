// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// CategoryAttributes is the golang structure of table category_attributes for DAO operations like Where/Data.
type CategoryAttributes struct {
	g.Meta       `orm:"table:category_attributes, do:true"`
	Id           any //
	CategoryId   any //
	AttributeId  any //
	GroupId      any //
	IsRequired   any //
	IsFilterable any //
	IsSearchable any //
	SortOrder    any //
	GroupName    any //
	CreatedAt    any //
	UpdatedAt    any //
}
