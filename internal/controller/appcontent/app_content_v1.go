package appcontent

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"

	"github.com/suxiangdong/indu-scope-backend/api/appcontent/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	appcontentModel "github.com/suxiangdong/indu-scope-backend/internal/model/appcontent"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) Get(ctx context.Context, req *v1.GetReq) (*v1.GetRes, error) {
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "page_key is required")
	}
	item, err := service.AppContent().Get(ctx, req.PageKey)
	if err != nil {
		if errors.Is(err, consts.ErrAppPageNotFound) {
			return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
		}
		glog.Errorf(ctx, "get app page content failed: %+v", err)
		return nil, gerror.NewCode(gcode.CodeInternalError, "internal server error")
	}
	return &v1.GetRes{PageContent: toAPI(item)}, nil
}

func toAPI(item appcontentModel.PageContent) v1.PageContent {
	return v1.PageContent{ID: item.ID, PageKey: item.PageKey, Title: item.Title, Summary: item.Summary, Body: item.Body, ContactPhone: item.ContactPhone, ContactEmail: item.ContactEmail, ServiceHours: item.ServiceHours, Address: item.Address, WechatAccount: item.WechatAccount, Status: item.Status}
}
