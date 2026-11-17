package v1

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type StockUnitInfo struct {
	ID        uint64    `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	SortOrder uint      `json:"sort_order"`
	Status    uint8     `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CategoryInfo struct {
	CategoryID   uint64    `json:"category_id"`
	CategoryName string    `json:"category_name"`
	SortOrder    uint      `json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ListReq struct {
	g.Meta `path:"/api/v1/admin/stock-units" tags:"AdminStockUnit" method:"get" summary:"List stock units"`
}

type ListRes struct {
	g.Meta `mime:"application/json"`
	Items  []StockUnitInfo `json:"items"`
}

type CreateReq struct {
	g.Meta    `path:"/api/v1/admin/stock-units/create" tags:"AdminStockUnit" method:"post" summary:"Create stock unit"`
	Code      string `json:"code" v:"required"`
	Name      string `json:"name" v:"required"`
	SortOrder uint   `json:"sort_order"`
	Status    *uint  `json:"status"`
}

type CreateRes struct {
	g.Meta `mime:"application/json"`
	StockUnitInfo
}

type UpdateReq struct {
	g.Meta    `path:"/api/v1/admin/stock-units/update" tags:"AdminStockUnit" method:"post" summary:"Update stock unit"`
	ID        uint64 `json:"id" v:"required"`
	Code      string `json:"code" v:"required"`
	Name      string `json:"name" v:"required"`
	SortOrder uint   `json:"sort_order"`
	Status    *uint  `json:"status"`
}

type UpdateRes struct {
	g.Meta `mime:"application/json"`
	StockUnitInfo
}

type DeleteReq struct {
	g.Meta `path:"/api/v1/admin/stock-units/delete" tags:"AdminStockUnit" method:"post" summary:"Delete stock unit"`
	ID     uint64 `json:"id" v:"required"`
}

type DeleteRes struct {
	g.Meta  `mime:"application/json"`
	Success bool `json:"success"`
}

type ListCategoriesReq struct {
	g.Meta      `path:"/api/v1/admin/stock-units/categories" tags:"AdminStockUnit" method:"get" summary:"List stock unit categories"`
	StockUnitID uint64 `json:"stock_unit_id" in:"query" v:"required"`
}

type ListCategoriesRes struct {
	g.Meta `mime:"application/json"`
	Items  []CategoryInfo `json:"items"`
}

type BindCategoryReq struct {
	g.Meta      `path:"/api/v1/admin/stock-units/categories/bind" tags:"AdminStockUnit" method:"post" summary:"Bind stock unit to category"`
	StockUnitID uint64 `json:"stock_unit_id" v:"required"`
	CategoryID  uint64 `json:"category_id" v:"required"`
	SortOrder   uint   `json:"sort_order"`
}

type BindCategoryRes struct {
	g.Meta `mime:"application/json"`
	CategoryInfo
}

type UnbindCategoryReq struct {
	g.Meta      `path:"/api/v1/admin/stock-units/categories/unbind" tags:"AdminStockUnit" method:"post" summary:"Unbind stock unit from category"`
	StockUnitID uint64 `json:"stock_unit_id" v:"required"`
	CategoryID  uint64 `json:"category_id" v:"required"`
}

type UnbindCategoryRes struct {
	g.Meta  `mime:"application/json"`
	Success bool `json:"success"`
}
