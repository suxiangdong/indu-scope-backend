package v1

import "github.com/gogf/gf/v2/frame/g"

type RecordClickReq struct {
	g.Meta      `path:"/api/v1/analytics/click" tags:"Analytics" method:"post" summary:"Record a product or requirement click"`
	ContentType string `json:"content_type" v:"required|in:product,requirement"`
	ContentID   uint64 `json:"content_id" v:"required"`
}

type RecordClickRes struct {
	g.Meta  `mime:"application/json"`
	Success bool `json:"success"`
}
