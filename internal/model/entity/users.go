// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Users is the golang structure for table users.
type Users struct {
	Id           uint64    `json:"id"           orm:"id"            description:""`  //
	Phone        string    `json:"phone"        orm:"phone"         description:""`  //
	CompanyName  string    `json:"companyName"  orm:"company_name"   description:""` //
	WechatOpenID string    `json:"wechatOpenID" orm:"wechat_openid" description:""`  //
	PasswordHash string    `json:"passwordHash" orm:"password_hash" description:""`  //
	Status       uint      `json:"status"       orm:"status"        description:""`  //
	Role         uint      `json:"role"         orm:"role"          description:""`  //
	UserType     uint      `json:"userType"     orm:"user_type"     description:""`  //
	CreatedAt    time.Time `json:"createdAt"    orm:"created_at"    description:""`  //
	UpdatedAt    time.Time `json:"updatedAt"    orm:"updated_at"    description:""`  //
}
