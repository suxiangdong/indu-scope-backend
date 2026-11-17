// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"

	authmodel "github.com/suxiangdong/indu-scope-backend/internal/model/auth"
	storePlatformModel "github.com/suxiangdong/indu-scope-backend/internal/model/storeplatform"
	verificationModel "github.com/suxiangdong/indu-scope-backend/internal/model/verification"
)

type (
	IAuth interface {
		Register(ctx context.Context, phone string, password string, userType uint8) (authmodel.AuthResult, error)
		Login(ctx context.Context, phone string, password string) (authmodel.AuthResult, error)
		WechatPhoneLogin(ctx context.Context, loginCode string, phoneCode string) (authmodel.AuthResult, error)
		ListStorePlatforms(ctx context.Context) ([]storePlatformModel.Item, error)
		GetVerification(ctx context.Context) (verificationModel.Summary, error)
		UploadBusinessLicense(ctx context.Context, input verificationModel.BusinessLicenseUploadInput) (verificationModel.BusinessLicenseInfo, error)
		RecognizeBusinessLicense(ctx context.Context) (verificationModel.BusinessLicenseOCR, error)
		UploadIDCard(ctx context.Context, input verificationModel.IDCardUploadInput) (verificationModel.IDCardUploadInfo, error)
		SubmitPersonalVerification(ctx context.Context, input verificationModel.PersonalInput) (verificationModel.Summary, error)
		SubmitCompanyVerification(ctx context.Context, input verificationModel.CompanyInput) (verificationModel.Summary, error)
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
