package admincustomerusers

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"

	"github.com/suxiangdong/indu-scope-backend/api/admincustomerusers/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	adminCustomerUserModel "github.com/suxiangdong/indu-scope-backend/internal/model/admincustomeruser"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (*v1.ListRes, error) {
	result, err := service.AdminCustomerUsers().List(ctx, adminCustomerUserModel.ListFilter{
		Phone: req.Phone, Role: req.Role, UserType: req.UserType, Status: req.Status,
	})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	items := make([]v1.CustomerUserInfo, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toAPI(item))
	}
	return &v1.ListRes{Items: items}, nil
}

func (c *ControllerV1) UpdateStatus(ctx context.Context, req *v1.UpdateStatusReq) (*v1.UpdateStatusRes, error) {
	item, err := service.AdminCustomerUsers().UpdateStatus(ctx, req.ID, req.Status)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.UpdateStatusRes{CustomerUserInfo: toAPI(item)}, nil
}

func toAPI(item adminCustomerUserModel.UserItem) v1.CustomerUserInfo {
	return v1.CustomerUserInfo{
		ID: item.ID, Phone: item.Phone, Role: item.Role, UserType: item.UserType, Status: item.Status,
		VerificationType: item.VerificationType, VerificationStatus: item.VerificationStatus,
		WechatBound: item.WechatBound, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}

func toAPIError(ctx context.Context, err error) error {
	if errors.Is(err, consts.ErrCustomerUserNotFound) {
		return gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
	}
	var parameterError *consts.InvalidParameterError
	if errors.As(err, &parameterError) {
		return gerror.NewCode(gcode.CodeInvalidParameter, parameterError.Error())
	}
	glog.Errorf(ctx, "admin customer users internal error: %+v", err)
	return gerror.NewCode(gcode.CodeInternalError, "internal server error")
}
