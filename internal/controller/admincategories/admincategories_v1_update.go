package admincategories

import (
	"context"

	"github.com/suxiangdong/indu-scope-backend/api/admincategories/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	item, err := service.AdminCategories().Update(ctx, req.ID, normalizeParentID(req.ParentID), req.Name, req.SortOrder, statusValue(req.Status))
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.UpdateRes{CategoryInfo: toAPICategory(item)}, nil
}
