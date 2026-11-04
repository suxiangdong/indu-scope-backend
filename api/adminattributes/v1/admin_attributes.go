package v1

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type AttributeInfo struct {
	ID              int64     `json:"id" dc:"Attribute ID"`
	Code            string    `json:"code" dc:"Machine-readable attribute code"`
	Name            string    `json:"name" dc:"Attribute name"`
	ValueType       uint8     `json:"value_type" dc:"1 text, 2 integer, 3 decimal, 4 boolean, 5 date, 6 JSON"`
	InputType       string    `json:"input_type" dc:"Input control type"`
	Unit            string    `json:"unit"`
	Placeholder     string    `json:"placeholder"`
	ValidationRules string    `json:"validation_rules"`
	HelpText        string    `json:"help_text"`
	Status          uint8     `json:"status" dc:"0 disabled, 1 active"`
	OptionCount     uint      `json:"option_count"`
	GroupCount      uint      `json:"group_count"`
	GroupIDs        []int64   `json:"group_ids"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ListReq struct {
	g.Meta `path:"/api/v1/admin/attributes" tags:"AdminAttributes" method:"get" summary:"List attributes"`
}

type ListRes struct {
	g.Meta `mime:"application/json"`
	Items  []AttributeInfo `json:"items"`
}

type CreateReq struct {
	g.Meta          `path:"/api/v1/admin/attributes/create" tags:"AdminAttributes" method:"post" summary:"Create attribute"`
	Code            string   `json:"code" v:"required"`
	Name            string   `json:"name" v:"required"`
	ValueType       uint     `json:"value_type" v:"required"`
	InputType       string   `json:"input_type" v:"required"`
	Unit            string   `json:"unit"`
	Placeholder     string   `json:"placeholder"`
	ValidationRules string   `json:"validation_rules"`
	HelpText        string   `json:"help_text"`
	Status          *uint    `json:"status"`
	GroupIDs        []uint64 `json:"group_ids"`
}

type CreateRes struct {
	g.Meta `mime:"application/json"`
	AttributeInfo
}

type UpdateReq struct {
	g.Meta          `path:"/api/v1/admin/attributes/update" tags:"AdminAttributes" method:"post" summary:"Update attribute"`
	ID              uint64   `json:"id" v:"required"`
	Code            string   `json:"code" v:"required"`
	Name            string   `json:"name" v:"required"`
	ValueType       uint     `json:"value_type" v:"required"`
	InputType       string   `json:"input_type" v:"required"`
	Unit            string   `json:"unit"`
	Placeholder     string   `json:"placeholder"`
	ValidationRules string   `json:"validation_rules"`
	HelpText        string   `json:"help_text"`
	Status          *uint    `json:"status"`
	GroupIDs        []uint64 `json:"group_ids"`
}

type UpdateRes struct {
	g.Meta `mime:"application/json"`
	AttributeInfo
}

type DeleteReq struct {
	g.Meta `path:"/api/v1/admin/attributes/delete" tags:"AdminAttributes" method:"post" summary:"Delete attribute"`
	ID     uint64 `json:"id" v:"required"`
}

type DeleteRes struct {
	g.Meta  `mime:"application/json"`
	Success bool `json:"success"`
}

type AttributeOptionInfo struct {
	ID          int64     `json:"id"`
	AttributeID int64     `json:"attribute_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	SortOrder   uint      `json:"sort_order"`
	Status      uint8     `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ListOptionsReq struct {
	g.Meta      `path:"/api/v1/admin/attributes/options" tags:"AdminAttributes" method:"get" summary:"List attribute options"`
	AttributeID uint64 `json:"attribute_id" v:"required"`
}

type ListOptionsRes struct {
	g.Meta `mime:"application/json"`
	Items  []AttributeOptionInfo `json:"items"`
}

type CreateOptionReq struct {
	g.Meta      `path:"/api/v1/admin/attributes/options/create" tags:"AdminAttributes" method:"post" summary:"Create attribute option"`
	AttributeID uint64 `json:"attribute_id" v:"required"`
	Code        string `json:"code" v:"required"`
	Name        string `json:"name" v:"required"`
	SortOrder   uint   `json:"sort_order"`
	Status      *uint  `json:"status"`
}

type CreateOptionRes struct {
	g.Meta `mime:"application/json"`
	AttributeOptionInfo
}

type UpdateOptionReq struct {
	g.Meta    `path:"/api/v1/admin/attributes/options/update" tags:"AdminAttributes" method:"post" summary:"Update attribute option"`
	ID        uint64 `json:"id" v:"required"`
	Code      string `json:"code" v:"required"`
	Name      string `json:"name" v:"required"`
	SortOrder uint   `json:"sort_order"`
	Status    *uint  `json:"status"`
}

type UpdateOptionRes struct {
	g.Meta `mime:"application/json"`
	AttributeOptionInfo
}

type DeleteOptionReq struct {
	g.Meta `path:"/api/v1/admin/attributes/options/delete" tags:"AdminAttributes" method:"post" summary:"Delete attribute option"`
	ID     uint64 `json:"id" v:"required"`
}

type DeleteOptionRes struct {
	g.Meta  `mime:"application/json"`
	Success bool `json:"success"`
}

type CategoryAttributeInfo struct {
	CategoryID    int64     `json:"category_id"`
	CategoryName  string    `json:"category_name"`
	AttributeID   int64     `json:"attribute_id"`
	AttributeCode string    `json:"attribute_code"`
	AttributeName string    `json:"attribute_name"`
	GroupID       *int64    `json:"group_id,omitempty"`
	GroupCode     string    `json:"group_code,omitempty"`
	GroupName     string    `json:"group_name"`
	IsRequired    uint8     `json:"is_required"`
	IsFilterable  uint8     `json:"is_filterable"`
	IsSearchable  uint8     `json:"is_searchable"`
	SortOrder     uint      `json:"sort_order"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ListCategoryBindingsReq struct {
	g.Meta      `path:"/api/v1/admin/category-attributes" tags:"AdminAttributes" method:"get" summary:"List category attribute bindings"`
	CategoryID  *uint64 `json:"category_id"`
	AttributeID *uint64 `json:"attribute_id"`
}

type ListCategoryBindingsRes struct {
	g.Meta `mime:"application/json"`
	Items  []CategoryAttributeInfo `json:"items"`
}

type SaveCategoryBindingReq struct {
	g.Meta       `path:"/api/v1/admin/category-attributes/save" tags:"AdminAttributes" method:"post" summary:"Save category attribute binding"`
	CategoryID   uint64  `json:"category_id" v:"required"`
	AttributeID  uint64  `json:"attribute_id" v:"required"`
	GroupID      *uint64 `json:"group_id"`
	IsRequired   uint    `json:"is_required"`
	IsFilterable uint    `json:"is_filterable"`
	IsSearchable uint    `json:"is_searchable"`
	SortOrder    uint    `json:"sort_order"`
	GroupName    string  `json:"group_name"`
}

type SaveCategoryBindingRes struct {
	g.Meta `mime:"application/json"`
	CategoryAttributeInfo
}

type DeleteCategoryBindingReq struct {
	g.Meta      `path:"/api/v1/admin/category-attributes/delete" tags:"AdminAttributes" method:"post" summary:"Delete category attribute binding"`
	CategoryID  uint64 `json:"category_id" v:"required"`
	AttributeID uint64 `json:"attribute_id" v:"required"`
}

type DeleteCategoryBindingRes struct {
	g.Meta  `mime:"application/json"`
	Success bool `json:"success"`
}

type AttributeGroupInfo struct {
	ID             int64     `json:"id"`
	Code           string    `json:"code"`
	Name           string    `json:"name"`
	SortOrder      uint      `json:"sort_order"`
	Status         uint8     `json:"status"`
	AttributeCount uint      `json:"attribute_count"`
	CategoryCount  uint      `json:"category_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ListGroupsReq struct {
	g.Meta `path:"/api/v1/admin/attribute-groups" tags:"AdminAttributes" method:"get" summary:"List reusable attribute groups"`
}

type ListGroupsRes struct {
	g.Meta `mime:"application/json"`
	Items  []AttributeGroupInfo `json:"items"`
}

type CreateGroupReq struct {
	g.Meta    `path:"/api/v1/admin/attribute-groups/create" tags:"AdminAttributes" method:"post" summary:"Create reusable attribute group"`
	Code      string `json:"code" v:"required"`
	Name      string `json:"name" v:"required"`
	SortOrder uint   `json:"sort_order"`
	Status    *uint  `json:"status"`
}

type CreateGroupRes struct {
	g.Meta `mime:"application/json"`
	AttributeGroupInfo
}

type UpdateGroupReq struct {
	g.Meta    `path:"/api/v1/admin/attribute-groups/update" tags:"AdminAttributes" method:"post" summary:"Update reusable attribute group"`
	ID        uint64 `json:"id" v:"required"`
	Code      string `json:"code" v:"required"`
	Name      string `json:"name" v:"required"`
	SortOrder uint   `json:"sort_order"`
	Status    *uint  `json:"status"`
}

type UpdateGroupRes struct {
	g.Meta `mime:"application/json"`
	AttributeGroupInfo
}

type DeleteGroupReq struct {
	g.Meta `path:"/api/v1/admin/attribute-groups/delete" tags:"AdminAttributes" method:"post" summary:"Delete reusable attribute group"`
	ID     uint64 `json:"id" v:"required"`
}

type DeleteGroupRes struct {
	g.Meta  `mime:"application/json"`
	Success bool `json:"success"`
}

type GroupAttributeInfo struct {
	GroupID     int64     `json:"group_id"`
	AttributeID int64     `json:"attribute_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	ValueType   uint8     `json:"value_type"`
	InputType   string    `json:"input_type"`
	SortOrder   uint      `json:"sort_order"`
	Status      uint8     `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ListGroupAttributesReq struct {
	g.Meta  `path:"/api/v1/admin/attribute-groups/attributes" tags:"AdminAttributes" method:"get" summary:"List attributes in a group"`
	GroupID uint64 `json:"group_id" v:"required"`
}

type ListGroupAttributesRes struct {
	g.Meta `mime:"application/json"`
	Items  []GroupAttributeInfo `json:"items"`
}

type SaveGroupAttributeReq struct {
	g.Meta      `path:"/api/v1/admin/attribute-groups/attributes/save" tags:"AdminAttributes" method:"post" summary:"Add or update a group attribute"`
	GroupID     uint64 `json:"group_id" v:"required"`
	AttributeID uint64 `json:"attribute_id" v:"required"`
	SortOrder   uint   `json:"sort_order"`
}

type SaveGroupAttributeRes struct {
	g.Meta `mime:"application/json"`
	GroupAttributeInfo
}

type DeleteGroupAttributeReq struct {
	g.Meta      `path:"/api/v1/admin/attribute-groups/attributes/delete" tags:"AdminAttributes" method:"post" summary:"Remove an attribute from a group"`
	GroupID     uint64 `json:"group_id" v:"required"`
	AttributeID uint64 `json:"attribute_id" v:"required"`
}

type DeleteGroupAttributeRes struct {
	g.Meta  `mime:"application/json"`
	Success bool `json:"success"`
}

type CategoryGroupInfo struct {
	CategoryID     int64     `json:"category_id"`
	CategoryName   string    `json:"category_name"`
	GroupID        int64     `json:"group_id"`
	GroupCode      string    `json:"group_code"`
	GroupName      string    `json:"group_name"`
	SortOrder      uint      `json:"sort_order"`
	Status         uint8     `json:"status"`
	AttributeCount uint      `json:"attribute_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ListCategoryGroupsReq struct {
	g.Meta     `path:"/api/v1/admin/category-groups" tags:"AdminAttributes" method:"get" summary:"List groups bound to a category"`
	CategoryID uint64 `json:"category_id" v:"required"`
}

type ListCategoryGroupsRes struct {
	g.Meta `mime:"application/json"`
	Items  []CategoryGroupInfo `json:"items"`
}

type BindCategoryGroupReq struct {
	g.Meta     `path:"/api/v1/admin/category-groups/bind" tags:"AdminAttributes" method:"post" summary:"Bind a group to a category"`
	CategoryID uint64 `json:"category_id" v:"required"`
	GroupID    uint64 `json:"group_id" v:"required"`
	SortOrder  uint   `json:"sort_order"`
	Status     *uint  `json:"status"`
}

type BindCategoryGroupRes struct {
	g.Meta `mime:"application/json"`
	CategoryGroupInfo
}

type UnbindCategoryGroupReq struct {
	g.Meta     `path:"/api/v1/admin/category-groups/unbind" tags:"AdminAttributes" method:"post" summary:"Unbind a group from a category"`
	CategoryID uint64 `json:"category_id" v:"required"`
	GroupID    uint64 `json:"group_id" v:"required"`
}

type UnbindCategoryGroupRes struct {
	g.Meta  `mime:"application/json"`
	Success bool `json:"success"`
}
