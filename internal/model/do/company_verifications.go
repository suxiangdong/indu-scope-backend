// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// CompanyVerifications is the golang structure of table company_verifications for DAO operations like Where/Data.
type CompanyVerifications struct {
	g.Meta                  `orm:"table:company_verifications, do:true"`
	Id                      any //
	UserId                  any //
	CompanyName             any //
	UnifiedSocialCreditCode any //
	BusinessLicenseBucket   any //
	BusinessLicenseEndpoint any //
	BusinessLicensePath     any //
	RegisteredAddress       any //
	VerifiedAt              any //
	ReviewStatus            any //
	ReviewReason            any //
	ReviewedBy              any //
	ReviewedAt              any //
	CreatedAt               any //
	UpdatedAt               any //
}
