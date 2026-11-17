package analytics

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/suxiangdong/indu-scope-backend/api/analytics/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	analyticsModel "github.com/suxiangdong/indu-scope-backend/internal/model/analytics"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) RecordClick(ctx context.Context, req *v1.RecordClickReq) (*v1.RecordClickRes, error) {
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "click request is required")
	}
	value := service.Bizctx().Get(ctx)
	if value == nil || value.User == nil {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "authorization is required")
	}
	contentType := analyticsModel.ContentType(req.ContentType)
	if err := service.Analytics().RecordClick(ctx, uint64(value.User.ID), contentType, req.ContentID); err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.RecordClickRes{Success: true}, nil
}

func toAPIError(ctx context.Context, err error) error {
	if errors.Is(err, consts.ErrTokenInvalid) {
		return gerror.NewCode(gcode.CodeNotAuthorized, err.Error())
	}
	var parameterError *consts.InvalidParameterError
	if errors.As(err, &parameterError) {
		return gerror.NewCode(gcode.CodeInvalidParameter, parameterError.Error())
	}
	return gerror.NewCode(gcode.CodeInternalError, "internal server error")
}
