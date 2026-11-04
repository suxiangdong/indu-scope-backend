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
	IAuth interface {
		Register(ctx context.Context, phone string, password string, role uint8, userType uint8) (authmodel.AuthResult, error)
		Login(ctx context.Context, phone string, password string) (authmodel.AuthResult, error)
		// Me returns the active user represented by a user access token and its Redis session.
		Me(ctx context.Context, accessToken string) (authmodel.User, error)
		// Logout removes the Redis session represented by a user access token.
		Logout(ctx context.Context, accessToken string) error
	}
)

var (
	localAuth IAuth
)

func Auth() IAuth {
	if localAuth == nil {
		panic("implement not found for interface IAuth, forgot register?")
	}
	return localAuth
}

func RegisterAuth(i IAuth) {
	localAuth = i
}
