// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// PersonalVerifications is the golang structure for table personal_verifications.
type PersonalVerifications struct {
	Id                uint64    `json:"id"                orm:"id"                 description:""`  //
	UserId            uint64    `json:"userId"            orm:"user_id"             description:""` //
	RealName          string    `json:"realName"          orm:"real_name"           description:""` //
	IdNumber          string    `json:"idNumber"          orm:"id_number"           description:""` //
	IdNumberEncrypted []byte    `json:"idNumberEncrypted" orm:"id_number_encrypted" description:""` //
	StorePlatformId   uint64    `json:"storePlatformId"   orm:"store_platform_id"   description:""` //
	StoreUrl          string    `json:"storeUrl"          orm:"store_url"           description:""` //
	IdFrontBucket     string    `json:"idFrontBucket"     orm:"id_front_bucket"    description:""`  //
	IdFrontEndpoint   string    `json:"idFrontEndpoint"   orm:"id_front_endpoint"  description:""`  //
	IdFrontPath       string    `json:"idFrontPath"       orm:"id_front_path"      description:""`  //
	IdBackBucket      string    `json:"idBackBucket"      orm:"id_back_bucket"     description:""`  //
	IdBackEndpoint    string    `json:"idBackEndpoint"    orm:"id_back_endpoint"   description:""`  //
	IdBackPath        string    `json:"idBackPath"        orm:"id_back_path"       description:""`  //
	VerifiedAt        time.Time `json:"verifiedAt"        orm:"verified_at"         description:""` //
	ReviewStatus      uint      `json:"reviewStatus"      orm:"review_status"       description:""` //
	ReviewReason      string    `json:"reviewReason"      orm:"review_reason"       description:""` //
	ReviewedBy        uint64    `json:"reviewedBy"        orm:"reviewed_by"         description:""` //
	ReviewedAt        time.Time `json:"reviewedAt"        orm:"reviewed_at"         description:""` //
	CreatedAt         time.Time `json:"createdAt"         orm:"created_at"          description:""` //
	UpdatedAt         time.Time `json:"updatedAt"         orm:"updated_at"          description:""` //
}
