// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// CompanyContacts is the golang structure for table company_contacts.
type CompanyContacts struct {
	Id        uint64    `json:"id"        orm:"id"         description:""` //
	UserId    uint64    `json:"userId"    orm:"user_id"    description:""` //
	Name      string    `json:"name"      orm:"name"       description:""` //
	Position  string    `json:"position"  orm:"position"   description:""` //
	Phone     string    `json:"phone"     orm:"phone"      description:""` //
	CreatedAt time.Time `json:"createdAt" orm:"created_at" description:""` //
	UpdatedAt time.Time `json:"updatedAt" orm:"updated_at" description:""` //
}
