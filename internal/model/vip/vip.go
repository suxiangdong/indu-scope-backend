package vip

import "time"

const (
	AgreementKeyService   = "service"
	AgreementKeyAutoRenew = "auto-renew"
	AgreementKeyUser      = "user"
	AgreementKeyPrivacy   = "privacy"
)

type Plan struct {
	ID               uint64
	Code             string
	Name             string
	Description      string
	BillingCycle     string
	PlanType         string
	DurationDays     uint
	PriceFen         uint
	OriginalPriceFen uint
	Currency         string
	AutoRenew        bool
	Status           uint8
	SortOrder        uint
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Subscription struct {
	ID               uint64
	UserID           uint64
	Plan             Plan
	Status           string
	AutoRenew        bool
	Provider         string
	ProviderContract string
	StartAt          *time.Time
	EndAt            *time.Time
	NextBillingAt    *time.Time
}

type Order struct {
	ID                        uint64
	OrderNo                   string
	UserID                    uint64
	UserCompany               string
	PlanID                    uint64
	PlanName                  string
	PlanCodeSnapshot          string
	PlanNameSnapshot          string
	PlanDescriptionSnapshot   string
	OriginalPriceFenSnapshot  uint
	PlanTypeSnapshot          string
	BillingCycleSnapshot      string
	DurationDaysSnapshot      uint
	AutoRenewEnabledSnapshot  bool
	SubscriptionID            *uint64
	AmountFen                 uint
	Currency                  string
	Status                    string
	PaymentMethod             string
	ProviderTransaction       string
	ServiceAgreementID        *uint64
	ServiceAgreementVersion   *uint
	AutoRenewAgreementID      *uint64
	AutoRenewAgreementVersion *uint
	CreatedAt                 time.Time
	PaidAt                    *time.Time
	EntitlementStartAt        *time.Time
	EntitlementEndAt          *time.Time
	LatestRefundID            *uint64
	RefundStatus              string
}

type PaymentParams struct {
	TimeStamp string `json:"time_stamp"`
	NonceStr  string `json:"nonce_str"`
	Package   string `json:"package"`
	SignType  string `json:"sign_type"`
	PaySign   string `json:"pay_sign"`
}

type Checkout struct {
	Order   Order
	Payment *PaymentParams
}

type PlanInput struct {
	Code             string
	Name             string
	Description      string
	BillingCycle     string
	PlanType         string
	DurationDays     uint
	PriceFen         uint
	OriginalPriceFen uint
	AutoRenewEnabled uint8
	SortOrder        uint
}

type Agreement struct {
	ID        uint64
	Key       string
	Title     string
	Content   string
	Status    uint8
	Version   uint
	UpdatedAt time.Time
}

type AgreementInput struct {
	Key         string
	Title       string
	Content     string
	Status      uint8
	AdminUserID uint64
}

type OrderListFilter struct {
	Status   *string
	OrderNo  string
	Page     uint
	PageSize uint
}

type OrderListResult struct {
	Items    []Order
	Total    uint64
	Page     uint
	PageSize uint
}
