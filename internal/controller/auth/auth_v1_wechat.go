package auth

import (
	"context"

	"github.com/suxiangdong/indu-scope-backend/api/auth/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) WechatPhoneLogin(ctx context.Context, req *v1.WechatPhoneLoginReq) (res *v1.WechatPhoneLoginRes, err error) {
	result, err := service.Auth().WechatPhoneLogin(ctx, req.LoginCode, req.PhoneCode)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.WechatPhoneLoginRes{AuthData: toAPIData(result)}, nil
}
