package do

import "github.com/gogf/gf/v2/frame/g"

// CategoryAttributeGroups is the data object for category attribute groups.
type CategoryAttributeGroups struct {
	g.Meta     `orm:"table:category_attribute_groups, do:true"`
	Id         any
	CategoryId any
	Code       any
	Name       any
	SortOrder  any
	Status     any
	CreatedAt  any
	UpdatedAt  any
}
