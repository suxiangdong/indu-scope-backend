// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// PersonalVerifications is the golang structure of table personal_verifications for DAO operations like Where/Data.
type PersonalVerifications struct {
	g.Meta            `orm:"table:personal_verifications, do:true"`
	Id                any    //
	UserId            any    //
	RealName          any    //
	IdNumber          any    //
	IdNumberEncrypted []byte //
	StorePlatformId   any    //
	StoreUrl          any    //
	IdFrontBucket     any    //
	IdFrontEndpoint   any    //
	IdFrontPath       any    //
	IdBackBucket      any    //
	IdBackEndpoint    any    //
	IdBackPath        any    //
	VerifiedAt        any    //
	ReviewStatus      any    //
	ReviewReason      any    //
	ReviewedBy        any    //
	ReviewedAt        any    //
	CreatedAt         any    //
	UpdatedAt         any    //
}
