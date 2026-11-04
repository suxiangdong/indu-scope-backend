package v1

import "github.com/gogf/gf/v2/frame/g"

type RegisterReq struct {
	g.Meta   `path:"/api/v1/auth/register" tags:"Auth" method:"post" summary:"Register with phone and password"`
	Phone    string `json:"phone" v:"required" dc:"Phone number"`
	Password string `json:"password" v:"required" dc:"Password, at least 8 characters"`
	Role     uint8  `json:"role" v:"required" dc:"1 buyer, 2 supplier"`
	UserType uint8  `json:"user_type" v:"required" dc:"1 personal, 2 company"`
}

type LoginReq struct {
	g.Meta   `path:"/api/v1/auth/login" tags:"Auth" method:"post" summary:"Login with phone and password"`
	Phone    string `json:"phone" v:"required" dc:"Phone number"`
	Password string `json:"password" v:"required" dc:"Password"`
}

type UserInfo struct {
	ID       int64  `json:"id" dc:"User ID"`
	Phone    string `json:"phone" dc:"Phone number"`
	Role     uint8  `json:"role" dc:"1 buyer, 2 supplier"`
	UserType uint8  `json:"user_type" dc:"1 personal, 2 company"`
	Status   uint8  `json:"status" dc:"0 unverified, 1 verified, 2 disabled"`
}

type AuthData struct {
	AccessToken string   `json:"access_token" dc:"JWT access token"`
	TokenType   string   `json:"token_type" example:"Bearer"`
	User        UserInfo `json:"user"`
}

type RegisterRes struct {
	g.Meta `mime:"application/json"`
	AuthData
}

type LoginRes struct {
	g.Meta `mime:"application/json"`
	AuthData
}

type MeReq struct {
	g.Meta `path:"/api/v1/auth/me" tags:"Auth" method:"get" summary:"Get current user"`
}

type MeRes struct {
	g.Meta `mime:"application/json"`
	UserInfo
}

type LogoutReq struct {
	g.Meta `path:"/api/v1/auth/logout" tags:"Auth" method:"post" summary:"Logout"`
}

type LogoutRes struct {
	g.Meta `mime:"application/json"`
}
