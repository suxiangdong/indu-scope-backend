package supplierproduct

import (
	"io"
	"time"

	userprofileModel "github.com/suxiangdong/indu-scope-backend/internal/model/userprofile"
)

type CategoryItem struct {
	ID        uint64
	ParentID  uint64
	Name      string
	SortOrder uint
}

type AttributeOptionItem struct {
	ID   uint64
	Code string
	Name string
}

type AttributeItem struct {
	ID               uint64
	Code             string
	Name             string
	ValueType        uint
	InputType        string
	Unit             string
	Placeholder      string
	HelpText         string
	IsRequired       uint8
	SortOrder        uint
	GroupDisplayName string
	Options          []AttributeOptionItem
}

type FormOptions struct {
	Categories []CategoryItem
	Attributes []AttributeItem
	StockUnits []StockUnitItem
}

type StockUnitItem struct {
	ID        uint64
	Code      string
	Name      string
	SortOrder uint
}

type AttributeValueInput struct {
	AttributeID uint64
	Value       string
	OptionID    uint64
	OptionIDs   []uint64
}

type CreateInput struct {
	SupplierUserID uint64
	CategoryID     uint64
	Name           string
	StockQuantity  float64
	StockUnitID    uint64
	ImagePaths     []string
	Attributes     []AttributeValueInput
}

type UpdateStockInput struct {
	SupplierUserID uint64
	ProductID      uint64
	StockQuantity  float64
}

type ProductImageItem struct {
	ID           uint64
	Endpoint     string
	Bucket       string
	Path         string
	URL          string
	OriginalName string
	ContentType  string
	Size         int64
	SortOrder    uint
	CreatedAt    time.Time
}

type AttributeValueItem struct {
	AttributeID uint64
	Code        string
	Name        string
	Unit        string
	Value       string
}

type ProductImageUploadInput struct {
	Reader      io.Reader
	Filename    string
	ContentType string
	Size        int64
}

type ProductImageUploadResult struct {
	Endpoint     string
	Bucket       string
	Path         string
	OriginalName string
	ContentType  string
	Size         int64
}

type ProductItem struct {
	ID             uint64
	SupplierUserID uint64
	CategoryID     uint64
	Name           string
	StockQuantity  float64
	StockUnitID    uint64
	StockUnitCode  string
	StockUnitName  string
	CategoryName   string
	CoverURL       string
	ImageCount     uint
	CreatedAt      time.Time
	Images         []ProductImageItem
	Attributes     []AttributeValueItem
	Contacts       []userprofileModel.Contact
	Status         uint
}

type ListResult struct {
	Items    []ProductItem
	Total    uint64
	Page     uint
	PageSize uint
}
