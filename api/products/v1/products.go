package v1

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type ListReq struct {
	g.Meta     `path:"/api/v1/products" tags:"Products" method:"get" summary:"List published products for buyers"`
	Keyword    string `json:"keyword" in:"query" dc:"Product or supplier keyword"`
	CategoryID uint64 `json:"category_id" in:"query" dc:"Category ID"`
	Page       uint   `json:"page" in:"query"`
	PageSize   uint   `json:"page_size" in:"query"`
}

type DetailReq struct {
	g.Meta `path:"/api/v1/products/detail" tags:"Products" method:"get" summary:"Get a published product detail"`
	ID     uint64 `json:"id" in:"query" v:"required"`
}

type ProductInfo struct {
	ID              uint64    `json:"id"`
	SupplierUserID  uint64    `json:"supplier_user_id"`
	SupplierCompany string    `json:"supplier_company,omitempty"`
	CategoryID      uint64    `json:"category_id"`
	CategoryName    string    `json:"category_name,omitempty"`
	Name            string    `json:"name"`
	StockQuantity   float64   `json:"stock_quantity"`
	StockUnitID     uint64    `json:"stock_unit_id"`
	StockUnitCode   string    `json:"stock_unit_code"`
	StockUnitName   string    `json:"stock_unit_name"`
	CoverURL        string    `json:"cover_url,omitempty"`
	ImageCount      uint      `json:"image_count"`
	Status          uint8     `json:"status"`
	Favorited       bool      `json:"favorited"`
	CreatedAt       time.Time `json:"created_at"`
}

type ProductImageInfo struct {
	ID           uint64 `json:"id"`
	URL          string `json:"url,omitempty"`
	Path         string `json:"path,omitempty"`
	OriginalName string `json:"original_name,omitempty"`
	ContentType  string `json:"content_type,omitempty"`
	Size         int64  `json:"size,omitempty"`
	SortOrder    uint   `json:"sort_order"`
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
	Images           []ProductImageInfo   `json:"images"`
	Attributes       []AttributeValueInfo `json:"attributes"`
	SupplierContacts []ContactInfo        `json:"supplier_contacts"`
	ContactLocked    bool                 `json:"contact_locked"`
}

type ListRes struct {
	g.Meta   `mime:"application/json"`
	Items    []ProductInfo `json:"items"`
	Total    uint64        `json:"total"`
	Page     uint          `json:"page"`
	PageSize uint          `json:"page_size"`
}
