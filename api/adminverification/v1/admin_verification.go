package v1

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type VerificationListItem struct {
	ID           uint64     `json:"id"`
	UserID       uint64     `json:"user_id"`
	Type         string     `json:"type"`
	Phone        string     `json:"phone"`
	UserType     uint8      `json:"user_type"`
	DisplayName  string     `json:"display_name"`
	ReviewStatus uint8      `json:"review_status"`
	ReviewReason string     `json:"review_reason,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	ReviewedAt   *time.Time `json:"reviewed_at,omitempty"`
}

type ListReq struct {
	g.Meta       `path:"/api/v1/admin/verifications" tags:"AdminVerification" method:"get" summary:"List verification submissions"`
	Type         string `json:"type" in:"query" dc:"personal or company"`
	ReviewStatus *uint8 `json:"review_status" in:"query" dc:"0 pending, 1 approved, 2 rejected"`
}

type ListRes struct {
	g.Meta `mime:"application/json"`
	Items  []VerificationListItem `json:"items"`
}

type ContactInfo struct {
	Name     string `json:"name"`
	Position string `json:"position,omitempty"`
	Phone    string `json:"phone"`
}

type PersonalDetail struct {
	RealName          string     `json:"real_name"`
	IDNumberMasked    string     `json:"id_number_masked,omitempty"`
	StorePlatformID   uint64     `json:"store_platform_id,omitempty"`
	StorePlatformName string     `json:"store_platform_name,omitempty"`
	StoreURL          string     `json:"store_url,omitempty"`
	IDFrontEndpoint   string     `json:"id_front_endpoint,omitempty"`
	IDFrontBucket     string     `json:"id_front_bucket,omitempty"`
	IDFrontPath       string     `json:"id_front_path,omitempty"`
	IDBackEndpoint    string     `json:"id_back_endpoint,omitempty"`
	IDBackBucket      string     `json:"id_back_bucket,omitempty"`
	IDBackPath        string     `json:"id_back_path,omitempty"`
	VerifiedAt        *time.Time `json:"verified_at,omitempty"`
}

type CompanyDetail struct {
	CompanyName                string        `json:"company_name"`
	UnifiedSocialCreditCode    string        `json:"unified_social_credit_code"`
	BusinessLicenseEndpoint    string        `json:"business_license_endpoint,omitempty"`
	BusinessLicenseBucket      string        `json:"business_license_bucket,omitempty"`
	BusinessLicensePath        string        `json:"business_license_path,omitempty"`
	BusinessLicenseURL         string        `json:"business_license_url,omitempty"`
	BusinessLicenseFileName    string        `json:"business_license_file_name,omitempty"`
	BusinessLicenseFileSize    int64         `json:"business_license_file_size,omitempty"`
	BusinessLicenseContentType string        `json:"business_license_content_type,omitempty"`
	RegisteredAddress          string        `json:"registered_address"`
	Introduction               string        `json:"introduction,omitempty"`
	Offices                    []string      `json:"offices"`
	Factories                  []string      `json:"factories"`
	Contacts                   []ContactInfo `json:"contacts"`
	VerifiedAt                 *time.Time    `json:"verified_at,omitempty"`
}

type DetailInfo struct {
	VerificationListItem
	UserStatus uint8           `json:"user_status"`
	Personal   *PersonalDetail `json:"personal,omitempty"`
	Company    *CompanyDetail  `json:"company,omitempty"`
}

type DetailReq struct {
	g.Meta `path:"/api/v1/admin/verifications/detail" tags:"AdminVerification" method:"get" summary:"Get verification detail"`
	ID     uint64 `json:"id" in:"query" v:"required"`
	Type   string `json:"type" in:"query" v:"required"`
}

type DetailRes struct {
	g.Meta `mime:"application/json"`
	DetailInfo
}

type ReviewReq struct {
	g.Meta `path:"/api/v1/admin/verifications/review" tags:"AdminVerification" method:"post" summary:"Approve or reject verification"`
	ID     uint64 `json:"id" v:"required"`
	Type   string `json:"type" v:"required"`
	Action string `json:"action" v:"required" dc:"approve or reject"`
	Reason string `json:"reason" dc:"Rejection reason"`
}

type ReviewRes struct {
	g.Meta `mime:"application/json"`
	DetailInfo
}
