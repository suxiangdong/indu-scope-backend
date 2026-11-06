// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package auth

import (
	"context"

	"github.com/suxiangdong/indu-scope-backend/api/auth/v1"
)

type IAuthV1 interface {
	Register(ctx context.Context, req *v1.RegisterReq) (res *v1.RegisterRes, err error)
	Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error)
	WechatPhoneLogin(ctx context.Context, req *v1.WechatPhoneLoginReq) (res *v1.WechatPhoneLoginRes, err error)
	ListStorePlatforms(ctx context.Context, req *v1.ListStorePlatformsReq) (res *v1.ListStorePlatformsRes, err error)
	GetVerification(ctx context.Context, req *v1.GetVerificationReq) (res *v1.GetVerificationRes, err error)
	UploadBusinessLicense(ctx context.Context, req *v1.UploadBusinessLicenseReq) (res *v1.UploadBusinessLicenseRes, err error)
	UploadIDCard(ctx context.Context, req *v1.UploadIDCardReq) (res *v1.UploadIDCardRes, err error)
	SubmitPersonalVerification(ctx context.Context, req *v1.SubmitPersonalVerificationReq) (res *v1.SubmitPersonalVerificationRes, err error)
	SubmitCompanyVerification(ctx context.Context, req *v1.SubmitCompanyVerificationReq) (res *v1.SubmitCompanyVerificationRes, err error)
	Me(ctx context.Context, req *v1.MeReq) (res *v1.MeRes, err error)
	Logout(ctx context.Context, req *v1.LogoutReq) (res *v1.LogoutRes, err error)
}
