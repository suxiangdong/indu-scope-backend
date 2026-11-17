package favorite

import "time"

type ProductItem struct {
	ID              uint64
	ProductID       uint64
	ProductName     string
	SupplierUserID  uint64
	SupplierCompany string
	CategoryID      uint64
	CategoryName    string
	StockQuantity   float64
	StockUnitName   string
	CoverURL        string
	ImageCount      uint
	Status          uint8
	FavoritedAt     time.Time
}

type ListResult struct {
	Items    []ProductItem
	Total    uint64
	Page     uint
	PageSize uint
}

type State struct {
	ProductID  uint64
	Favorited  bool
	FavoriteID uint64
}
