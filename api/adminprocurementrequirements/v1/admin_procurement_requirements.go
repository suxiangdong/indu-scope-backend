package v1

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type ListReq struct {
	g.Meta       `path:"/api/v1/admin/procurement/requirements" tags:"AdminProcurementRequirements" method:"get" summary:"List procurement requirements"`
	Keyword      string  `json:"keyword" in:"query"`
	CategoryID   *uint64 `json:"category_id" in:"query"`
	BuyerUserID  *uint64 `json:"buyer_user_id" in:"query"`
	ReviewStatus *uint8  `json:"review_status" in:"query"`
	Status       *uint8  `json:"status" in:"query"`
	Page         uint    `json:"page" in:"query"`
	PageSize     uint    `json:"page_size" in:"query"`
}

type ListRes struct {
	g.Meta   `mime:"application/json"`
	Items    []RequirementInfo `json:"items"`
	Total    uint64            `json:"total"`
	Page     uint              `json:"page"`
	PageSize uint              `json:"page_size"`
}

type DetailReq struct {
	g.Meta `path:"/api/v1/admin/procurement/requirements/detail" tags:"AdminProcurementRequirements" method:"get" summary:"Get procurement requirement detail"`
	ID     uint64 `json:"id" in:"query" v:"required"`
}

type ReviewReq struct {
	g.Meta       `path:"/api/v1/admin/procurement/requirements/review" tags:"AdminProcurementRequirements" method:"post" summary:"Review procurement requirement"`
	ID           uint64 `json:"id" v:"required"`
	ReviewStatus uint8  `json:"review_status"`
	Reason       string `json:"reason"`
}

type ReviewRes struct {
	g.Meta `mime:"application/json"`
	RequirementInfo
}

type RequirementAttributeValueInfo struct {
	AttributeID   uint64   `json:"attribute_id"`
	AttributeName string   `json:"attribute_name,omitempty"`
	Value         string   `json:"value"`
	OptionID      uint64   `json:"option_id,omitempty"`
	OptionIDs     []uint64 `json:"option_ids,omitempty"`
}

type RequirementFileInfo struct {
	ID           uint64    `json:"id"`
	Endpoint     string    `json:"endpoint"`
	Bucket       string    `json:"bucket"`
	Path         string    `json:"path"`
	URL          string    `json:"url,omitempty"`
	OriginalName string    `json:"original_name"`
	ContentType  string    `json:"content_type"`
	Size         int64     `json:"size"`
	SortOrder    uint      `json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
}

type RequirementResponseInfo struct {
	ID                uint64    `json:"id"`
	RequirementID     uint64    `json:"requirement_id"`
	SupplierUserID    uint64    `json:"supplier_user_id"`
	SupplierCompany   string    `json:"supplier_company,omitempty"`
	Message           string    `json:"message,omitempty"`
	AvailableQuantity *float64  `json:"available_quantity,omitempty"`
	LeadTimeDays      *uint     `json:"lead_time_days,omitempty"`
	Status            uint8     `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type RequirementInfo struct {
	ID                    uint64                          `json:"id"`
	BuyerUserID           uint64                          `json:"buyer_user_id"`
	BuyerCompany          string                          `json:"buyer_company,omitempty"`
	CategoryID            uint64                          `json:"category_id"`
	CategoryName          string                          `json:"category_name,omitempty"`
	CategoryNameSnapshot  string                          `json:"category_name_snapshot"`
	Name                  string                          `json:"name"`
	Quantity              float64                         `json:"quantity"`
	StockUnitID           uint64                          `json:"stock_unit_id"`
	StockUnitName         string                          `json:"stock_unit_name,omitempty"`
	StockUnitNameSnapshot string                          `json:"stock_unit_name_snapshot"`
	Description           string                          `json:"description,omitempty"`
	DeliveryAddress       string                          `json:"delivery_address"`
	ExpectedDeliveryDate  *time.Time                      `json:"expected_delivery_date,omitempty"`
	ContactName           string                          `json:"contact_name"`
	ContactPhone          string                          `json:"contact_phone"`
	ReviewStatus          uint8                           `json:"review_status"`
	ReviewReason          string                          `json:"review_reason,omitempty"`
	Status                uint8                           `json:"status"`
	PublishedAt           *time.Time                      `json:"published_at,omitempty"`
	ExpiresAt             *time.Time                      `json:"expires_at,omitempty"`
	ResponseCount         uint                            `json:"response_count"`
	CreatedAt             time.Time                       `json:"created_at"`
	UpdatedAt             time.Time                       `json:"updated_at"`
	Attributes            []RequirementAttributeValueInfo `json:"attributes,omitempty"`
	Files                 []RequirementFileInfo           `json:"files,omitempty"`
}

type DetailRes struct {
	g.Meta `mime:"application/json"`
	RequirementInfo
	Responses []RequirementResponseInfo `json:"responses"`
}
