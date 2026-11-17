package adminauth

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/suxiangdong/indu-scope-backend/api/adminauth/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error) {
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "admin login request is required")
	}
	result, err := service.AdminAuth().Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.LoginRes{AuthData: toAuthData(result)}, nil
}
