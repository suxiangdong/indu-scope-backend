package vip

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/suxiangdong/indu-scope-backend/api/vip/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	vipModel "github.com/suxiangdong/indu-scope-backend/internal/model/vip"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) Plans(ctx context.Context, _ *v1.PlansReq) (*v1.PlansRes, error) {
	items, err := service.Vip().ListPlans(ctx, true)
	if err != nil {
		return nil, toAPIError(err)
	}
	result := make([]v1.PlanInfo, 0, len(items))
	for _, item := range items {
		result = append(result, toPlanAPI(item))
	}
	return &v1.PlansRes{Items: result}, nil
}

func (c *ControllerV1) Agreements(ctx context.Context, _ *v1.AgreementsReq) (*v1.AgreementsRes, error) {
	items, err := service.Vip().ListAgreements(ctx, true)
	if err != nil {
		return nil, toAPIError(err)
	}
	result := make([]v1.AgreementInfo, 0, len(items))
	for _, item := range items {
		if item.Key != vipModel.AgreementKeyService && item.Key != vipModel.AgreementKeyAutoRenew {
			continue
		}
		result = append(result, toAgreementAPI(item))
	}
	return &v1.AgreementsRes{Items: result}, nil
}

func (c *ControllerV1) Me(ctx context.Context, _ *v1.MeReq) (*v1.MeRes, error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := service.Vip().GetSubscription(ctx, userID)
	if err != nil {
		return nil, toAPIError(err)
	}
	if item == nil {
		return &v1.MeRes{}, nil
	}
	return &v1.MeRes{Subscription: toSubscriptionAPI(*item)}, nil
}

func (c *ControllerV1) CreateOrder(ctx context.Context, req *v1.CreateOrderReq) (*v1.CreateOrderRes, error) {
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "vip order request is required")
	}
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	order, _, err := service.Vip().CreateOrder(ctx, userID, req.PlanID, req.AutoRenew, req.Idempotency)
	if err != nil {
		return nil, toAPIError(err)
	}
	response, err := checkoutResponse(ctx, userID, order.ID)
	if err != nil {
		// The business order already exists, even if WeChat is unavailable.
		// Return its ID so the client can resume/close it instead of losing it.
		return &v1.CreateOrderRes{OrderID: order.ID, OrderNo: order.OrderNo, AmountFen: order.AmountFen, Currency: order.Currency,
			Message: "支付暂未就绪，可稍后继续该订单或取消订单"}, nil
	}
	return response, nil
}

func (c *ControllerV1) PayOrder(ctx context.Context, req *v1.PayOrderReq) (*v1.PayOrderRes, error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	return checkoutResponse(ctx, userID, req.ID)
}

func checkoutResponse(ctx context.Context, userID, orderID uint64) (*v1.CreateOrderRes, error) {
	checkout, err := service.Vip().PreparePayment(ctx, userID, orderID)
	if err != nil {
		g.Log().Errorf(ctx, "prepare vip payment failed: %v", err)
		return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, "订单暂不可支付，请刷新会员中心后重试；若仍失败请联系管理员")
	}
	order := checkout.Order
	result := &v1.CreateOrderRes{OrderID: order.ID, OrderNo: order.OrderNo, AmountFen: order.AmountFen, Currency: order.Currency,
		Activated: order.Status == "paid", ServiceAgreementVersion: order.ServiceAgreementVersion, AutoRenewAgreementVersion: order.AutoRenewAgreementVersion}
	if result.Activated {
		result.Message = "会员已生效"
	} else if p := checkout.Payment; p != nil {
		result.PaymentReady = true
		result.Payment = &v1.PaymentParams{TimeStamp: p.TimeStamp, NonceStr: p.NonceStr, Package: p.Package, SignType: p.SignType, PaySign: p.PaySign}
	} else {
		result.Message = "支付暂未开放，请联系平台客服"
	}
	return result, nil
}

func (c *ControllerV1) Order(ctx context.Context, req *v1.OrderReq) (*v1.OrderRes, error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	order, err := service.Vip().SyncOrder(ctx, userID, req.ID)
	if err != nil {
		return nil, paymentStatusError(ctx, err)
	}
	return orderStatusResponse(order), nil
}

func (c *ControllerV1) CancelOrder(ctx context.Context, req *v1.CancelOrderReq) (*v1.CancelOrderRes, error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	order, err := service.Vip().CancelOrder(ctx, userID, req.ID)
	if err != nil {
		return nil, paymentStatusError(ctx, err)
	}
	return orderStatusResponse(order), nil
}

func orderStatusResponse(o vipModel.Order) *v1.OrderRes {
	return &v1.OrderRes{OrderID: o.ID, OrderNo: o.OrderNo, Status: o.Status, AmountFen: o.AmountFen}
}
func paymentStatusError(ctx context.Context, err error) error {
	g.Log().Errorf(ctx, "vip payment status failed: %v", err)
	return gerror.NewCode(gcode.CodeBusinessValidationFailed, "订单状态暂未确认，请稍后刷新，避免重复付款")
}

func (c *ControllerV1) CancelAutoRenew(ctx context.Context, _ *v1.CancelAutoRenewReq) (*v1.CancelAutoRenewRes, error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := service.Vip().CancelAutoRenew(ctx, userID); err != nil {
		return nil, toAPIError(err)
	}
	return &v1.CancelAutoRenewRes{Success: true}, nil
}

func currentUserID(ctx context.Context) (uint64, error) {
	value := service.Bizctx().Get(ctx)
	if value == nil || value.User == nil {
		return 0, gerror.NewCode(gcode.CodeNotAuthorized, "authorization is required")
	}
	return uint64(value.User.ID), nil
}

func toPlanAPI(item vipModel.Plan) v1.PlanInfo {
	return v1.PlanInfo{ID: item.ID, Code: item.Code, Name: item.Name, Description: item.Description, BillingCycle: item.BillingCycle, PlanType: item.PlanType, DurationDays: item.DurationDays, PriceFen: item.PriceFen, OriginalPriceFen: item.OriginalPriceFen, Currency: item.Currency, AutoRenew: item.AutoRenew, SortOrder: item.SortOrder, CreatedAt: item.CreatedAt}
}
func toAgreementAPI(item vipModel.Agreement) v1.AgreementInfo {
	return v1.AgreementInfo{ID: item.ID, Key: item.Key, Title: item.Title, Content: item.Content, Status: item.Status, Version: item.Version, UpdatedAt: item.UpdatedAt}
}
func toSubscriptionAPI(item vipModel.Subscription) *v1.SubscriptionInfo {
	return &v1.SubscriptionInfo{ID: item.ID, Plan: toPlanAPI(item.Plan), Status: item.Status, AutoRenew: item.AutoRenew, Provider: item.Provider, StartAt: item.StartAt, EndAt: item.EndAt, NextBillingAt: item.NextBillingAt, ProviderContract: item.ProviderContract}
}
func toAPIError(err error) error {
	if errors.Is(err, consts.ErrTokenInvalid) {
		return gerror.NewCode(gcode.CodeNotAuthorized, err.Error())
	}
	var p *consts.InvalidParameterError
	if errors.As(err, &p) {
		return gerror.NewCode(gcode.CodeInvalidParameter, p.Error())
	}
	return gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
}
