package adminusers

import (
	"context"

	"github.com/suxiangdong/indu-scope-backend/api/adminusers/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	if err := service.AdminUsers().Delete(ctx, req.ID); err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.DeleteRes{Success: true}, nil
}
