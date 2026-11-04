package admincategories

import (
	"context"

	"github.com/suxiangdong/indu-scope-backend/api/admincategories/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	result, err := service.AdminCategories().List(ctx)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	items := make([]v1.CategoryInfo, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toAPICategory(item))
	}
	return &v1.ListRes{Items: items}, nil
}
