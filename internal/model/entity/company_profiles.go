// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// CompanyProfiles is the golang structure for table company_profiles.
type CompanyProfiles struct {
	UserId           uint64    `json:"userId"           orm:"user_id"           description:""` //
	IntroductionHtml string    `json:"introductionHtml" orm:"introduction_html" description:""` //
	Offices          string    `json:"offices"          orm:"offices"           description:""` //
	Factories        string    `json:"factories"        orm:"factories"         description:""` //
	CreatedAt        time.Time `json:"createdAt"        orm:"created_at"        description:""` //
	UpdatedAt        time.Time `json:"updatedAt"        orm:"updated_at"        description:""` //
}
