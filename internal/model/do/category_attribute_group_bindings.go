package do

import "github.com/gogf/gf/v2/frame/g"

// CategoryAttributeGroupBindings is the data object for category_attribute_group_bindings.
type CategoryAttributeGroupBindings struct {
	g.Meta     `orm:"table:category_attribute_group_bindings, do:true"`
	Id         any
	CategoryId any
	GroupId    any
	SortOrder  any
	Status     any
	CreatedAt  any
	UpdatedAt  any
}
