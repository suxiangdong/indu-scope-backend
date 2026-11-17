package adminvip

import (
	"context"

	v1 "github.com/suxiangdong/indu-scope-backend/api/adminvip/v1"
)

type IAdminVipV1 interface {
	ListPlans(ctx context.Context, req *v1.ListPlansReq) (*v1.ListPlansRes, error)
	ListAgreements(ctx context.Context, req *v1.ListAgreementsReq) (*v1.ListAgreementsRes, error)
	UpdateAgreement(ctx context.Context, req *v1.UpdateAgreementReq) (*v1.UpdateAgreementRes, error)
	CreatePlan(ctx context.Context, req *v1.CreatePlanReq) (*v1.PlanRes, error)
	UpdatePlan(ctx context.Context, req *v1.UpdatePlanReq) (*v1.PlanRes, error)
	UpdatePlanStatus(ctx context.Context, req *v1.UpdatePlanStatusReq) (*v1.PlanRes, error)
	ListOrders(ctx context.Context, req *v1.ListOrdersReq) (*v1.ListOrdersRes, error)
	Refund(ctx context.Context, req *v1.RefundReq) (*v1.RefundRes, error)
	ListRefunds(ctx context.Context, req *v1.ListRefundsReq) (*v1.ListRefundsRes, error)
}

type IAdminVipRefundV1 interface {
	CreateRefund(ctx context.Context, req *v1.CreateRefundReq) (*v1.RefundRes, error)
	RetryRefund(ctx context.Context, req *v1.RetryRefundReq) (*v1.RefundRes, error)
}
