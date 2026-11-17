package admincategories

import (
	"context"

	"github.com/suxiangdong/indu-scope-backend/api/admincategories/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	if req == nil {
		return nil, invalidRequest("category create request is required")
	}
	item, err := service.AdminCategories().Create(ctx, normalizeParentID(req.ParentID), req.Name, req.SortOrder, statusValue(req.Status))
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.CreateRes{CategoryInfo: toAPICategory(item)}, nil
}
