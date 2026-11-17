package adminusers

import (
	"context"

	"github.com/suxiangdong/indu-scope-backend/api/adminusers/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	if req == nil {
		return nil, invalidRequest("admin user create request is required")
	}
	item, err := service.AdminUsers().Create(ctx, req.Username, req.Name, req.Phone, req.Password, statusValue(req.Status))
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.CreateRes{AdminUserInfo: toAPIUser(item)}, nil
}
