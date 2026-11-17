package auth

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/suxiangdong/indu-scope-backend/api/auth/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) WechatPhoneLogin(ctx context.Context, req *v1.WechatPhoneLoginReq) (res *v1.WechatPhoneLoginRes, err error) {
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "wechat login request is required")
	}
	result, err := service.Auth().WechatPhoneLogin(ctx, req.LoginCode, req.PhoneCode)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.WechatPhoneLoginRes{AuthData: toAPIData(result)}, nil
}
