package category

import "time"

// CategoryItem is the category data exposed by management APIs.
type CategoryItem struct {
	ID uint64
	// ParentID is always non-nil in responses; 0 identifies a top-level category.
	ParentID  *uint64
	Name      string
	SortOrder uint
	Status    uint8
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CategoryListResult contains all categories in display order.
type CategoryListResult struct {
	Items []CategoryItem
}
