// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// AdminUsers is the golang structure for table admin_users.
type AdminUsers struct {
	Id           uint64    `json:"id"           orm:"id"            description:""`  //
	Username     string    `json:"username"     orm:"username"      description:""`  //
	Name         string    `json:"name"         orm:"name"          description:""`  //
	Phone        string    `json:"phone"        orm:"phone"         description:""`  //
	PasswordHash string    `json:"passwordHash" orm:"password_hash" description:""`  //
	Status       uint      `json:"status"       orm:"status"        description:""`  //
	IsSuperAdmin uint      `json:"isSuperAdmin" orm:"is_super_admin" description:""` //
	LastLoginAt  time.Time `json:"lastLoginAt"  orm:"last_login_at" description:""`  //
	CreatedAt    time.Time `json:"createdAt"    orm:"created_at"    description:""`  //
	UpdatedAt    time.Time `json:"updatedAt"    orm:"updated_at"    description:""`  //
}
