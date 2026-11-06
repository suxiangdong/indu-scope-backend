package auth

import (
	"context"

	"github.com/suxiangdong/indu-scope-backend/api/auth/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) Me(ctx context.Context, _ *v1.MeReq) (res *v1.MeRes, err error) {
	if value := service.Bizctx().Get(ctx); value != nil && value.User != nil {
		return &v1.MeRes{UserInfo: toUserInfo(*value.User)}, nil
	}
	token, err := authorizationToken(ctx)
	if err != nil {
		return nil, err
	}
	user, err := service.Auth().Me(ctx, token)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.MeRes{UserInfo: toUserInfo(user)}, nil
}
