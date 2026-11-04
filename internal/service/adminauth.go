// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"

	authmodel "github.com/suxiangdong/indu-scope-backend/internal/model/auth"
)

type (
	IAdminAuth interface {
		Login(ctx context.Context, identifier string, password string) (authmodel.AdminAuthResult, error)
		// Me returns the active administrator represented by an admin access token.
		Me(ctx context.Context, accessToken string) (authmodel.AdminUser, error)
		// Logout removes the Redis session represented by an admin access token.
		Logout(ctx context.Context, accessToken string) error
	}
)

var (
	localAdminAuth IAdminAuth
)

func AdminAuth() IAdminAuth {
	if localAdminAuth == nil {
		panic("implement not found for interface IAdminAuth, forgot register?")
	}
	return localAdminAuth
}

func RegisterAdminAuth(i IAdminAuth) {
	localAdminAuth = i
}
