package v1

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type ProductImageInfo struct {
	ID           uint64 `json:"id"`
	URL          string `json:"url,omitempty"`
	Path         string `json:"path"`
	OriginalName string `json:"original_name"`
	ContentType  string `json:"content_type"`
	Size         int64  `json:"size"`
	SortOrder    uint   `json:"sort_order"`
}

type AttributeValueInfo struct {
	AttributeID uint64 `json:"attribute_id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Unit        string `json:"unit,omitempty"`
	Value       string `json:"value"`
}

type ProductInfo struct {
	ID              uint64               `json:"id"`
	SupplierUserID  uint64               `json:"supplier_user_id"`
	SupplierCompany string               `json:"supplier_company,omitempty"`
	CategoryID      uint64               `json:"category_id"`
	CategoryName    string               `json:"category_name,omitempty"`
	Name            string               `json:"name"`
	StockQuantity   float64              `json:"stock_quantity"`
	StockUnitID     uint64               `json:"stock_unit_id"`
	StockUnitName   string               `json:"stock_unit_name"`
	CoverURL        string               `json:"cover_url,omitempty"`
	ImageCount      uint                 `json:"image_count"`
	Status          uint8                `json:"status"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
	Images          []ProductImageInfo   `json:"images,omitempty"`
	Attributes      []AttributeValueInfo `json:"attributes,omitempty"`
}

type ListReq struct {
	g.Meta         `path:"/api/v1/admin/products" tags:"AdminProducts" method:"get" summary:"List products"`
	Keyword        string  `json:"keyword" in:"query"`
	CategoryID     *uint64 `json:"category_id" in:"query"`
	SupplierUserID *uint64 `json:"supplier_user_id" in:"query"`
	Status         *uint8  `json:"status" in:"query"`
	Page           uint    `json:"page" in:"query"`
	PageSize       uint    `json:"page_size" in:"query"`
}

type ListRes struct {
	g.Meta   `mime:"application/json"`
	Items    []ProductInfo `json:"items"`
	Total    uint64        `json:"total"`
	Page     uint          `json:"page"`
	PageSize uint          `json:"page_size"`
}

type DetailReq struct {
	g.Meta `path:"/api/v1/admin/products/detail" tags:"AdminProducts" method:"get" summary:"Get product detail"`
	ID     uint64 `json:"id" in:"query" v:"required"`
}

type DetailRes struct {
	g.Meta `mime:"application/json"`
	ProductInfo
}

type UpdateStatusReq struct {
	g.Meta `path:"/api/v1/admin/products/status" tags:"AdminProducts" method:"post" summary:"Update product status"`
	ID     uint64 `json:"id" v:"required"`
	Status uint8  `json:"status"`
}

type UpdateStatusRes struct {
	g.Meta `mime:"application/json"`
	ProductInfo
}
