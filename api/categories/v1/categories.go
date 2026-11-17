package v1

import "github.com/gogf/gf/v2/frame/g"

type HotReq struct {
	g.Meta `path:"/api/v1/categories/hot" tags:"Categories" method:"get" summary:"List hot categories"`
}

type HotCategoryInfo struct {
	ID         uint64 `json:"id"`
	CategoryID uint64 `json:"category_id"`
	Name       string `json:"name"`
	SortOrder  uint   `json:"sort_order"`
}

type HotRes struct {
	g.Meta `mime:"application/json"`
	Items  []HotCategoryInfo `json:"items"`
}
