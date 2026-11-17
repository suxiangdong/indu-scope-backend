package v1

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
	vipModel "github.com/suxiangdong/indu-scope-backend/internal/model/vip"
)

type PlanInfo struct {
	ID               uint64    `json:"id"`
	Code             string    `json:"code"`
	Name             string    `json:"name"`
	Description      string    `json:"description,omitempty"`
	BillingCycle     string    `json:"billing_cycle"`
	PlanType         string    `json:"plan_type"`
	DurationDays     uint      `json:"duration_days"`
	PriceFen         uint      `json:"price_fen"`
	OriginalPriceFen uint      `json:"original_price_fen"`
	Currency         string    `json:"currency"`
	AutoRenew        bool      `json:"auto_renew_enabled"`
	Status           uint8     `json:"status"`
	SortOrder        uint      `json:"sort_order"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type ListPlansReq struct {
	g.Meta `path:"/api/v1/admin/vip/plans" tags:"AdminVip" method:"get" summary:"List VIP plans"`
	Status *uint8 `json:"status" in:"query"`
}

type ListPlansRes struct {
	g.Meta `mime:"application/json"`
	Items  []PlanInfo `json:"items"`
}

type AgreementInfo struct {
	ID        uint64    `json:"id"`
	Key       string    `json:"key"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Status    uint8     `json:"status"`
	Version   uint      `json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListAgreementsReq struct {
	g.Meta `path:"/api/v1/admin/vip/agreements" tags:"AdminVip" method:"get" summary:"List VIP agreements"`
}

type ListAgreementsRes struct {
	g.Meta `mime:"application/json"`
	Items  []AgreementInfo `json:"items"`
}

type UpdateAgreementReq struct {
	g.Meta  `path:"/api/v1/admin/vip/agreements/update" tags:"AdminVip" method:"post" summary:"Update VIP agreement"`
	Key     string `json:"key" v:"required|in:service,auto-renew"`
	Title   string `json:"title" v:"required"`
	Content string `json:"content" v:"required"`
	Status  *uint8 `json:"status"`
}

type UpdateAgreementRes struct {
	g.Meta `mime:"application/json"`
	AgreementInfo
}

type CreatePlanReq struct {
	g.Meta           `path:"/api/v1/admin/vip/plans/create" tags:"AdminVip" method:"post" summary:"Create VIP plan"`
	Name             string `json:"name" v:"required"`
	Description      string `json:"description"`
	BillingCycle     string `json:"billing_cycle" v:"required|in:month,quarter,year,custom"`
	PlanType         string `json:"plan_type" v:"required|in:normal,auto_renew"`
	DurationDays     uint   `json:"duration_days" v:"required"`
	PriceFen         uint   `json:"price_fen"`
	OriginalPriceFen uint   `json:"original_price_fen"`
	AutoRenewEnabled *uint8 `json:"auto_renew_enabled"`
	SortOrder        uint   `json:"sort_order"`
}

type UpdatePlanReq struct {
	g.Meta           `path:"/api/v1/admin/vip/plans/update" tags:"AdminVip" method:"post" summary:"Update VIP plan"`
	ID               uint64 `json:"id" v:"required"`
	Name             string `json:"name" v:"required"`
	Description      string `json:"description"`
	BillingCycle     string `json:"billing_cycle" v:"required|in:month,quarter,year,custom"`
	PlanType         string `json:"plan_type" v:"required|in:normal,auto_renew"`
	DurationDays     uint   `json:"duration_days" v:"required"`
	PriceFen         uint   `json:"price_fen"`
	OriginalPriceFen uint   `json:"original_price_fen"`
	AutoRenewEnabled *uint8 `json:"auto_renew_enabled"`
	SortOrder        uint   `json:"sort_order"`
}

type UpdatePlanStatusReq struct {
	g.Meta `path:"/api/v1/admin/vip/plans/status" tags:"AdminVip" method:"post" summary:"Update VIP plan status"`
	ID     uint64 `json:"id" v:"required"`
	Status uint8  `json:"status"`
}

type PlanRes struct {
	g.Meta `mime:"application/json"`
	PlanInfo
}

type OrderInfo struct {
	ID                        uint64     `json:"id"`
	OrderNo                   string     `json:"order_no"`
	UserID                    uint64     `json:"user_id"`
	UserCompany               string     `json:"user_company,omitempty"`
	PlanName                  string     `json:"plan_name"`
	PlanCodeSnapshot          string     `json:"plan_code_snapshot,omitempty"`
	PlanNameSnapshot          string     `json:"plan_name_snapshot,omitempty"`
	PlanDescriptionSnapshot   string     `json:"plan_description_snapshot,omitempty"`
	OriginalPriceFenSnapshot  uint       `json:"original_price_fen_snapshot,omitempty"`
	PlanTypeSnapshot          string     `json:"plan_type_snapshot,omitempty"`
	BillingCycleSnapshot      string     `json:"billing_cycle_snapshot,omitempty"`
	DurationDaysSnapshot      uint       `json:"duration_days_snapshot,omitempty"`
	AutoRenewEnabledSnapshot  bool       `json:"auto_renew_enabled_snapshot"`
	AmountFen                 uint       `json:"amount_fen"`
	Currency                  string     `json:"currency"`
	Status                    string     `json:"status"`
	PaymentMethod             string     `json:"payment_method"`
	ProviderTransaction       string     `json:"provider_transaction_id,omitempty"`
	ServiceAgreementID        *uint64    `json:"service_agreement_id,omitempty"`
	ServiceAgreementVersion   *uint      `json:"service_agreement_version,omitempty"`
	AutoRenewAgreementID      *uint64    `json:"auto_renew_agreement_id,omitempty"`
	AutoRenewAgreementVersion *uint      `json:"auto_renew_agreement_version,omitempty"`
	CreatedAt                 time.Time  `json:"created_at"`
	PaidAt                    *time.Time `json:"paid_at,omitempty"`
	EntitlementStartAt        *time.Time `json:"entitlement_start_at,omitempty"`
	EntitlementEndAt          *time.Time `json:"entitlement_end_at,omitempty"`
	LatestRefundID            *uint64    `json:"latest_refund_id,omitempty"`
	RefundStatus              string     `json:"refund_status,omitempty"`
}

type ListOrdersReq struct {
	g.Meta   `path:"/api/v1/admin/vip/orders" tags:"AdminVip" method:"get" summary:"List VIP payment orders"`
	Status   *string `json:"status" in:"query"`
	OrderNo  string  `json:"order_no" in:"query"`
	Page     uint    `json:"page" in:"query"`
	PageSize uint    `json:"page_size" in:"query"`
}

type ListOrdersRes struct {
	g.Meta   `mime:"application/json"`
	Items    []OrderInfo `json:"items"`
	Total    uint64      `json:"total"`
	Page     uint        `json:"page"`
	PageSize uint        `json:"page_size"`
}

type CreateRefundReq struct {
	g.Meta `path:"/api/v1/admin/vip/orders/{id}/refund" tags:"AdminVip" method:"post" summary:"Create a full VIP order refund"`
	ID     uint64 `json:"id" v:"required|min:1"`
	Reason string `json:"reason" v:"required"`
}
type RefundReq struct {
	g.Meta `path:"/api/v1/admin/vip/refunds/{id}" tags:"AdminVip" method:"get" summary:"Query and reconcile a VIP refund"`
	ID     uint64 `json:"id" v:"required|min:1"`
}
type RetryRefundReq struct {
	g.Meta `path:"/api/v1/admin/vip/refunds/{id}/retry" tags:"AdminVip" method:"post" summary:"Retry an unconfirmed refund with the same refund number"`
	ID     uint64 `json:"id" v:"required|min:1"`
}
type RefundRes struct {
	g.Meta `mime:"application/json"`
	vipModel.Refund
}
type ListRefundsReq struct {
	g.Meta   `path:"/api/v1/admin/vip/refunds" tags:"AdminVip" method:"get" summary:"List VIP refunds"`
	OrderID  uint64 `json:"order_id" in:"query"`
	OrderNo  string `json:"order_no" in:"query"`
	RefundNo string `json:"refund_no" in:"query"`
	Status   string `json:"status" in:"query" v:"in:created,processing,success,closed,abnormal"`
	Page     uint   `json:"page" in:"query"`
	PageSize uint   `json:"page_size" in:"query"`
}
type ListRefundsRes struct {
	g.Meta `mime:"application/json"`
	vipModel.RefundListResult
}
