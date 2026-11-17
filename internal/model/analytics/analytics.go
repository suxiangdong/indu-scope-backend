package analytics

import "time"

type ContentType string

const (
	ContentTypeProduct     ContentType = "product"
	ContentTypeRequirement ContentType = "requirement"
)

type DailyStat struct {
	Date              string
	ProductClicks     uint64
	RequirementClicks uint64
	TotalClicks       uint64
}

type TopContent struct {
	ID     uint64
	Name   string
	Clicks uint64
}

type Overview struct {
	Days              uint
	StartDate         time.Time
	TotalClicks       uint64
	ProductClicks     uint64
	RequirementClicks uint64
	Daily             []DailyStat
	TopProducts       []TopContent
	TopRequirements   []TopContent
}
