package v1

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type OverviewReq struct {
	g.Meta `path:"/api/v1/admin/analytics/overview" tags:"AdminAnalytics" method:"get" summary:"Get content click overview"`
	Days   uint `json:"days" in:"query"`
}

type DailyStat struct {
	Date              string `json:"date"`
	ProductClicks     uint64 `json:"product_clicks"`
	RequirementClicks uint64 `json:"requirement_clicks"`
	TotalClicks       uint64 `json:"total_clicks"`
}

type TopContent struct {
	ID     uint64 `json:"id"`
	Name   string `json:"name"`
	Clicks uint64 `json:"clicks"`
}

type OverviewInfo struct {
	Days              uint         `json:"days"`
	StartDate         time.Time    `json:"start_date"`
	TotalClicks       uint64       `json:"total_clicks"`
	ProductClicks     uint64       `json:"product_clicks"`
	RequirementClicks uint64       `json:"requirement_clicks"`
	Daily             []DailyStat  `json:"daily"`
	TopProducts       []TopContent `json:"top_products"`
	TopRequirements   []TopContent `json:"top_requirements"`
}

type OverviewRes struct {
	g.Meta `mime:"application/json"`
	OverviewInfo
}
