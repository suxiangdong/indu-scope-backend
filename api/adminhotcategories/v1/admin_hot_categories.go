package v1

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type HotCategoryInfo struct {
	ID         uint64    `json:"id"`
	CategoryID uint64    `json:"category_id"`
	Name       string    `json:"name"`
	SortOrder  uint      `json:"sort_order"`
	Status     uint8     `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ListReq struct {
	g.Meta  `path:"/api/v1/admin/hot-categories" tags:"AdminHotCategories" method:"get" summary:"List hot categories"`
	Keyword string `json:"keyword" in:"query"`
	Status  *uint8 `json:"status" in:"query"`
}

type ListRes struct {
	g.Meta `mime:"application/json"`
	Items  []HotCategoryInfo `json:"items"`
}

type CreateReq struct {
	g.Meta     `path:"/api/v1/admin/hot-categories/create" tags:"AdminHotCategories" method:"post" summary:"Create hot category"`
	CategoryID uint64 `json:"category_id" v:"required"`
	SortOrder  uint   `json:"sort_order"`
	Status     *uint8 `json:"status"`
}

type CreateRes struct {
	g.Meta `mime:"application/json"`
	HotCategoryInfo
}

type UpdateReq struct {
	g.Meta     `path:"/api/v1/admin/hot-categories/update" tags:"AdminHotCategories" method:"post" summary:"Update hot category"`
	ID         uint64 `json:"id" v:"required"`
	CategoryID uint64 `json:"category_id" v:"required"`
	SortOrder  uint   `json:"sort_order"`
	Status     *uint8 `json:"status"`
}

type UpdateRes struct {
	g.Meta `mime:"application/json"`
	HotCategoryInfo
}

type DeleteReq struct {
	g.Meta `path:"/api/v1/admin/hot-categories/delete" tags:"AdminHotCategories" method:"post" summary:"Delete hot category"`
	ID     uint64 `json:"id" v:"required"`
}

type DeleteRes struct {
	g.Meta  `mime:"application/json"`
	Success bool `json:"success"`
}

type StatusReq struct {
	g.Meta `path:"/api/v1/admin/hot-categories/status" tags:"AdminHotCategories" method:"post" summary:"Update hot category status"`
	ID     uint64 `json:"id" v:"required"`
	Status uint8  `json:"status"`
}

type StatusRes struct {
	g.Meta `mime:"application/json"`
	HotCategoryInfo
}

type SortItem struct {
	ID        uint64 `json:"id" v:"required"`
	SortOrder uint   `json:"sort_order"`
}

type SortReq struct {
	g.Meta `path:"/api/v1/admin/hot-categories/sort" tags:"AdminHotCategories" method:"post" summary:"Sort hot categories"`
	Items  []SortItem `json:"items" v:"required"`
}

type SortRes struct {
	g.Meta  `mime:"application/json"`
	Success bool              `json:"success"`
	Items   []HotCategoryInfo `json:"items,omitempty"`
}
