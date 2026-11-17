// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AdminUsersDao is the data access object for the table admin_users.
type AdminUsersDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  AdminUsersColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// AdminUsersColumns defines and stores column names for the table admin_users.
type AdminUsersColumns struct {
	Id           string //
	Username     string //
	Name         string //
	Phone        string //
	PasswordHash string //
	Status       string //
	IsSuperAdmin string //
	LastLoginAt  string //
	CreatedAt    string //
	UpdatedAt    string //
}

// adminUsersColumns holds the columns for the table admin_users.
var adminUsersColumns = AdminUsersColumns{
	Id:           "id",
	Username:     "username",
	Name:         "name",
	Phone:        "phone",
	PasswordHash: "password_hash",
	Status:       "status",
	IsSuperAdmin: "is_super_admin",
	LastLoginAt:  "last_login_at",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
}

// NewAdminUsersDao creates and returns a new DAO object for table data access.
func NewAdminUsersDao(handlers ...gdb.ModelHandler) *AdminUsersDao {
	return &AdminUsersDao{
		group:    "default",
		table:    "admin_users",
		columns:  adminUsersColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AdminUsersDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AdminUsersDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AdminUsersDao) Columns() AdminUsersColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AdminUsersDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AdminUsersDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *AdminUsersDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
