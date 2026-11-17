package vip

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	vipModel "github.com/suxiangdong/indu-scope-backend/internal/model/vip"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
	"github.com/suxiangdong/indu-scope-backend/internal/wechatpay"
)

type sVip struct {
	paymentDatabase gdb.DB
	refundGateway   refundGateway
	refundConfig    wechatpay.Config
}

const maxMembershipDurationDays uint = 36500

var _ service.IVip = (*sVip)(nil)

func New() service.IVip { return &sVip{} }

func init() { service.RegisterVip(New()) }

func (s *sVip) ListPlans(ctx context.Context, activeOnly bool) ([]vipModel.Plan, error) {
	query := s.paymentDB().Model("vip_plans").Ctx(ctx)
	if activeOnly {
		// The app catalog must only advertise purchasable plans. Recurring
		// contracts are not available until their signing flow is implemented.
		query = query.Where("status", 1).Where("plan_type", "normal")
	}
	var rows []planRow
	if err := query.OrderAsc("sort_order").OrderAsc("id").Scan(&rows); err != nil {
		return nil, gerror.Wrap(err, "list vip plans")
	}
	items := make([]vipModel.Plan, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.plan())
	}
	return items, nil
}

func (s *sVip) GetSubscription(ctx context.Context, userID uint64) (*vipModel.Subscription, error) {
	if userID == 0 {
		return nil, consts.ErrTokenInvalid
	}
	var row *subscriptionRow
	err := s.paymentDB().Model("vip_subscriptions s").Ctx(ctx).
		LeftJoin("vip_plans p", "p.id = s.plan_id").
		Fields("s.id, s.user_id, s.status, s.auto_renew, s.provider, s.provider_contract_id, s.start_at, s.end_at, s.next_billing_at, s.plan_id, p.code AS plan_code, p.name AS plan_name, p.description AS plan_description, p.billing_cycle AS plan_billing_cycle, s.plan_type_snapshot AS plan_plan_type, p.duration_days AS plan_duration_days, p.price_fen AS plan_price_fen, s.currency_snapshot AS plan_currency, (s.plan_type_snapshot = 'auto_renew') AS plan_auto_renew, p.status AS plan_status, p.sort_order AS plan_sort_order, p.created_at AS plan_created_at, p.updated_at AS plan_updated_at").
		Where("s.user_id", userID).Where("s.status", "active").Where("(s.end_at IS NULL OR s.end_at > NOW())").
		Order("s.end_at IS NULL DESC, s.end_at DESC, s.id DESC").Limit(1).Scan(&row)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && row == nil) {
		return nil, nil
	}
	if err != nil {
		return nil, gerror.Wrap(err, "get vip subscription")
	}
	item := row.subscription()
	return &item, nil
}

func (s *sVip) CancelAutoRenew(ctx context.Context, userID uint64) error {
	if userID == 0 {
		return consts.ErrTokenInvalid
	}
	result, err := g.DB().Model("vip_subscriptions").Where("user_id", userID).Where("status", "active").Data(g.Map{"auto_renew": 0, "cancel_reason": "user requested cancellation"}).Update()
	if err != nil {
		return gerror.Wrap(err, "cancel vip auto renewal")
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return gerror.New("active vip subscription not found")
	}
	return nil
}

func (s *sVip) CreatePlan(ctx context.Context, input vipModel.PlanInput) (vipModel.Plan, error) {
	input = normalizePlanInput(input)
	input.Code = generatePlanCode()
	if err := validatePlanInput(input); err != nil {
		return vipModel.Plan{}, err
	}
	if count, err := g.DB().Model("vip_plans").Where("code", input.Code).Count(); err != nil {
		return vipModel.Plan{}, gerror.Wrap(err, "check vip plan code")
	} else if count > 0 {
		return vipModel.Plan{}, gerror.New("vip plan code already exists")
	}
	id, err := g.DB().Model("vip_plans").Data(planData(input)).InsertAndGetId()
	if err != nil {
		return vipModel.Plan{}, gerror.Wrap(err, "create vip plan")
	}
	return s.loadPlan(ctx, uint64(id))
}

func (s *sVip) UpdatePlan(ctx context.Context, id uint64, input vipModel.PlanInput) (vipModel.Plan, error) {
	if id == 0 {
		return vipModel.Plan{}, &consts.InvalidParameterError{Message: "id is required"}
	}
	var existing planRow
	if err := g.DB().Model("vip_plans").Where("id", id).Scan(&existing); err != nil {
		return vipModel.Plan{}, gerror.Wrap(err, "query vip plan")
	}
	if existing.ID == 0 {
		return vipModel.Plan{}, gerror.New("vip plan not found")
	}
	input = normalizePlanInput(input)
	input.Code = existing.Code
	if err := validatePlanInput(input); err != nil {
		return vipModel.Plan{}, err
	}
	if _, err := g.DB().Model("vip_plans").Where("id", id).Data(planData(input)).Update(); err != nil {
		return vipModel.Plan{}, gerror.Wrap(err, "update vip plan")
	}
	return s.loadPlan(ctx, id)
}

