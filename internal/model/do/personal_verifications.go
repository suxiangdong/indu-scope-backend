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
	UserId            any    //
	RealName          any    //
	IdNumberEncrypted []byte //
	IdFrontFileId     any    //
	IdBackFileId      any    //
	VerifiedAt        any    //
	CreatedAt         any    //
	UpdatedAt         any    //
}
