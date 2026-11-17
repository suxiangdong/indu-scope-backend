package hotcategory

import "time"

type Item struct {
	ID         uint64
	CategoryID uint64
	Name       string
	SortOrder  uint
	Status     uint8
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type ListFilter struct {
	Keyword string
	Status  *uint8
}

type SortItem struct {
	ID        uint64
	SortOrder uint
}

type ListResult struct {
	Items []Item
}
