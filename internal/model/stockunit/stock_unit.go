package stockunit

import "time"

type Item struct {
	ID        uint64
	Code      string
	Name      string
	SortOrder uint
	Status    uint8
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Input struct {
	Code      string
	Name      string
	SortOrder uint
	Status    uint
}

type ListResult struct {
	Items []Item
}

type CategoryItem struct {
	CategoryID   uint64
	CategoryName string
	SortOrder    uint
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CategoryListResult struct {
	Items []CategoryItem
}
