// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// CompanyProfiles is the golang structure of table company_profiles for DAO operations like Where/Data.
type CompanyProfiles struct {
	g.Meta           `orm:"table:company_profiles, do:true"`
	Id               any //
	UserId           any //
	IntroductionHtml any //
	Offices          any //
	Factories        any //
	CreatedAt        any //
	UpdatedAt        any //
}
