package v1

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type StorePlatformInfo struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	SortOrder uint      `json:"sort_order"`
	Status    uint8     `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListReq struct {
	g.Meta `path:"/api/v1/admin/store-platforms" tags:"AdminStorePlatform" method:"get" summary:"List store platforms"`
}

type ListRes struct {
	g.Meta `mime:"application/json"`
	Items  []StorePlatformInfo `json:"items"`
}

type CreateReq struct {
	g.Meta    `path:"/api/v1/admin/store-platforms/create" tags:"AdminStorePlatform" method:"post" summary:"Create store platform"`
	Name      string `json:"name" v:"required"`
	SortOrder uint   `json:"sort_order"`
	Status    *uint  `json:"status"`
}

type CreateRes struct {
	g.Meta `mime:"application/json"`
	StorePlatformInfo
}

type UpdateReq struct {
	g.Meta    `path:"/api/v1/admin/store-platforms/update" tags:"AdminStorePlatform" method:"post" summary:"Update store platform"`
	ID        uint64 `json:"id" v:"required"`
	Name      string `json:"name" v:"required"`
	SortOrder uint   `json:"sort_order"`
	Status    *uint  `json:"status"`
}

type UpdateRes struct {
	g.Meta `mime:"application/json"`
	StorePlatformInfo
}

type DeleteReq struct {
	g.Meta `path:"/api/v1/admin/store-platforms/delete" tags:"AdminStorePlatform" method:"post" summary:"Delete store platform"`
	ID     uint64 `json:"id" v:"required"`
}

type DeleteRes struct {
	g.Meta  `mime:"application/json"`
	Success bool `json:"success"`
}
