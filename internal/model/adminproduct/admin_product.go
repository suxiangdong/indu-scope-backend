package adminproduct

import "time"

type ListFilter struct {
	Keyword        string
	CategoryID     *uint64
	SupplierUserID *uint64
	Status         *uint8
	Page           uint
	PageSize       uint
}

type ProductImage struct {
	ID           uint64
	URL          string
	Path         string
	OriginalName string
	ContentType  string
	Size         int64
	SortOrder    uint
	CreatedAt    time.Time
}

type AttributeValue struct {
	AttributeID uint64
	Code        string
	Name        string
	Unit        string
	Value       string
}

type ProductItem struct {
	ID              uint64
	SupplierUserID  uint64
	SupplierCompany string
	CategoryID      uint64
	CategoryName    string
	Name            string
	StockQuantity   float64
	StockUnitID     uint64
	StockUnitName   string
	CoverURL        string
	ImageCount      uint
	Status          uint8
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Images          []ProductImage
	Attributes      []AttributeValue
}

type ListResult struct {
	Items    []ProductItem
	Total    uint64
	Page     uint
	PageSize uint
}
