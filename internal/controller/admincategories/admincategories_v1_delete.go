package admincategories

import (
	"context"

	"github.com/suxiangdong/indu-scope-backend/api/admincategories/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	if req == nil {
		return nil, invalidRequest("category delete request is required")
	}
	if err := service.AdminCategories().Delete(ctx, req.ID); err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.DeleteRes{Success: true}, nil
}
