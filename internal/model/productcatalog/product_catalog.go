package productcatalog

import (
	"time"

	userprofileModel "github.com/suxiangdong/indu-scope-backend/internal/model/userprofile"
)

type ListFilter struct {
	Keyword    string
	CategoryID uint64
	Page       uint
	PageSize   uint
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
	StockUnitCode   string
	StockUnitName   string
	CoverURL        string
	ImageCount      uint
	Status          uint8
	Favorited       bool
	CreatedAt       time.Time
}

type ProductImage struct {
	ID           uint64
	URL          string
	Path         string
	OriginalName string
	ContentType  string
	Size         int64
	SortOrder    uint
}

type AttributeValue struct {
	AttributeID uint64
	Code        string
	Name        string
	Unit        string
	Value       string
}

type ProductDetail struct {
	ProductItem
	Images           []ProductImage
	Attributes       []AttributeValue
	SupplierContacts []userprofileModel.Contact
}

type ListResult struct {
	Items    []ProductItem
	Total    uint64
	Page     uint
	PageSize uint
}