func (s *sVip) UpdatePlanStatus(ctx context.Context, id uint64, status uint8) (vipModel.Plan, error) {
	if status > 1 {
		return vipModel.Plan{}, &consts.InvalidParameterError{Message: "status must be 0 or 1"}
	}
	if _, err := g.DB().Model("vip_plans").Where("id", id).Data(g.Map{"status": status}).Update(); err != nil {
		return vipModel.Plan{}, gerror.Wrap(err, "update vip plan status")
	}
	return s.loadPlan(ctx, id)
}

func (s *sVip) ListOrders(ctx context.Context, filter vipModel.OrderListFilter) (vipModel.OrderListResult, error) {
	if filter.Page == 0 {
		filter.Page = 1
	}
	if filter.PageSize == 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	query := s.paymentDB().Model("payment_orders o").Ctx(ctx).LeftJoin("vip_plans p", "p.id = o.plan_id").LeftJoin("users u", "u.id = o.user_id")
	if filter.Status != nil && strings.TrimSpace(*filter.Status) != "" {
		query = query.Where("o.status", strings.TrimSpace(*filter.Status))
	}
	if filter.OrderNo != "" {
		query = query.Where("o.order_no", strings.TrimSpace(filter.OrderNo))
	}
	total, err := query.Count()
	if err != nil {
		return vipModel.OrderListResult{}, gerror.Wrap(err, "count vip orders")
	}
	var rows []orderRow
	if err := query.Fields("o.*, p.name AS plan_name, u.company_name AS user_company, (SELECT r.id FROM payment_refunds r WHERE r.order_id = o.id ORDER BY r.id DESC LIMIT 1) AS latest_refund_id, (SELECT r.status FROM payment_refunds r WHERE r.order_id = o.id ORDER BY r.id DESC LIMIT 1) AS refund_status").OrderDesc("o.created_at").Limit(int((filter.Page-1)*filter.PageSize), int(filter.PageSize)).Scan(&rows); err != nil {
		return vipModel.OrderListResult{}, gerror.Wrap(err, "list vip orders")
	}
	items := make([]vipModel.Order, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.order())
	}
	return vipModel.OrderListResult{Items: items, Total: uint64(total), Page: filter.Page, PageSize: filter.PageSize}, nil
}

func (s *sVip) loadPlan(ctx context.Context, id uint64) (vipModel.Plan, error) {
	var row planRow
	if err := g.DB().Model("vip_plans").Where("id", id).Scan(&row); err != nil {
		return vipModel.Plan{}, gerror.Wrap(err, "get vip plan")
	}
	if row.ID == 0 {
		return vipModel.Plan{}, gerror.New("vip plan not found")
	}
	return row.plan(), nil
}

type planRow struct {
	ID               uint64    `orm:"id"`
	Code             string    `orm:"code"`
	Name             string    `orm:"name"`
	Description      string    `orm:"description"`
	BillingCycle     string    `orm:"billing_cycle"`
	PlanType         string    `orm:"plan_type"`
	DurationDays     uint      `orm:"duration_days"`
	PriceFen         uint      `orm:"price_fen"`
	OriginalPriceFen uint      `orm:"original_price_fen"`
	Currency         string    `orm:"currency"`
	AutoRenewEnabled uint8     `orm:"auto_renew_enabled"`
	Status           uint8     `orm:"status"`
	SortOrder        uint      `orm:"sort_order"`
	CreatedAt        time.Time `orm:"created_at"`
	UpdatedAt        time.Time `orm:"updated_at"`
}

