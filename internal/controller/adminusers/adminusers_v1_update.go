package adminusers

import (
	"context"

	"github.com/suxiangdong/indu-scope-backend/api/adminusers/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	item, err := service.AdminUsers().Update(ctx, req.ID, req.Username, req.Name, req.Phone, req.Password, statusValue(req.Status))
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.UpdateRes{AdminUserInfo: toAPIUser(item)}, nil
}
