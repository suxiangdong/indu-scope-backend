package adminhotcategories

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"

	"github.com/suxiangdong/indu-scope-backend/api/adminhotcategories/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	hotcategoryModel "github.com/suxiangdong/indu-scope-backend/internal/model/hotcategory"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (*v1.ListRes, error) {
	filter := hotcategoryModel.ListFilter{}
	if req != nil {
		filter.Keyword, filter.Status = req.Keyword, req.Status
	}
	result, err := service.HotCategories().List(ctx, filter)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	items := make([]v1.HotCategoryInfo, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toAPI(item))
	}
	return &v1.ListRes{Items: items}, nil
}

func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (*v1.CreateRes, error) {
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "hot category create request is required")
	}
	item, err := service.HotCategories().Create(ctx, req.CategoryID, req.SortOrder, statusValue(req.Status))
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.CreateRes{HotCategoryInfo: toAPI(item)}, nil
}

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (*v1.UpdateRes, error) {
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "hot category update request is required")
	}
	item, err := service.HotCategories().Update(ctx, req.ID, req.CategoryID, req.SortOrder, statusValue(req.Status))
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.UpdateRes{HotCategoryInfo: toAPI(item)}, nil
}

func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (*v1.DeleteRes, error) {
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "hot category delete request is required")
	}
	if err := service.HotCategories().Delete(ctx, req.ID); err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.DeleteRes{Success: true}, nil
}

func (c *ControllerV1) Status(ctx context.Context, req *v1.StatusReq) (*v1.StatusRes, error) {
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "hot category status request is required")
	}
	item, err := service.HotCategories().UpdateStatus(ctx, req.ID, req.Status)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.StatusRes{HotCategoryInfo: toAPI(item)}, nil
}

func (c *ControllerV1) Sort(ctx context.Context, req *v1.SortReq) (*v1.SortRes, error) {
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "hot category sort request is required")
	}
	items := make([]hotcategoryModel.SortItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, hotcategoryModel.SortItem{ID: item.ID, SortOrder: item.SortOrder})
	}
	result, err := service.HotCategories().Sort(ctx, items)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	response := make([]v1.HotCategoryInfo, 0, len(result.Items))
	for _, item := range result.Items {
		response = append(response, toAPI(item))
	}
	return &v1.SortRes{Success: true, Items: response}, nil
}

func statusValue(value *uint8) uint8 {
	if value == nil {
		return 1
	}
	return *value
}

func toAPI(item hotcategoryModel.Item) v1.HotCategoryInfo {
	return v1.HotCategoryInfo{ID: item.ID, CategoryID: item.CategoryID, Name: item.Name, SortOrder: item.SortOrder, Status: item.Status, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}

func toAPIError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, consts.ErrHotCategoryAlreadyExists), errors.Is(err, consts.ErrHotCategoryNotFound), errors.Is(err, consts.ErrCategoryNotFound):
		return gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
	}
	var parameterError *consts.InvalidParameterError
	if errors.As(err, &parameterError) {
		return gerror.NewCode(gcode.CodeInvalidParameter, parameterError.Error())
	}
	glog.Errorf(ctx, "hot categories internal error: %+v", err)
	return gerror.NewCode(gcode.CodeInternalError, "internal server error")
}
