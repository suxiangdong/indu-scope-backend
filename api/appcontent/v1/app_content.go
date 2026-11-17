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

type GetReq struct {
	g.Meta  `path:"/api/v1/app-pages/{page_key}" tags:"AppContent" method:"get" summary:"Get app page content"`
	PageKey string `json:"page_key" in:"path" v:"required"`
}

type GetRes struct {
	g.Meta `mime:"application/json"`
	PageContent
}
