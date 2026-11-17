package adminvip

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	v1 "github.com/suxiangdong/indu-scope-backend/api/adminvip/v1"
	vipModel "github.com/suxiangdong/indu-scope-backend/internal/model/vip"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *RefundControllerV1) CreateRefund(ctx context.Context, req *v1.CreateRefundReq) (*v1.RefundRes, error) {
	value := service.Bizctx().Get(ctx)
	if value == nil || value.AdminUser == nil {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "管理员登录已失效")
	}
	item, err := service.Vip().CreateRefund(ctx, req.ID, uint64(value.AdminUser.ID), req.Reason)
	if err != nil {
		return nil, toAPIError(err)
	}
	return &v1.RefundRes{Refund: *item}, nil
}

func (c *ControllerV1) Refund(ctx context.Context, req *v1.RefundReq) (*v1.RefundRes, error) {
	item, err := service.Vip().SyncRefund(ctx, req.ID)
	if err != nil {
		return nil, toAPIError(err)
	}
	return &v1.RefundRes{Refund: *item}, nil
}

func (c *RefundControllerV1) RetryRefund(ctx context.Context, req *v1.RetryRefundReq) (*v1.RefundRes, error) {
	item, err := service.Vip().RetryRefund(ctx, req.ID)
	if err != nil {
		return nil, toAPIError(err)
	}
	return &v1.RefundRes{Refund: *item}, nil
}

func (c *ControllerV1) ListRefunds(ctx context.Context, req *v1.ListRefundsReq) (*v1.ListRefundsRes, error) {
	result, err := service.Vip().ListRefunds(ctx, vipModel.RefundListFilter{OrderID: req.OrderID, OrderNo: req.OrderNo, RefundNo: req.RefundNo, Status: req.Status, Page: req.Page, PageSize: req.PageSize})
	if err != nil {
		return nil, toAPIError(err)
	}
	return &v1.ListRefundsRes{RefundListResult: result}, nil
}
