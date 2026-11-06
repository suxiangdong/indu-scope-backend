package adminstoreplatform

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"

	"github.com/suxiangdong/indu-scope-backend/api/adminstoreplatform/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	storePlatformModel "github.com/suxiangdong/indu-scope-backend/internal/model/storeplatform"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) List(ctx context.Context, _ *v1.ListReq) (*v1.ListRes, error) {
	result, err := service.AdminStorePlatform().List(ctx)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	items := make([]v1.StorePlatformInfo, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toAPI(item))
	}
	return &v1.ListRes{Items: items}, nil
}

func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (*v1.CreateRes, error) {
	item, err := service.AdminStorePlatform().Create(ctx, storePlatformModel.Input{Name: req.Name, SortOrder: req.SortOrder, Status: statusValue(req.Status)})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.CreateRes{StorePlatformInfo: toAPI(item)}, nil
}

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (*v1.UpdateRes, error) {
	item, err := service.AdminStorePlatform().Update(ctx, req.ID, storePlatformModel.Input{Name: req.Name, SortOrder: req.SortOrder, Status: statusValue(req.Status)})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.UpdateRes{StorePlatformInfo: toAPI(item)}, nil
}

func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (*v1.DeleteRes, error) {
	if err := service.AdminStorePlatform().Delete(ctx, req.ID); err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.DeleteRes{Success: true}, nil
}

func statusValue(value *uint) uint {
	if value == nil {
		return 1
	}
	return *value
}

func toAPI(item storePlatformModel.Item) v1.StorePlatformInfo {
	return v1.StorePlatformInfo{ID: item.ID, Name: item.Name, SortOrder: item.SortOrder, Status: item.Status, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}

func toAPIError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, consts.ErrStorePlatformAlreadyExists), errors.Is(err, consts.ErrStorePlatformNotFound), errors.Is(err, consts.ErrStorePlatformInUse):
		return gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
	}
	var parameterError *consts.InvalidParameterError
	if errors.As(err, &parameterError) {
		return gerror.NewCode(gcode.CodeInvalidParameter, parameterError.Error())
	}
	glog.Errorf(ctx, "admin store platform internal error: %+v", err)
	return gerror.NewCode(gcode.CodeInternalError, "internal server error")
}
