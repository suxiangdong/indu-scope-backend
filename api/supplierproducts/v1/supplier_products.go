package v1

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

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
	StockUnits []StockUnitInfo `json:"stock_units"`
}

type ListReq struct {
	g.Meta   `path:"/api/v1/supplier/products" tags:"SupplierProducts" method:"get" summary:"List supplier products"`
	Page     uint `json:"page" in:"query"`
	PageSize uint `json:"page_size" in:"query"`
}

type DetailReq struct {
	g.Meta `path:"/api/v1/supplier/products/detail" tags:"SupplierProducts" method:"get" summary:"Get own product detail"`
	ID     uint64 `json:"id" in:"query" v:"required"`
}

type ListRes struct {
	g.Meta   `mime:"application/json"`
	Items    []ProductInfo `json:"items"`
	Total    uint64        `json:"total"`
	Page     uint          `json:"page"`
	PageSize uint          `json:"page_size"`
}

type StockUnitInfo struct {
	ID        uint64 `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	SortOrder uint   `json:"sort_order"`
}

type AttributeValueReq struct {
	AttributeID uint64   `json:"attribute_id"`
	Value       string   `json:"value"`
	OptionID    uint64   `json:"option_id,omitempty"`
	OptionIDs   []uint64 `json:"option_ids,omitempty"`
}

type CreateReq struct {
	g.Meta        `path:"/api/v1/supplier/products" tags:"SupplierProducts" method:"post" summary:"Create supplier product"`
	CategoryID    uint64              `json:"category_id" v:"required"`
	Name          string              `json:"name" v:"required"`
	StockQuantity float64             `json:"stock_quantity"`
	StockUnitID   uint64              `json:"stock_unit_id" v:"required"`
	ImagePaths    []string            `json:"image_paths"`
	Attributes    []AttributeValueReq `json:"attributes"`
}

type ProductImageInfo struct {
	ID           uint64    `json:"id"`
	Endpoint     string    `json:"endpoint"`
	Bucket       string    `json:"bucket"`
	Path         string    `json:"path"`
	URL          string    `json:"url,omitempty"`
	OriginalName string    `json:"original_name"`
	ContentType  string    `json:"content_type"`
	Size         int64     `json:"size"`
	SortOrder    uint      `json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
}

type ProductInfo struct {
	ID             uint64               `json:"id"`
	SupplierUserID uint64               `json:"supplier_user_id"`
	CategoryID     uint64               `json:"category_id"`
	Name           string               `json:"name"`
	StockQuantity  float64              `json:"stock_quantity"`
	StockUnitID    uint64               `json:"stock_unit_id"`
	StockUnitCode  string               `json:"stock_unit_code"`
	StockUnitName  string               `json:"stock_unit_name"`
	CategoryName   string               `json:"category_name,omitempty"`
	CoverURL       string               `json:"cover_url,omitempty"`
	ImageCount     uint                 `json:"image_count"`
	Images         []ProductImageInfo   `json:"images"`
	Attributes     []AttributeValueInfo `json:"attributes,omitempty"`
	CreatedAt      time.Time            `json:"created_at"`
	Status         uint                 `json:"status"`
}

type AttributeValueInfo struct {
	AttributeID uint64 `json:"attribute_id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Unit        string `json:"unit,omitempty"`
	Value       string `json:"value"`
}

type ContactInfo struct {
	Name     string `json:"name"`
	Position string `json:"position,omitempty"`
	Phone    string `json:"phone"`
}

type DetailRes struct {
	g.Meta `mime:"application/json"`
	ProductInfo
	SupplierContacts []ContactInfo `json:"supplier_contacts"`
}

type CreateRes struct {
	g.Meta `mime:"application/json"`
	ProductInfo
}

type UpdateStockReq struct {
	g.Meta        `path:"/api/v1/supplier/products/stock" tags:"SupplierProducts" method:"post" summary:"Update supplier product stock"`
	ID            uint64  `json:"id" v:"required"`
	StockQuantity float64 `json:"stock_quantity"`
}

type UpdateStockRes struct {
	g.Meta `mime:"application/json"`
	ProductInfo
}

type UpdateStatusReq struct {
	g.Meta `path:"/api/v1/supplier/products/status" tags:"SupplierProducts" method:"post" summary:"Update supplier product status"`
	ID     uint64 `json:"id" v:"required"`
	Status uint8  `json:"status"`
}

type UpdateStatusRes struct {
	g.Meta `mime:"application/json"`
	ProductInfo
}

type UploadImageReq struct {
	g.Meta `path:"/api/v1/supplier/products/images/upload" tags:"SupplierProducts" method:"post" summary:"Upload supplier product image"`
	File   *ghttp.UploadFile `json:"file" type:"file" v:"required" dc:"Product image"`
}

type UploadImageRes struct {
	g.Meta       `mime:"application/json"`
	Endpoint     string `json:"endpoint"`
	Bucket       string `json:"bucket"`
	Path         string `json:"path"`
	OriginalName string `json:"original_name"`
	ContentType  string `json:"content_type"`
	Size         int64  `json:"size"`
}
