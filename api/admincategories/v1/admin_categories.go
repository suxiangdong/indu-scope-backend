package v1

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type CategoryInfo struct {
	ID        int64     `json:"id" dc:"Category ID"`
	ParentID  *int64    `json:"parent_id" dc:"Parent category ID, 0 for top-level categories"`
	Name      string    `json:"name" dc:"Category name"`
	SortOrder uint      `json:"sort_order" dc:"Sort order"`
	Status    uint8     `json:"status" dc:"0 disabled, 1 active"`
	CreatedAt time.Time `json:"created_at" dc:"Created time"`
	UpdatedAt time.Time `json:"updated_at" dc:"Updated time"`
}

type ListReq struct {
	g.Meta `path:"/api/v1/admin/categories" tags:"AdminCategories" method:"get" summary:"List categories"`
}

type ListRes struct {
	g.Meta `mime:"application/json"`
	Items  []CategoryInfo `json:"items"`
}

type CreateReq struct {
	g.Meta    `path:"/api/v1/admin/categories/create" tags:"AdminCategories" method:"post" summary:"Create category"`
	ParentID  *uint64 `json:"parent_id" dc:"Parent category ID, 0 for top-level categories"`
	Name      string  `json:"name" v:"required" dc:"Category name"`
	SortOrder uint    `json:"sort_order" dc:"Sort order"`
	Status    *uint   `json:"status" dc:"0 disabled, 1 active"`
}

type CreateRes struct {
	g.Meta `mime:"application/json"`
	CategoryInfo
}

type UpdateReq struct {
	g.Meta    `path:"/api/v1/admin/categories/update" tags:"AdminCategories" method:"post" summary:"Update category"`
	ID        uint64  `json:"id" v:"required" dc:"Category ID"`
	ParentID  *uint64 `json:"parent_id" dc:"Parent category ID, 0 for top-level categories"`
	Name      string  `json:"name" v:"required" dc:"Category name"`
	SortOrder uint    `json:"sort_order" dc:"Sort order"`
	Status    *uint   `json:"status" dc:"0 disabled, 1 active"`
}

type UpdateRes struct {
	g.Meta `mime:"application/json"`
	CategoryInfo
}

type DeleteReq struct {
	g.Meta `path:"/api/v1/admin/categories/delete" tags:"AdminCategories" method:"post" summary:"Delete category"`
	ID     uint64 `json:"id" v:"required" dc:"Category ID"`
}

type DeleteRes struct {
	g.Meta  `mime:"application/json"`
	Success bool `json:"success"`
}
