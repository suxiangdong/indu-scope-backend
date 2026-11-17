package v1

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type FavoriteProductInfo struct {
	FavoriteID      uint64    `json:"favorite_id"`
	ProductID       uint64    `json:"product_id"`
	ProductName     string    `json:"product_name"`
	SupplierUserID  uint64    `json:"supplier_user_id"`
	SupplierCompany string    `json:"supplier_company,omitempty"`
	CategoryID      uint64    `json:"category_id"`
	CategoryName    string    `json:"category_name,omitempty"`
	StockQuantity   float64   `json:"stock_quantity"`
	StockUnitName   string    `json:"stock_unit_name"`
	CoverURL        string    `json:"cover_url,omitempty"`
	ImageCount      uint      `json:"image_count"`
	Status          uint8     `json:"status"`
	FavoritedAt     time.Time `json:"favorited_at"`
}

type FavoriteState struct {
	ProductID  uint64 `json:"product_id"`
	Favorited  bool   `json:"favorited"`
	FavoriteID uint64 `json:"favorite_id,omitempty"`
}

type ListReq struct {
	g.Meta   `path:"/api/v1/favorites" tags:"Favorites" method:"get" summary:"List favorite products"`
	Page     uint `json:"page" in:"query"`
	PageSize uint `json:"page_size" in:"query"`
}

type ListRes struct {
	g.Meta   `mime:"application/json"`
	Items    []FavoriteProductInfo `json:"items"`
	Total    uint64                `json:"total"`
	Page     uint                  `json:"page"`
	PageSize uint                  `json:"page_size"`
}

type ToggleReq struct {
	g.Meta    `path:"/api/v1/favorites/toggle" tags:"Favorites" method:"post" summary:"Toggle product favorite"`
	ProductID uint64 `json:"product_id" v:"required"`
}

type ToggleRes struct {
	g.Meta `mime:"application/json"`
	FavoriteState
}

type StatusReq struct {
	g.Meta    `path:"/api/v1/favorites/status" tags:"Favorites" method:"get" summary:"Get product favorite status"`
	ProductID uint64 `json:"product_id" in:"query" v:"required"`
}

type StatusRes struct {
	g.Meta `mime:"application/json"`
	FavoriteState
}
