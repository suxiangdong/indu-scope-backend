package admincategories

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"

	"github.com/suxiangdong/indu-scope-backend/api/admincategories/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	categorymodel "github.com/suxiangdong/indu-scope-backend/internal/model/category"
)

func toAPICategory(item categorymodel.CategoryItem) v1.CategoryInfo {
	var parentID *int64
	if item.ParentID != nil {
		value := int64(*item.ParentID)
		parentID = &value
	}
	return v1.CategoryInfo{
		ID: int64(item.ID), ParentID: parentID, Name: item.Name,
		SortOrder: item.SortOrder, Status: item.Status,
		CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}

func toAPIError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, consts.ErrCategoryAlreadyExists),
		errors.Is(err, consts.ErrCategoryNotFound),
		errors.Is(err, consts.ErrCategoryParentNotFound),
		errors.Is(err, consts.ErrCategoryParentCycle),
		errors.Is(err, consts.ErrCategoryHasChildren),
		errors.Is(err, consts.ErrCategoryHasProducts):
		return gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
	}
	var parameterError *consts.InvalidParameterError
	if errors.As(err, &parameterError) {
		return gerror.NewCode(gcode.CodeInvalidParameter, parameterError.Error())
	}
	glog.Errorf(ctx, "admin categories internal error: %+v", err)
	return gerror.NewCode(gcode.CodeInternalError, "internal server error")
}

func normalizeParentID(parentID *uint64) *uint64 {
	if parentID == nil {
		value := uint64(0)
		return &value
	}
	value := *parentID
	return &value
}

func statusValue(status *uint) uint {
	if status == nil {
		return 1
	}
	return *status
}
