package service

import (
	"context"
	"net/http"

	vipModel "github.com/suxiangdong/indu-scope-backend/internal/model/vip"
)

type IVip interface {
	ListPlans(ctx context.Context, activeOnly bool) ([]vipModel.Plan, error)
	ListAgreements(ctx context.Context, activeOnly bool) ([]vipModel.Agreement, error)
	UpdateAgreement(ctx context.Context, input vipModel.AgreementInput) (vipModel.Agreement, error)
	GetSubscription(ctx context.Context, userID uint64) (*vipModel.Subscription, error)
	CreateOrder(ctx context.Context, userID uint64, planID uint64, autoRenew bool, idempotencyKey string) (vipModel.Order, bool, error)
	PreparePayment(ctx context.Context, userID, orderID uint64) (*vipModel.Checkout, error)
	SyncOrder(ctx context.Context, userID, orderID uint64) (vipModel.Order, error)
	CancelOrder(ctx context.Context, userID, orderID uint64) (vipModel.Order, error)
	HandleWechatPaymentNotification(ctx context.Context, request *http.Request) error
	CancelAutoRenew(ctx context.Context, userID uint64) error
	CreatePlan(ctx context.Context, input vipModel.PlanInput) (vipModel.Plan, error)
	UpdatePlan(ctx context.Context, id uint64, input vipModel.PlanInput) (vipModel.Plan, error)
	UpdatePlanStatus(ctx context.Context, id uint64, status uint8) (vipModel.Plan, error)
	ListOrders(ctx context.Context, filter vipModel.OrderListFilter) (vipModel.OrderListResult, error)
	CreateRefund(ctx context.Context, orderID, adminID uint64, reason string) (*vipModel.Refund, error)
	RetryRefund(ctx context.Context, id uint64) (*vipModel.Refund, error)
	SyncRefund(ctx context.Context, id uint64) (*vipModel.Refund, error)
	ListRefunds(ctx context.Context, filter vipModel.RefundListFilter) (vipModel.RefundListResult, error)
	HandleWechatRefundNotification(ctx context.Context, request *http.Request) error
}

var localVip IVip

func Vip() IVip {
	if localVip == nil {
		panic("implement not found for interface IVip, forgot register?")
	}
	return localVip
}

func RegisterVip(i IVip) {
	localVip = i
}
