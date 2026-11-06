package adminverification

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/suxiangdong/indu-scope-backend/api/adminverification/v1"
	adminverificationmodel "github.com/suxiangdong/indu-scope-backend/internal/model/adminverification"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (*v1.ListRes, error) {
	result, err := service.AdminVerification().List(ctx, adminverificationmodel.ListFilter{Type: req.Type, ReviewStatus: req.ReviewStatus})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	items := make([]v1.VerificationListItem, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toAPIListItem(item))
	}
	return &v1.ListRes{Items: items}, nil
}

func (c *ControllerV1) Detail(ctx context.Context, req *v1.DetailReq) (*v1.DetailRes, error) {
	result, err := service.AdminVerification().Detail(ctx, req.Type, req.ID)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.DetailRes{DetailInfo: toAPIDetail(result)}, nil
}

func (c *ControllerV1) Review(ctx context.Context, req *v1.ReviewReq) (*v1.ReviewRes, error) {
	value := service.Bizctx().Get(ctx)
	if value == nil || value.AdminUser == nil {
		return nil, toAPIError(ctx, gerror.New("admin authorization is required"))
	}
	result, err := service.AdminVerification().Review(ctx, uint64(value.AdminUser.ID), adminverificationmodel.ReviewInput{
		ID: req.ID, Type: req.Type, Action: req.Action, Reason: req.Reason,
	})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.ReviewRes{DetailInfo: toAPIDetail(result)}, nil
}
