package do

import "github.com/gogf/gf/v2/frame/g"

// AttributeGroups is the data object for attribute_groups.
type AttributeGroups struct {
	g.Meta    `orm:"table:attribute_groups, do:true"`
	Id        any
	Code      any
	Name      any
	SortOrder any
	Status    any
	CreatedAt any
	UpdatedAt any
}
