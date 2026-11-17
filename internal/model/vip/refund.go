package vip

import "time"

type Refund struct {
	ID                 uint64     `json:"id"`
	RefundNo           string     `json:"refund_no"`
	OrderID            uint64     `json:"order_id"`
	OrderNo            string     `json:"order_no"`
	UserID             uint64     `json:"user_id"`
	UserCompany        string     `json:"user_company,omitempty"`
	PlanName           string     `json:"plan_name,omitempty"`
	RequestedByAdminID *uint64    `json:"requested_by_admin_id,omitempty"`
	AdminName          string     `json:"admin_name,omitempty"`
	Reason             string     `json:"reason"`
	AmountFen          uint       `json:"amount_fen"`
	TotalFen           uint       `json:"total_fen"`
	Currency           string     `json:"currency"`
	Status             string     `json:"status"`
	ProviderRefundID   string     `json:"provider_refund_id,omitempty"`
	PayerRefundFen     *uint      `json:"payer_refund_fen,omitempty"`
	ReceivedAccount    string     `json:"received_account,omitempty"`
	EntitlementStatus  string     `json:"entitlement_status"`
	ReclaimedSeconds   uint64     `json:"reclaimed_seconds"`
	LastError          string     `json:"last_error,omitempty"`
	SuccessAt          *time.Time `json:"success_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type RefundListFilter struct {
	OrderID                   uint64
	OrderNo, RefundNo, Status string
	Page, PageSize            uint
}
type RefundListResult struct {
	Items    []Refund `json:"items"`
	Total    uint64   `json:"total"`
	Page     uint     `json:"page"`
	PageSize uint     `json:"page_size"`
}
