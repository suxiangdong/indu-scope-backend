package v1

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
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
	SortOrder        uint      `json:"sort_order"`
	CreatedAt        time.Time `json:"created_at"`
}

type SubscriptionInfo struct {
	ID               uint64     `json:"id"`
	Plan             PlanInfo   `json:"plan"`
	Status           string     `json:"status"`
	AutoRenew        bool       `json:"auto_renew"`
	Provider         string     `json:"provider"`
	StartAt          *time.Time `json:"start_at,omitempty"`
	EndAt            *time.Time `json:"end_at,omitempty"`
	NextBillingAt    *time.Time `json:"next_billing_at,omitempty"`
	ProviderContract string     `json:"provider_contract_id,omitempty"`
}

type PlansReq struct {
	g.Meta `path:"/api/v1/vip/plans" tags:"Vip" method:"get" summary:"List active VIP plans"`
}

type PlansRes struct {
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

type AgreementsReq struct {
	g.Meta `path:"/api/v1/vip/agreements" tags:"Vip" method:"get" summary:"List active VIP agreements"`
}

type AgreementsRes struct {
	g.Meta `mime:"application/json"`
	Items  []AgreementInfo `json:"items"`
}

type MeReq struct {
	g.Meta `path:"/api/v1/vip/me" tags:"Vip" method:"get" summary:"Get current VIP subscription"`
}

type MeRes struct {
	g.Meta       `mime:"application/json"`
	Subscription *SubscriptionInfo `json:"subscription,omitempty"`
}

type CreateOrderReq struct {
	g.Meta      `path:"/api/v1/vip/orders" tags:"Vip" method:"post" summary:"Create VIP payment order"`
	PlanID      uint64 `json:"plan_id" v:"required"`
	AutoRenew   bool   `json:"auto_renew"`
	Idempotency string `json:"idempotency_key"`
}

type CreateOrderRes struct {
	g.Meta                    `mime:"application/json"`
	OrderID                   uint64         `json:"order_id"`
	OrderNo                   string         `json:"order_no"`
	AmountFen                 uint           `json:"amount_fen"`
	Currency                  string         `json:"currency"`
	PaymentReady              bool           `json:"payment_ready"`
	Activated                 bool           `json:"activated"`
	Payment                   *PaymentParams `json:"payment,omitempty"`
	Message                   string         `json:"message,omitempty"`
	ServiceAgreementVersion   *uint          `json:"service_agreement_version,omitempty"`
	AutoRenewAgreementVersion *uint          `json:"auto_renew_agreement_version,omitempty"`
}

type PaymentParams struct {
	TimeStamp string `json:"time_stamp"`
	NonceStr  string `json:"nonce_str"`
	Package   string `json:"package"`
	SignType  string `json:"sign_type"`
	PaySign   string `json:"pay_sign"`
}

type OrderReq struct {
	g.Meta `path:"/api/v1/vip/orders/{id}" tags:"Vip" method:"get" summary:"Get and reconcile own VIP order"`
	ID     uint64 `json:"id" v:"required|min:1"`
}

type OrderRes struct {
	g.Meta    `mime:"application/json"`
	OrderID   uint64 `json:"order_id"`
	OrderNo   string `json:"order_no"`
	Status    string `json:"status"`
	AmountFen uint   `json:"amount_fen"`
}

type PayOrderReq struct {
	g.Meta `path:"/api/v1/vip/orders/{id}/pay" tags:"Vip" method:"post" summary:"Resume own VIP order payment"`
	ID     uint64 `json:"id" v:"required|min:1"`
}
type PayOrderRes = CreateOrderRes

type CancelOrderReq struct {
	g.Meta `path:"/api/v1/vip/orders/{id}/close" tags:"Vip" method:"post" summary:"Close own unpaid VIP order"`
	ID     uint64 `json:"id" v:"required|min:1"`
}
type CancelOrderRes = OrderRes

type CancelAutoRenewReq struct {
	g.Meta `path:"/api/v1/vip/subscription/cancel-auto-renew" tags:"Vip" method:"post" summary:"Cancel VIP auto renewal"`
}

type CancelAutoRenewRes struct {
	g.Meta  `mime:"application/json"`
	Success bool `json:"success"`
}
