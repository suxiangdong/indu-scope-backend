package v1

import "github.com/gogf/gf/v2/frame/g"

type LoginReq struct {
	g.Meta   `path:"/api/v1/admin/auth/login" tags:"AdminAuth" method:"post" summary:"Admin login"`
	Username string `json:"username" v:"required" dc:"Admin username or phone"`
	Password string `json:"password" v:"required" dc:"Admin password"`
}

type AdminUserInfo struct {
	ID           int64  `json:"id" dc:"Admin user ID"`
	Username     string `json:"username" dc:"Admin username"`
	Name         string `json:"name" dc:"Admin name"`
	Phone        string `json:"phone" dc:"Phone number"`
	Status       uint8  `json:"status" dc:"0 disabled, 1 active"`
	IsSuperAdmin bool   `json:"is_super_admin" dc:"Whether this administrator can perform super-admin operations"`
}

type AuthData struct {
	AccessToken string        `json:"access_token" dc:"JWT access token"`
	TokenType   string        `json:"token_type" example:"Bearer"`
	User        AdminUserInfo `json:"user"`
}

type LoginRes struct {
	g.Meta `mime:"application/json"`
	AuthData
}

type MeReq struct {
	g.Meta `path:"/api/v1/admin/me" tags:"AdminAuth" method:"get" summary:"Get current admin user"`
}

type MeRes struct {
	g.Meta `mime:"application/json"`
	AdminUserInfo
}

type LogoutReq struct {
	g.Meta `path:"/api/v1/admin/auth/logout" tags:"AdminAuth" method:"post" summary:"Admin logout"`
}

type LogoutRes struct {
	g.Meta `mime:"application/json"`
}
