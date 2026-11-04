package adminusers

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"

	"github.com/suxiangdong/indu-scope-backend/api/adminusers/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	adminusermodel "github.com/suxiangdong/indu-scope-backend/internal/model/adminuser"
)

func toAPIUser(user adminusermodel.AdminUserItem) v1.AdminUserInfo {
	return v1.AdminUserInfo{
		ID: user.ID, Username: user.Username, Name: user.Name, Phone: user.Phone,
		Status: user.Status, LastLoginAt: user.LastLoginAt,
		CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt,
	}
}

func toAPIError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, consts.ErrAdminUserAlreadyExists),
		errors.Is(err, consts.ErrAdminUserNotFound),
		errors.Is(err, consts.ErrAdminUserSelfDelete):
		return gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
	}
	var parameterError *consts.InvalidParameterError
	if errors.As(err, &parameterError) {
		return gerror.NewCode(gcode.CodeInvalidParameter, parameterError.Error())
	}
	glog.Errorf(ctx, "admin users internal error: %+v", err)
	return gerror.NewCode(gcode.CodeInternalError, "internal server error")
}

func statusValue(status *uint) uint {
	if status == nil {
		return 1
	}
	return *status
}
