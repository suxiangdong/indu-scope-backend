package auth

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/suxiangdong/indu-scope-backend/api/auth/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) Register(ctx context.Context, req *v1.RegisterReq) (res *v1.RegisterRes, err error) {
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "register request is required")
	}
	result, err := service.Auth().Register(ctx, req.Phone, req.Password, req.UserType)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.RegisterRes{AuthData: toAPIData(result)}, nil
}
