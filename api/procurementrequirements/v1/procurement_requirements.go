package v1

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	userprofilesv1 "github.com/suxiangdong/indu-scope-backend/api/userprofiles/v1"
)

type CategoryInfo struct {
	ID        uint64 `json:"id"`
	ParentID  uint64 `json:"parent_id"`
	Name      string `json:"name"`
	SortOrder uint   `json:"sort_order"`
}

type AttributeOptionInfo struct {
	ID   uint64 `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type AttributeInfo struct {
	ID               uint64                `json:"id"`
	Code             string                `json:"code"`
	Name             string                `json:"name"`
	ValueType        uint                  `json:"value_type"`
	InputType        string                `json:"input_type"`
	Unit             string                `json:"unit,omitempty"`
	Placeholder      string                `json:"placeholder,omitempty"`
	HelpText         string                `json:"help_text,omitempty"`
	IsRequired       uint8                 `json:"is_required"`
	SortOrder        uint                  `json:"sort_order"`
	GroupDisplayName string                `json:"group_display_name,omitempty"`
	Options          []AttributeOptionInfo `json:"options,omitempty"`
}

type StockUnitInfo struct {
	ID        uint64 `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	SortOrder uint   `json:"sort_order"`
}

type FormOptionsReq struct {
	g.Meta     `path:"/api/v1/procurement/requirements/form-options" tags:"ProcurementRequirements" method:"get" summary:"Get procurement requirement form options"`
	CategoryID *uint64 `json:"category_id" in:"query"`
}

type FormOptionsRes struct {
	g.Meta     `mime:"application/json"`
	Categories []CategoryInfo  `json:"categories"`
	Attributes []AttributeInfo `json:"attributes"`
	StockUnits []StockUnitInfo `json:"stock_units"`
}

type ListReq struct {
	g.Meta     `path:"/api/v1/procurement/requirements" tags:"ProcurementRequirements" method:"get" summary:"List procurement requirements"`
	Scope      string  `json:"scope" in:"query"`
	Keyword    string  `json:"keyword" in:"query"`
	CategoryID *uint64 `json:"category_id" in:"query"`
	Page       uint    `json:"page" in:"query"`
	PageSize   uint    `json:"page_size" in:"query"`
}

type ListRes struct {
	g.Meta   `mime:"application/json"`
	Items    []RequirementInfo `json:"items"`
	Total    uint64            `json:"total"`
	Page     uint              `json:"page"`
	PageSize uint              `json:"page_size"`
}

type DetailReq struct {
	g.Meta `path:"/api/v1/procurement/requirements/detail" tags:"ProcurementRequirements" method:"get" summary:"Get procurement requirement detail"`
	ID     uint64 `json:"id" in:"query" v:"required"`
}

type AttributeValueReq struct {
	AttributeID uint64   `json:"attribute_id"`
	Value       string   `json:"value"`
	OptionID    uint64   `json:"option_id,omitempty"`
	OptionIDs   []uint64 `json:"option_ids,omitempty"`
}

type CreateReq struct {
	g.Meta               `path:"/api/v1/procurement/requirements" tags:"ProcurementRequirements" method:"post" summary:"Create procurement requirement"`
	CategoryID           uint64              `json:"category_id" v:"required"`
	Name                 string              `json:"name" v:"required"`
	Quantity             float64             `json:"quantity"`
	StockUnitID          uint64              `json:"stock_unit_id" v:"required"`
	Description          string              `json:"description"`
	DeliveryAddress      string              `json:"delivery_address" v:"required"`
	ExpectedDeliveryDate string              `json:"expected_delivery_date"`
	ContactName          string              `json:"contact_name" v:"required"`
	ContactPhone         string              `json:"contact_phone" v:"required"`
	FilePaths            []string            `json:"file_paths"`
	Attributes           []AttributeValueReq `json:"attributes"`
}

type UpdateStatusReq struct {
	g.Meta `path:"/api/v1/procurement/requirements/status" tags:"ProcurementRequirements" method:"post" summary:"Update procurement requirement status"`
	ID     uint64 `json:"id" v:"required"`
	Status uint8  `json:"status"`
}

type UploadFileReq struct {
	g.Meta `path:"/api/v1/procurement/requirements/files/upload" tags:"ProcurementRequirements" method:"post" summary:"Upload procurement requirement attachment"`
	File   *ghttp.UploadFile `json:"file" type:"file" v:"required" dc:"Requirement attachment"`
}

type CreateResponseReq struct {
	g.Meta            `path:"/api/v1/procurement/requirements/responses" tags:"ProcurementRequirements" method:"post" summary:"Create procurement requirement response"`
	RequirementID     uint64   `json:"requirement_id" v:"required"`
	Message           string   `json:"message"`
	AvailableQuantity *float64 `json:"available_quantity"`
	LeadTimeDays      *uint    `json:"lead_time_days"`
}

type UpdateResponseStatusReq struct {
	g.Meta `path:"/api/v1/procurement/requirements/responses/status" tags:"ProcurementRequirements" method:"post" summary:"Update procurement response status"`
	ID     uint64 `json:"id" v:"required"`
	Status uint8  `json:"status"`
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
	ID                uint64                         `json:"id"`
	RequirementID     uint64                         `json:"requirement_id"`
	SupplierUserID    uint64                         `json:"supplier_user_id"`
	SupplierCompany   string                         `json:"supplier_company,omitempty"`
	CompanyInfo       *userprofilesv1.CompanySummary `json:"company_info,omitempty"`
	CompanyLocked     bool                           `json:"company_locked"`
	Message           string                         `json:"message,omitempty"`
	AvailableQuantity *float64                       `json:"available_quantity,omitempty"`
	LeadTimeDays      *uint                          `json:"lead_time_days,omitempty"`
	Status            uint8                          `json:"status"`
	CreatedAt         time.Time                      `json:"created_at"`
	UpdatedAt         time.Time                      `json:"updated_at"`
}

type ContactInfo struct {
	Name     string `json:"name"`
	Position string `json:"position,omitempty"`
	Phone    string `json:"phone"`
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
	g.Meta        `mime:"application/json"`
	Requirement   RequirementInfo                `json:"requirement"`
	CompanyInfo   *userprofilesv1.CompanySummary `json:"company_info,omitempty"`
	CompanyLocked bool                           `json:"company_locked"`
	Responses     []RequirementResponseInfo      `json:"responses"`
	BuyerContacts []ContactInfo                  `json:"buyer_contacts"`
}

type CreateRes struct {
	g.Meta `mime:"application/json"`
	RequirementInfo
}

type UpdateStatusRes struct {
	g.Meta `mime:"application/json"`
	RequirementInfo
}

type UploadFileRes struct {
	g.Meta       `mime:"application/json"`
	Endpoint     string `json:"endpoint"`
	Bucket       string `json:"bucket"`
	Path         string `json:"path"`
	OriginalName string `json:"original_name"`
	ContentType  string `json:"content_type"`
	Size         int64  `json:"size"`
}

type CreateResponseRes struct {
	g.Meta `mime:"application/json"`
	RequirementResponseInfo
}

type UpdateResponseStatusRes struct {
	g.Meta `mime:"application/json"`
	RequirementResponseInfo
}
