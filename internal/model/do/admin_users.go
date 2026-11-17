// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// AdminUsers is the golang structure of table admin_users for DAO operations like Where/Data.
type AdminUsers struct {
	g.Meta       `orm:"table:admin_users, do:true"`
	Id           any //
	Username     any //
	Name         any //
	Phone        any //
	PasswordHash any //
	Status       any //
	IsSuperAdmin any //
	LastLoginAt  any //
	CreatedAt    any //
	UpdatedAt    any //
}