func (r planRow) plan() vipModel.Plan {
	return vipModel.Plan{ID: r.ID, Code: r.Code, Name: r.Name, Description: r.Description, BillingCycle: r.BillingCycle, PlanType: r.PlanType, DurationDays: r.DurationDays, PriceFen: r.PriceFen, OriginalPriceFen: r.OriginalPriceFen, Currency: r.Currency, AutoRenew: r.AutoRenewEnabled == 1, Status: r.Status, SortOrder: r.SortOrder, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

type subscriptionRow struct {
	ID                 uint64     `orm:"id"`
	UserID             uint64     `orm:"user_id"`
	Status             string     `orm:"status"`
	AutoRenew          uint8      `orm:"auto_renew"`
	Provider           string     `orm:"provider"`
	ProviderContractID string     `orm:"provider_contract_id"`
	StartAt            *time.Time `orm:"start_at"`
	EndAt              *time.Time `orm:"end_at"`
	NextBillingAt      *time.Time `orm:"next_billing_at"`
	PlanID             uint64     `orm:"plan_id"`
	PlanCode           string     `orm:"plan_code"`
	PlanName           string     `orm:"plan_name"`
	PlanDescription    string     `orm:"plan_description"`
	PlanBillingCycle   string     `orm:"plan_billing_cycle"`
	PlanType           string     `orm:"plan_plan_type"`
	PlanDurationDays   uint       `orm:"plan_duration_days"`
	PlanPriceFen       uint       `orm:"plan_price_fen"`
	PlanCurrency       string     `orm:"plan_currency"`
	PlanAutoRenew      uint8      `orm:"plan_auto_renew"`
	PlanStatus         uint8      `orm:"plan_status"`
	PlanSortOrder      uint       `orm:"plan_sort_order"`
	PlanCreatedAt      time.Time  `orm:"plan_created_at"`
	PlanUpdatedAt      time.Time  `orm:"plan_updated_at"`
}

func (r subscriptionRow) subscription() vipModel.Subscription {
	return vipModel.Subscription{ID: r.ID, UserID: r.UserID, Status: r.Status, AutoRenew: r.AutoRenew == 1, Provider: r.Provider, ProviderContract: r.ProviderContractID, StartAt: r.StartAt, EndAt: r.EndAt, NextBillingAt: r.NextBillingAt, Plan: vipModel.Plan{ID: r.PlanID, Code: r.PlanCode, Name: r.PlanName, Description: r.PlanDescription, BillingCycle: r.PlanBillingCycle, PlanType: r.PlanType, DurationDays: r.PlanDurationDays, PriceFen: r.PlanPriceFen, Currency: r.PlanCurrency, AutoRenew: r.PlanAutoRenew == 1, Status: r.PlanStatus, SortOrder: r.PlanSortOrder, CreatedAt: r.PlanCreatedAt, UpdatedAt: r.PlanUpdatedAt}}
}

type orderRow struct {
	ID                        uint64     `orm:"id"`
	OrderNo                   string     `orm:"order_no"`
	UserID                    uint64     `orm:"user_id"`
	UserCompany               string     `orm:"user_company"`
	PlanID                    uint64     `orm:"plan_id"`
	PlanName                  string     `orm:"plan_name"`
	PlanCodeSnapshot          string     `orm:"plan_code_snapshot"`
	PlanNameSnapshot          string     `orm:"plan_name_snapshot"`
	PlanDescriptionSnapshot   string     `orm:"plan_description_snapshot"`
	OriginalPriceFenSnapshot  uint       `orm:"original_price_fen_snapshot"`
	PlanTypeSnapshot          string     `orm:"plan_type_snapshot"`
	BillingCycleSnapshot      string     `orm:"billing_cycle_snapshot"`
	DurationDaysSnapshot      uint       `orm:"duration_days_snapshot"`
	AutoRenewEnabledSnapshot  uint8      `orm:"auto_renew_enabled_snapshot"`
	SubscriptionID            *uint64    `orm:"subscription_id"`
	AmountFen                 uint       `orm:"amount_fen"`
	Currency                  string     `orm:"currency"`
	Status                    string     `orm:"status"`
	PaymentMethod             string     `orm:"payment_method"`
	ProviderTransactionID     string     `orm:"provider_transaction_id"`
	ServiceAgreementID        *uint64    `orm:"service_agreement_id"`
	ServiceAgreementVersion   *uint      `orm:"service_agreement_version"`
	AutoRenewAgreementID      *uint64    `orm:"auto_renew_agreement_id"`
	AutoRenewAgreementVersion *uint      `orm:"auto_renew_agreement_version"`
	CreatedAt                 time.Time  `orm:"created_at"`
	PaidAt                    *time.Time `orm:"paid_at"`
	EntitlementStartAt        *time.Time `orm:"entitlement_start_at"`
	EntitlementEndAt          *time.Time `orm:"entitlement_end_at"`
	EntitlementRevokedAt      *time.Time `orm:"entitlement_revoked_at"`
	LatestRefundID            *uint64    `orm:"latest_refund_id"`
	RefundStatus              string     `orm:"refund_status"`
	PrepayID                  string     `orm:"prepay_id"`
	ExpiresAt                 *time.Time `orm:"expires_at"`
	PaymentAppID              string     `orm:"payment_app_id"`
	PaymentMchID              string     `orm:"payment_mch_id"`
	PayerOpenID               string     `orm:"payer_openid"`
	Provider                  string     `orm:"provider"`
}

func (r orderRow) order() vipModel.Order {
	planName := r.PlanName
	if r.PlanNameSnapshot != "" {
		planName = r.PlanNameSnapshot
	}
	return vipModel.Order{ID: r.ID, OrderNo: r.OrderNo, UserID: r.UserID, UserCompany: r.UserCompany, PlanID: r.PlanID, PlanName: planName, PlanCodeSnapshot: r.PlanCodeSnapshot, PlanNameSnapshot: r.PlanNameSnapshot, PlanDescriptionSnapshot: r.PlanDescriptionSnapshot, OriginalPriceFenSnapshot: r.OriginalPriceFenSnapshot, PlanTypeSnapshot: r.PlanTypeSnapshot, BillingCycleSnapshot: r.BillingCycleSnapshot, DurationDaysSnapshot: r.DurationDaysSnapshot, AutoRenewEnabledSnapshot: r.AutoRenewEnabledSnapshot == 1, SubscriptionID: r.SubscriptionID, AmountFen: r.AmountFen, Currency: r.Currency, Status: r.Status, PaymentMethod: r.PaymentMethod, ProviderTransaction: r.ProviderTransactionID, EntitlementStartAt: r.EntitlementStartAt, EntitlementEndAt: r.EntitlementEndAt, LatestRefundID: r.LatestRefundID, RefundStatus: r.RefundStatus, ServiceAgreementID: r.ServiceAgreementID, ServiceAgreementVersion: r.ServiceAgreementVersion, AutoRenewAgreementID: r.AutoRenewAgreementID, AutoRenewAgreementVersion: r.AutoRenewAgreementVersion, CreatedAt: r.CreatedAt, PaidAt: r.PaidAt}
}

func normalizePlanInput(input vipModel.PlanInput) vipModel.PlanInput {
	input.Code = strings.TrimSpace(input.Code)
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.PlanType = strings.TrimSpace(input.PlanType)
	if input.PlanType == "" {
		if input.AutoRenewEnabled == 1 {
			input.PlanType = "auto_renew"
		} else {
			input.PlanType = "normal"
		}
	}
	if input.AutoRenewEnabled > 1 {
		input.AutoRenewEnabled = 0
	}
	input.AutoRenewEnabled = 0
	if input.PlanType == "auto_renew" {
		input.AutoRenewEnabled = 1
	}
	return input
}

func generatePlanCode() string {
	return fmt.Sprintf("vip_%d", time.Now().UnixNano())
}
func validatePlanInput(input vipModel.PlanInput) error {
	if input.Code == "" || len([]rune(input.Code)) > 64 {
		return &consts.InvalidParameterError{Message: "code must contain 1 to 64 characters"}
	}
	if input.Name == "" || len([]rune(input.Name)) > 100 {
		return &consts.InvalidParameterError{Message: "name must contain 1 to 100 characters"}
	}
	if input.DurationDays == 0 || input.DurationDays > maxMembershipDurationDays {
		return &consts.InvalidParameterError{Message: "duration_days must be between 1 and 36500"}
	}
	if input.PlanType != "normal" && input.PlanType != "auto_renew" {
		return &consts.InvalidParameterError{Message: "plan_type is invalid"}
	}
	if input.OriginalPriceFen != 0 && input.OriginalPriceFen < input.PriceFen {
		return &consts.InvalidParameterError{Message: "original_price_fen must be greater than or equal to price_fen"}
	}
	switch input.BillingCycle {
	case "month", "quarter", "year", "custom":
	default:
		return &consts.InvalidParameterError{Message: "billing_cycle is invalid"}
	}
	return nil
}
func planData(input vipModel.PlanInput) g.Map {
	return g.Map{"code": input.Code, "name": input.Name, "description": input.Description, "billing_cycle": input.BillingCycle, "plan_type": input.PlanType, "duration_days": input.DurationDays, "price_fen": input.PriceFen, "original_price_fen": input.OriginalPriceFen, "currency": "CNY", "auto_renew_enabled": input.AutoRenewEnabled, "sort_order": input.SortOrder}
}
func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.TrimSpace(value)
}

func uint64Ptr(value uint64) *uint64 { return &value }
func uintPtr(value uint) *uint       { return &value }

func agreementID(value *vipModel.Agreement) any {
	if value == nil {
		return nil
	}
	return value.ID
}

func agreementVersion(value *vipModel.Agreement) any {
	if value == nil {
		return nil
	}
	return value.Version
}

func agreementIDPtr(value *vipModel.Agreement) *uint64 {
	if value == nil {
		return nil
	}
	return uint64Ptr(value.ID)
}

func agreementVersionPtr(value *vipModel.Agreement) *uint {
	if value == nil {
		return nil
	}
	return uintPtr(value.Version)
}
