package userprofile

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/suxiangdong/indu-scope-backend/api/userprofiles/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	userprofileModel "github.com/suxiangdong/indu-scope-backend/internal/model/userprofile"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) Profile(ctx context.Context, req *v1.ProfileReq) (*v1.ProfileRes, error) {
	if req == nil || req.ID == 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "id is required")
	}
	item, err := service.UserProfile().Profile(ctx, req.ID)
	if err != nil {
		if errors.Is(err, consts.ErrCustomerUserNotFound) {
			return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, "用户主页不存在或暂不可查看")
		}
		var parameterError *consts.InvalidParameterError
		if errors.As(err, &parameterError) {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, parameterError.Error())
		}
		return nil, gerror.NewCode(gcode.CodeInternalError, "internal server error")
	}
	return &v1.ProfileRes{ProfileInfo: toAPI(item)}, nil
}

func toAPI(item userprofileModel.Profile) v1.ProfileInfo {
	roleName := ""
	if item.Role == 1 {
		roleName = "采购方"
	} else if item.Role == 2 {
		roleName = "供货方"
	}
	return v1.ProfileInfo{
		ID: item.ID, CompanyName: item.CompanyName, Role: item.Role, RoleName: roleName,
		Introduction: item.Introduction, RegisteredAddress: item.RegisteredAddress,
		Offices: item.Offices, Factories: item.Factories,
	}
}
