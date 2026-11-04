package auth

import (
	"context"

	"github.com/suxiangdong/indu-scope-backend/api/auth/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) Logout(ctx context.Context, _ *v1.LogoutReq) (res *v1.LogoutRes, err error) {
	token, err := authorizationToken(ctx)
	if err != nil {
		return nil, err
	}
	if err := service.Auth().Logout(ctx, token); err != nil {
		return nil, toAPIError(err)
	}
	return &v1.LogoutRes{}, nil
}
