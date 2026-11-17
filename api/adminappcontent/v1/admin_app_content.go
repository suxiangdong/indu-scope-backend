package v1

import "github.com/gogf/gf/v2/frame/g"

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

type ListReq struct {
	g.Meta `path:"/api/v1/admin/app-pages" tags:"AdminAppContent" method:"get" summary:"List app page contents"`
}
type ListRes struct {
	g.Meta `mime:"application/json"`
	Items  []PageContent `json:"items"`
}

type UpdateReq struct {
	g.Meta        `path:"/api/v1/admin/app-pages/update" tags:"AdminAppContent" method:"post" summary:"Update app page content"`
	PageKey       string `json:"page_key" v:"required"`
	Title         string `json:"title" v:"required"`
	Summary       string `json:"summary"`
	Body          string `json:"body" v:"required"`
	ContactPhone  string `json:"contact_phone"`
	ContactEmail  string `json:"contact_email"`
	ServiceHours  string `json:"service_hours"`
	Address       string `json:"address"`
	WechatAccount string `json:"wechat_account"`
	Status        *uint8 `json:"status"`
}
type UpdateRes struct {
	g.Meta `mime:"application/json"`
	PageContent
}
