package v1

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type CustomerUserInfo struct {
	ID                 int64     `json:"id"`
	Phone              string    `json:"phone"`
	Role               uint8     `json:"role"`
	UserType           uint8     `json:"user_type"`
	Status             uint8     `json:"status"`
	VerificationType   string    `json:"verification_type,omitempty"`
	VerificationStatus string    `json:"verification_status"`
	WechatBound        bool      `json:"wechat_bound"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type ListReq struct {
	g.Meta   `path:"/api/v1/admin/customer-users" tags:"AdminCustomerUsers" method:"get" summary:"List frontend users"`
	Phone    string `json:"phone" in:"query"`
	Role     *uint8 `json:"role" in:"query"`
	UserType *uint8 `json:"user_type" in:"query"`
	Status   *uint8 `json:"status" in:"query"`
}

type ListRes struct {
	g.Meta `mime:"application/json"`
	Items  []CustomerUserInfo `json:"items"`
}

type UpdateStatusReq struct {
	g.Meta `path:"/api/v1/admin/customer-users/status" tags:"AdminCustomerUsers" method:"post" summary:"Update frontend user status"`
	ID     uint64 `json:"id" v:"required"`
	Status uint8  `json:"status" dc:"0 unverified, 1 verified, 2 disabled"`
}

type UpdateStatusRes struct {
	g.Meta `mime:"application/json"`
	CustomerUserInfo
}
