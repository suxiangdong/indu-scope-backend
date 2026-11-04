// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"

	adminusermodel "github.com/suxiangdong/indu-scope-backend/internal/model/adminuser"
)

type (
	IAdminUsers interface {
		// List returns all administrator accounts.
		List(ctx context.Context) (adminusermodel.AdminUserListResult, error)
		// Create creates an administrator account.
		Create(ctx context.Context, username string, name string, phone string, password string, status uint) (adminusermodel.AdminUserItem, error)
		// Update updates an administrator account. An empty password keeps the old password.
		Update(ctx context.Context, id uint64, username string, name string, phone string, password string, status uint) (adminusermodel.AdminUserItem, error)
		// Delete permanently deletes an administrator account.
		Delete(ctx context.Context, id uint64) error
	}
)

var (
	localAdminUsers IAdminUsers
)

func AdminUsers() IAdminUsers {
	if localAdminUsers == nil {
		panic("implement not found for interface IAdminUsers, forgot register?")
	}
	return localAdminUsers
}

func RegisterAdminUsers(i IAdminUsers) {
	localAdminUsers = i
}
