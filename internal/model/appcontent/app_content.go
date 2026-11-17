package appcontent

import "time"

const (
	PageKeyAboutUs = "about-us"
	PageKeyContact = "contact"
)

type PageContent struct {
	ID            uint64    `json:"id"`
	PageKey       string    `json:"page_key"`
	Title         string    `json:"title"`
	Summary       string    `json:"summary"`
	Body          string    `json:"body"`
	ContactPhone  string    `json:"contact_phone"`
	ContactEmail  string    `json:"contact_email"`
	ServiceHours  string    `json:"service_hours"`
	Address       string    `json:"address"`
	WechatAccount string    `json:"wechat_account"`
	Status        uint8     `json:"status"`
	UpdatedBy     *uint64   `json:"updated_by,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type UpdateInput struct {
	AdminUserID   uint64
	PageKey       string
	Title         string
	Summary       string
	Body          string
	ContactPhone  string
	ContactEmail  string
	ServiceHours  string
	Address       string
	WechatAccount string
	Status        uint8
}

type ListResult struct {
	Items []PageContent `json:"items"`
}
