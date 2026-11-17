package v1

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type PageContent struct {
	ID            uint64 `json:"id"`
	PageKey       string `json:"page_key"`
	Title         string `json:"title"`
	Summary       string `json:"summary"`
	Body          string `json:"body"`
	ContactPhone  string `json:"contact_phone"`
	ContactEmail  string `json:"contact_email"`
	ServiceHours  string `json:"service_hours"`
	Address       string `json:"address"`
	WechatAccount string `json:"wechat_account"`
	Status        uint8  `json:"status"`
}

type GetReq struct {
	g.Meta  `path:"/api/v1/app-pages/{page_key}" tags:"AppContent" method:"get" summary:"Get app page content"`
	PageKey string `json:"page_key" in:"path" v:"required"`
}

type GetRes struct {
	g.Meta `mime:"application/json"`
	PageContent
}

type AgreementContent struct {
	ID        uint64    `json:"id"`
	Key       string    `json:"key"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Status    uint8     `json:"status"`
	Version   uint      `json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AgreementsReq struct {
	g.Meta `path:"/api/v1/app-agreements" tags:"AppContent" method:"get" summary:"List current user agreement and privacy policy"`
}

type AgreementsRes struct {
	g.Meta `mime:"application/json"`
	Items  []AgreementContent `json:"items"`
}
