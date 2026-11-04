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
	ISession interface {
		// SetUser stores the authenticated user in the current request session.
		SetUser(ctx context.Context, user authmodel.User) error
		// GetUser retrieves the authenticated user from the current request session.
		GetUser(ctx context.Context) (authmodel.User, error)
		// RemoveUser removes the authenticated user from the current request session.
		RemoveUser(ctx context.Context) error
		// RemoveAdminUser removes the authenticated administrator from the current session.
		RemoveAdminUser(ctx context.Context) error
		// SetAdminUser stores the authenticated administrator in the current session.
		SetAdminUser(ctx context.Context, user authmodel.AdminUser) error
		// GetAdminUser retrieves the authenticated administrator from the current session.
		GetAdminUser(ctx context.Context) (authmodel.AdminUser, error)
	}
)

var (
	localSession ISession
)

func Session() ISession {
	if localSession == nil {
		panic("implement not found for interface ISession, forgot register?")
	}
	return localSession
}

func RegisterSession(i ISession) {
	localSession = i
}
