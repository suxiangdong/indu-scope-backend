// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// CompanyVerifications is the golang structure for table company_verifications.
type CompanyVerifications struct {
	UserId                  uint64    `json:"userId"                  orm:"user_id"                    description:""` //
	CompanyName             string    `json:"companyName"             orm:"company_name"               description:""` //
	UnifiedSocialCreditCode string    `json:"unifiedSocialCreditCode" orm:"unified_social_credit_code" description:""` //
	BusinessLicenseFileId   string    `json:"businessLicenseFileId"   orm:"business_license_file_id"   description:""` //
	RegisteredAddress       string    `json:"registeredAddress"       orm:"registered_address"         description:""` //
	VerifiedAt              time.Time `json:"verifiedAt"              orm:"verified_at"                description:""` //
	CreatedAt               time.Time `json:"createdAt"               orm:"created_at"                 description:""` //
	UpdatedAt               time.Time `json:"updatedAt"               orm:"updated_at"                 description:""` //
}
