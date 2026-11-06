package v1

import "github.com/gogf/gf/v2/frame/g"

type CategoryInfo struct {
	ID        uint64 `json:"id"`
	ParentID  uint64 `json:"parent_id"`
	Name      string `json:"name"`
	SortOrder uint   `json:"sort_order"`
}

type AttributeOptionInfo struct {
	ID   uint64 `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type AttributeInfo struct {
	ID               uint64                `json:"id"`
	Code             string                `json:"code"`
	Name             string                `json:"name"`
	ValueType        uint                  `json:"value_type"`
	InputType        string                `json:"input_type"`
	Unit             string                `json:"unit,omitempty"`
	Placeholder      string                `json:"placeholder,omitempty"`
	HelpText         string                `json:"help_text,omitempty"`
	IsRequired       uint8                 `json:"is_required"`
	SortOrder        uint                  `json:"sort_order"`
	GroupDisplayName string                `json:"group_display_name,omitempty"`
	Options          []AttributeOptionInfo `json:"options,omitempty"`
}

type FormOptionsReq struct {
	g.Meta     `path:"/api/v1/supplier/products/form-options" tags:"SupplierProducts" method:"get" summary:"Get supplier product form options"`
	CategoryID *uint64 `json:"category_id" in:"query"`
}

type FormOptionsRes struct {
	g.Meta     `mime:"application/json"`
	Categories []CategoryInfo  `json:"categories"`
	Attributes []AttributeInfo `json:"attributes"`
}

type AttributeValueReq struct {
	AttributeID uint64   `json:"attribute_id"`
	Value       string   `json:"value"`
	OptionID    uint64   `json:"option_id,omitempty"`
	OptionIDs   []uint64 `json:"option_ids,omitempty"`
}

type CreateReq struct {
	g.Meta          `path:"/api/v1/supplier/products" tags:"SupplierProducts" method:"post" summary:"Create supplier product"`
	CategoryID      uint64              `json:"category_id" v:"required"`
	Name            string              `json:"name" v:"required"`
	DescriptionHTML string              `json:"description_html"`
	ConditionCode   string              `json:"condition_code"`
	StockQuantity   float64             `json:"stock_quantity"`
	StockUnitCode   string              `json:"stock_unit_code"`
	Attributes      []AttributeValueReq `json:"attributes"`
}

type ProductInfo struct {
	ID              uint64  `json:"id"`
	SupplierUserID  uint64  `json:"supplier_user_id"`
	CategoryID      uint64  `json:"category_id"`
	Name            string  `json:"name"`
	DescriptionHTML string  `json:"description_html,omitempty"`
	ConditionCode   string  `json:"condition_code"`
	StockQuantity   float64 `json:"stock_quantity"`
	StockUnitCode   string  `json:"stock_unit_code"`
	Status          uint    `json:"status"`
}

type CreateRes struct {
	g.Meta `mime:"application/json"`
	ProductInfo
}
