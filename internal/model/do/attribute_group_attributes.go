package do

import "github.com/gogf/gf/v2/frame/g"

// AttributeGroupAttributes is the data object for attribute_group_attributes.
type AttributeGroupAttributes struct {
	g.Meta      `orm:"table:attribute_group_attributes, do:true"`
	GroupId     any
	AttributeId any
	SortOrder   any
	CreatedAt   any
	UpdatedAt   any
}
