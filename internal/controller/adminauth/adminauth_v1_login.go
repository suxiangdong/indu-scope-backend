package adminauth

import (
	"context"

	"github.com/suxiangdong/indu-scope-backend/api/adminauth/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error) {
	result, err := service.AdminAuth().Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.LoginRes{AuthData: toAuthData(result)}, nil
}
