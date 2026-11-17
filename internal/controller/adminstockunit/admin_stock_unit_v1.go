package adminstockunit

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"

	"github.com/suxiangdong/indu-scope-backend/api/adminstockunit/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	stockUnitModel "github.com/suxiangdong/indu-scope-backend/internal/model/stockunit"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) List(ctx context.Context, _ *v1.ListReq) (*v1.ListRes, error) {
	result, err := service.AdminStockUnit().List(ctx)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	items := make([]v1.StockUnitInfo, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toAPI(item))
	}
	return &v1.ListRes{Items: items}, nil
}

func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (*v1.CreateRes, error) {
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "stock unit create request is required")
	}
	item, err := service.AdminStockUnit().Create(ctx, stockUnitModel.Input{Code: req.Code, Name: req.Name, SortOrder: req.SortOrder, Status: statusValue(req.Status)})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.CreateRes{StockUnitInfo: toAPI(item)}, nil
}

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (*v1.UpdateRes, error) {
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "stock unit update request is required")
	}
	item, err := service.AdminStockUnit().Update(ctx, req.ID, stockUnitModel.Input{Code: req.Code, Name: req.Name, SortOrder: req.SortOrder, Status: statusValue(req.Status)})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.UpdateRes{StockUnitInfo: toAPI(item)}, nil
}

func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (*v1.DeleteRes, error) {
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "stock unit delete request is required")
	}
	if err := service.AdminStockUnit().Delete(ctx, req.ID); err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.DeleteRes{Success: true}, nil
}

func (c *ControllerV1) ListCategories(ctx context.Context, req *v1.ListCategoriesReq) (*v1.ListCategoriesRes, error) {
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "stock unit category list request is required")
	}
	result, err := service.AdminStockUnit().ListCategories(ctx, req.StockUnitID)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	items := make([]v1.CategoryInfo, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toAPICategory(item))
	}
	return &v1.ListCategoriesRes{Items: items}, nil
}

func (c *ControllerV1) BindCategory(ctx context.Context, req *v1.BindCategoryReq) (*v1.BindCategoryRes, error) {
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "stock unit category bind request is required")
	}
	item, err := service.AdminStockUnit().BindCategory(ctx, req.StockUnitID, req.CategoryID, req.SortOrder)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.BindCategoryRes{CategoryInfo: toAPICategory(item)}, nil
}

func (c *ControllerV1) UnbindCategory(ctx context.Context, req *v1.UnbindCategoryReq) (*v1.UnbindCategoryRes, error) {
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "stock unit category unbind request is required")
	}
	if err := service.AdminStockUnit().UnbindCategory(ctx, req.StockUnitID, req.CategoryID); err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.UnbindCategoryRes{Success: true}, nil
}

func statusValue(value *uint) uint {
	if value == nil {
		return 1
	}
	return *value
}

func toAPI(item stockUnitModel.Item) v1.StockUnitInfo {
	return v1.StockUnitInfo{ID: item.ID, Code: item.Code, Name: item.Name, SortOrder: item.SortOrder, Status: item.Status, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}

func toAPICategory(item stockUnitModel.CategoryItem) v1.CategoryInfo {
	return v1.CategoryInfo{CategoryID: item.CategoryID, CategoryName: item.CategoryName, SortOrder: item.SortOrder, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}

func toAPIError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, consts.ErrStockUnitAlreadyExists), errors.Is(err, consts.ErrStockUnitNotFound), errors.Is(err, consts.ErrStockUnitInUse), errors.Is(err, consts.ErrStockUnitCategoryExists), errors.Is(err, consts.ErrStockUnitCategoryNotFound), errors.Is(err, consts.ErrCategoryNotFound), errors.Is(err, consts.ErrCategoryHasChildren):
		return gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
	}
	var parameterError *consts.InvalidParameterError
	if errors.As(err, &parameterError) {
		return gerror.NewCode(gcode.CodeInvalidParameter, parameterError.Error())
	}
	glog.Errorf(ctx, "admin stock unit internal error: %+v", err)
	return gerror.NewCode(gcode.CodeInternalError, "internal server error")
}
