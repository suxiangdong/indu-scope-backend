package categories

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"

	"github.com/suxiangdong/indu-scope-backend/api/categories/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) Hot(ctx context.Context, _ *v1.HotReq) (*v1.HotRes, error) {
	result, err := service.HotCategories().ListPublic(ctx)
	if err != nil {
		glog.Errorf(ctx, "list public hot categories: %+v", err)
		return nil, gerror.NewCode(gcode.CodeInternalError, "internal server error")
	}
	items := make([]v1.HotCategoryInfo, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, v1.HotCategoryInfo{ID: item.ID, CategoryID: item.CategoryID, Name: item.Name, SortOrder: item.SortOrder})
	}
	return &v1.HotRes{Items: items}, nil
}
