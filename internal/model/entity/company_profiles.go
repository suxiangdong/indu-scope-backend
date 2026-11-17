// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// CompanyProfiles is the golang structure for table company_profiles.
type CompanyProfiles struct {
	Id               uint64    `json:"id"               orm:"id"                description:""`  //
	UserId           uint64    `json:"userId"           orm:"user_id"           description:""`  //
	IntroductionHtml string    `json:"introductionHtml" orm:"introduction_html" description:""`  //
	StorePlatformId  uint64    `json:"storePlatformId"  orm:"store_platform_id"  description:""` //
	StoreUrl         string    `json:"storeUrl"         orm:"store_url"         description:""`  //
	Offices          string    `json:"offices"          orm:"offices"           description:""`  //
	Factories        string    `json:"factories"        orm:"factories"         description:""`  //
	CreatedAt        time.Time `json:"createdAt"        orm:"created_at"        description:""`  //
	UpdatedAt        time.Time `json:"updatedAt"        orm:"updated_at"        description:""`  //
}
