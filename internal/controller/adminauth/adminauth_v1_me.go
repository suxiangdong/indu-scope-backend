package adminauth

import (
	"context"

	"github.com/suxiangdong/indu-scope-backend/api/adminauth/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) Me(ctx context.Context, _ *v1.MeReq) (res *v1.MeRes, err error) {
	if value := service.Bizctx().Get(ctx); value != nil && value.AdminUser != nil {
		return &v1.MeRes{AdminUserInfo: toUserInfo(*value.AdminUser)}, nil
	}
	token, err := authorizationToken(ctx)
	if err != nil {
		return nil, err
	}
	user, err := service.AdminAuth().Me(ctx, token)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.MeRes{AdminUserInfo: toUserInfo(user)}, nil
}
