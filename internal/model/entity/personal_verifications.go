// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// PersonalVerifications is the golang structure for table personal_verifications.
type PersonalVerifications struct {
	UserId            uint64    `json:"userId"            orm:"user_id"             description:""` //
	RealName          string    `json:"realName"          orm:"real_name"           description:""` //
	IdNumberEncrypted []byte    `json:"idNumberEncrypted" orm:"id_number_encrypted" description:""` //
	IdFrontFileId     string    `json:"idFrontFileId"     orm:"id_front_file_id"    description:""` //
	IdBackFileId      string    `json:"idBackFileId"      orm:"id_back_file_id"     description:""` //
	VerifiedAt        time.Time `json:"verifiedAt"        orm:"verified_at"         description:""` //
	CreatedAt         time.Time `json:"createdAt"         orm:"created_at"          description:""` //
	UpdatedAt         time.Time `json:"updatedAt"         orm:"updated_at"          description:""` //
}
