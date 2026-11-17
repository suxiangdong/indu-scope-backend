package adminvip

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/suxiangdong/indu-scope-backend/api/adminvip/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	vipModel "github.com/suxiangdong/indu-scope-backend/internal/model/vip"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) ListPlans(ctx context.Context, req *v1.ListPlansReq) (*v1.ListPlansRes, error) {
	items, err := service.Vip().ListPlans(ctx, false)
	if err != nil {
		return nil, toAPIError(err)
	}
	result := make([]v1.PlanInfo, 0, len(items))
	for _, item := range items {
		if req != nil && req.Status != nil && item.Status != *req.Status {
			continue
		}
		result = append(result, toPlanAPI(item))
	}
	return &v1.ListPlansRes{Items: result}, nil
}

func (c *ControllerV1) ListAgreements(ctx context.Context, _ *v1.ListAgreementsReq) (*v1.ListAgreementsRes, error) {
	items, err := service.Vip().ListAgreements(ctx, false)
	if err != nil {
		return nil, toAPIError(err)
	}
	result := make([]v1.AgreementInfo, 0, len(items))
	for _, item := range items {
		result = append(result, toAgreementAPI(item))
	}
	return &v1.ListAgreementsRes{Items: result}, nil
}

func (c *ControllerV1) UpdateAgreement(ctx context.Context, req *v1.UpdateAgreementReq) (*v1.UpdateAgreementRes, error) {
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "vip agreement request is required")
	}
	adminID := uint64(0)
	if value := service.Bizctx().Get(ctx); value != nil && value.AdminUser != nil {
		adminID = uint64(value.AdminUser.ID)
	}
	item, err := service.Vip().UpdateAgreement(ctx, vipModel.AgreementInput{Key: req.Key, Title: req.Title, Content: req.Content, Status: statusValue(req.Status), AdminUserID: adminID})
	if err != nil {
		return nil, toAPIError(err)
	}
	return &v1.UpdateAgreementRes{AgreementInfo: toAgreementAPI(item)}, nil
}

func (c *ControllerV1) CreatePlan(ctx context.Context, req *v1.CreatePlanReq) (*v1.PlanRes, error) {
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "vip plan request is required")
	}
	item, err := service.Vip().CreatePlan(ctx, vipModel.PlanInput{Name: req.Name, Description: req.Description, BillingCycle: req.BillingCycle, PlanType: req.PlanType, DurationDays: req.DurationDays, PriceFen: req.PriceFen, OriginalPriceFen: req.OriginalPriceFen, AutoRenewEnabled: statusValue(req.AutoRenewEnabled), SortOrder: req.SortOrder})
	if err != nil {
		return nil, toAPIError(err)
	}
	return &v1.PlanRes{PlanInfo: toPlanAPI(item)}, nil
}

func (c *ControllerV1) UpdatePlan(ctx context.Context, req *v1.UpdatePlanReq) (*v1.PlanRes, error) {
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "vip plan request is required")
	}
	item, err := service.Vip().UpdatePlan(ctx, req.ID, vipModel.PlanInput{Name: req.Name, Description: req.Description, BillingCycle: req.BillingCycle, PlanType: req.PlanType, DurationDays: req.DurationDays, PriceFen: req.PriceFen, OriginalPriceFen: req.OriginalPriceFen, AutoRenewEnabled: statusValue(req.AutoRenewEnabled), SortOrder: req.SortOrder})
	if err != nil {
		return nil, toAPIError(err)
	}
	return &v1.PlanRes{PlanInfo: toPlanAPI(item)}, nil
}

func (c *ControllerV1) UpdatePlanStatus(ctx context.Context, req *v1.UpdatePlanStatusReq) (*v1.PlanRes, error) {
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "vip plan status request is required")
	}
	item, err := service.Vip().UpdatePlanStatus(ctx, req.ID, req.Status)
	if err != nil {
		return nil, toAPIError(err)
	}
	return &v1.PlanRes{PlanInfo: toPlanAPI(item)}, nil
}

func (c *ControllerV1) ListOrders(ctx context.Context, req *v1.ListOrdersReq) (*v1.ListOrdersRes, error) {
	filter := vipModel.OrderListFilter{Page: 1, PageSize: 20}
	if req != nil {
		filter.Status, filter.Page, filter.PageSize = req.Status, req.Page, req.PageSize
		filter.OrderNo = req.OrderNo
	}
	result, err := service.Vip().ListOrders(ctx, filter)
	if err != nil {
		return nil, toAPIError(err)
	}
	items := make([]v1.OrderInfo, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, v1.OrderInfo{ID: item.ID, OrderNo: item.OrderNo, UserID: item.UserID, UserCompany: item.UserCompany, PlanName: item.PlanName, PlanCodeSnapshot: item.PlanCodeSnapshot, PlanNameSnapshot: item.PlanNameSnapshot, PlanDescriptionSnapshot: item.PlanDescriptionSnapshot, OriginalPriceFenSnapshot: item.OriginalPriceFenSnapshot, PlanTypeSnapshot: item.PlanTypeSnapshot, BillingCycleSnapshot: item.BillingCycleSnapshot, DurationDaysSnapshot: item.DurationDaysSnapshot, AutoRenewEnabledSnapshot: item.AutoRenewEnabledSnapshot, AmountFen: item.AmountFen, Currency: item.Currency, Status: item.Status, PaymentMethod: item.PaymentMethod, ProviderTransaction: item.ProviderTransaction, EntitlementStartAt: item.EntitlementStartAt, EntitlementEndAt: item.EntitlementEndAt, LatestRefundID: item.LatestRefundID, RefundStatus: item.RefundStatus, ServiceAgreementID: item.ServiceAgreementID, ServiceAgreementVersion: item.ServiceAgreementVersion, AutoRenewAgreementID: item.AutoRenewAgreementID, AutoRenewAgreementVersion: item.AutoRenewAgreementVersion, CreatedAt: item.CreatedAt, PaidAt: item.PaidAt})
	}
	return &v1.ListOrdersRes{Items: items, Total: result.Total, Page: result.Page, PageSize: result.PageSize}, nil
}

func statusValue(v *uint8) uint8 {
	if v == nil {
		return 0
	}
	return *v
}
func toPlanAPI(item vipModel.Plan) v1.PlanInfo {
	return v1.PlanInfo{ID: item.ID, Code: item.Code, Name: item.Name, Description: item.Description, BillingCycle: item.BillingCycle, PlanType: item.PlanType, DurationDays: item.DurationDays, PriceFen: item.PriceFen, OriginalPriceFen: item.OriginalPriceFen, Currency: item.Currency, AutoRenew: item.AutoRenew, Status: item.Status, SortOrder: item.SortOrder, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}
func toAgreementAPI(item vipModel.Agreement) v1.AgreementInfo {
	return v1.AgreementInfo{ID: item.ID, Key: item.Key, Title: item.Title, Content: item.Content, Status: item.Status, Version: item.Version, UpdatedAt: item.UpdatedAt}
}
func toAPIError(err error) error {
	var p *consts.InvalidParameterError
	if errors.As(err, &p) {
		return gerror.NewCode(gcode.CodeInvalidParameter, p.Error())
	}
	return gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
}
