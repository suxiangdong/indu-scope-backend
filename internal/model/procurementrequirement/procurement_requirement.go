package procurementrequirement

import (
	"io"
	"time"

	supplierProductModel "github.com/suxiangdong/indu-scope-backend/internal/model/supplierproduct"
	userprofileModel "github.com/suxiangdong/indu-scope-backend/internal/model/userprofile"
)

type AttributeValueInput = supplierProductModel.AttributeValueInput

type CreateInput struct {
	BuyerUserID          uint64
	CategoryID           uint64
	Name                 string
	Quantity             float64
	StockUnitID          uint64
	Description          string
	DeliveryAddress      string
	ExpectedDeliveryDate string
	ContactName          string
	ContactPhone         string
	FilePaths            []string
	Attributes           []AttributeValueInput
}

type ListFilter struct {
	UserID     uint64
	Scope      string
	Keyword    string
	CategoryID *uint64
	Page       uint
	PageSize   uint
}

type AdminListFilter struct {
	Keyword      string
	CategoryID   *uint64
	BuyerUserID  *uint64
	ReviewStatus *uint8
	Status       *uint8
	Page         uint
	PageSize     uint
}

type RequirementItem struct {
	ID                    uint64
	BuyerUserID           uint64
	BuyerCompany          string
	CategoryID            uint64
	CategoryName          string
	CategoryNameSnapshot  string
	Name                  string
	Quantity              float64
	StockUnitID           uint64
	StockUnitName         string
	StockUnitNameSnapshot string
	Description           string
	DeliveryAddress       string
	ExpectedDeliveryDate  *time.Time
	ContactName           string
	ContactPhone          string
	ReviewStatus          uint8
	ReviewReason          string
	Status                uint8
	PublishedAt           *time.Time
	ExpiresAt             *time.Time
	ResponseCount         uint
	CreatedAt             time.Time
	UpdatedAt             time.Time
	Attributes            []AttributeValueItem
	Files                 []FileItem
}

type AttributeValueItem struct {
	AttributeID   uint64
	AttributeName string
	Value         string
	OptionID      uint64
	OptionIDs     []uint64
}

type FileItem struct {
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

type ResponseItem struct {
	ID                uint64
	RequirementID     uint64
	SupplierUserID    uint64
	SupplierCompany   string
	Message           string
	AvailableQuantity *float64
	LeadTimeDays      *uint
	Status            uint8
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Detail struct {
	Requirement   RequirementItem
	Responses     []ResponseItem
	BuyerContacts []userprofileModel.Contact
}

type ListResult struct {
	Items    []RequirementItem
	Total    uint64
	Page     uint
	PageSize uint
}

type ResponseInput struct {
	RequirementID     uint64
	SupplierUserID    uint64
	Message           string
	AvailableQuantity *float64
	LeadTimeDays      *uint
}

type FileUploadInput struct {
	Reader      io.Reader
	Filename    string
	ContentType string
	Size        int64
}

type FileUploadResult struct {
	Endpoint     string
	Bucket       string
	Path         string
	OriginalName string
	ContentType  string
	Size         int64
}

type FormOptions = supplierProductModel.FormOptions
