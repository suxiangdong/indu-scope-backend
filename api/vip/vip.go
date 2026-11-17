package vip

import (
	"context"

	v1 "github.com/suxiangdong/indu-scope-backend/api/vip/v1"
)

type IVipV1 interface {
	Plans(ctx context.Context, req *v1.PlansReq) (*v1.PlansRes, error)
	Agreements(ctx context.Context, req *v1.AgreementsReq) (*v1.AgreementsRes, error)
	Me(ctx context.Context, req *v1.MeReq) (*v1.MeRes, error)
	CreateOrder(ctx context.Context, req *v1.CreateOrderReq) (*v1.CreateOrderRes, error)
	Order(ctx context.Context, req *v1.OrderReq) (*v1.OrderRes, error)
	PayOrder(ctx context.Context, req *v1.PayOrderReq) (*v1.PayOrderRes, error)
	CancelOrder(ctx context.Context, req *v1.CancelOrderReq) (*v1.CancelOrderRes, error)
	CancelAutoRenew(ctx context.Context, req *v1.CancelAutoRenewReq) (*v1.CancelAutoRenewRes, error)
}
