package adminappcontent

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/suxiangdong/indu-scope-backend/api/adminappcontent/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	appcontentModel "github.com/suxiangdong/indu-scope-backend/internal/model/appcontent"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) List(ctx context.Context, _ *v1.ListReq) (*v1.ListRes, error) {
	result, err := service.AppContent().List(ctx)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInternalError, "internal server error")
	}
	items := make([]v1.PageContent, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toAPI(item))
	}
	return &v1.ListRes{Items: items}, nil
}

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (*v1.UpdateRes, error) {
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "update request is required")
	}
	adminID := uint64(0)
	if admin := service.Bizctx().Get(ctx); admin != nil && admin.AdminUser != nil {
		adminID = uint64(admin.AdminUser.ID)
	}
	item, err := service.AppContent().Update(ctx, appcontentModel.UpdateInput{AdminUserID: adminID, PageKey: req.PageKey, Title: req.Title, Summary: req.Summary, Body: req.Body, ContactPhone: req.ContactPhone, ContactEmail: req.ContactEmail, ServiceHours: req.ServiceHours, Address: req.Address, WechatAccount: req.WechatAccount, Status: statusValue(req.Status)})
	if err != nil {
		return nil, toAPIError(err)
	}
	return &v1.UpdateRes{PageContent: toAPI(item)}, nil
}

func statusValue(value *uint8) uint8 {
	if value == nil {
		return 1
	}
	return *value
}
func toAPI(item appcontentModel.PageContent) v1.PageContent {
	return v1.PageContent{ID: item.ID, PageKey: item.PageKey, Title: item.Title, Summary: item.Summary, Body: item.Body, ContactPhone: item.ContactPhone, ContactEmail: item.ContactEmail, ServiceHours: item.ServiceHours, Address: item.Address, WechatAccount: item.WechatAccount, Status: item.Status}
}
func toAPIError(err error) error {
	if errors.Is(err, consts.ErrAppPageNotFound) {
		return gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
	}
	var parameterError *consts.InvalidParameterError
	if errors.As(err, &parameterError) {
		return gerror.NewCode(gcode.CodeInvalidParameter, parameterError.Error())
	}
	return gerror.NewCode(gcode.CodeInternalError, "internal server error")
}
