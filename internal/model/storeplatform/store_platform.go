package storeplatform

import "time"

type Item struct {
	ID        uint64
	Name      string
	SortOrder uint
	Status    uint8
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ListResult struct {
	Items []Item
}

type Input struct {
	Name      string
	SortOrder uint
	Status    uint
}
