// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// CompanyVerifications is the golang structure for table company_verifications.
type CompanyVerifications struct {
	Id                      uint64    `json:"id"                      orm:"id"                       description:""`   //
	UserId                  uint64    `json:"userId"                  orm:"user_id"                    description:""` //
	CompanyName             string    `json:"companyName"             orm:"company_name"               description:""` //
	UnifiedSocialCreditCode string    `json:"unifiedSocialCreditCode" orm:"unified_social_credit_code" description:""` //
	BusinessLicenseBucket   string    `json:"businessLicenseBucket"   orm:"business_license_bucket"   description:""`  //
	BusinessLicenseEndpoint string    `json:"businessLicenseEndpoint" orm:"business_license_endpoint" description:""`  //
	BusinessLicensePath     string    `json:"businessLicensePath"     orm:"business_license_path"     description:""`  //
	RegisteredAddress       string    `json:"registeredAddress"       orm:"registered_address"         description:""` //
	VerifiedAt              time.Time `json:"verifiedAt"              orm:"verified_at"                description:""` //
	ReviewStatus            uint      `json:"reviewStatus"            orm:"review_status"             description:""`  //
	ReviewReason            string    `json:"reviewReason"            orm:"review_reason"             description:""`  //
	ReviewedBy              uint64    `json:"reviewedBy"              orm:"reviewed_by"               description:""`  //
	ReviewedAt              time.Time `json:"reviewedAt"              orm:"reviewed_at"               description:""`  //
	CreatedAt               time.Time `json:"createdAt"               orm:"created_at"                 description:""` //
	UpdatedAt               time.Time `json:"updatedAt"               orm:"updated_at"                 description:""` //
}
