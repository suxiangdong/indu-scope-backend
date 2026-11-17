package v1

import "github.com/gogf/gf/v2/frame/g"

type ProfileReq struct {
	g.Meta `path:"/api/v1/user-profiles/profile" tags:"UserProfiles" method:"get" summary:"Get a verified user profile"`
	ID     uint64 `json:"id" in:"query" v:"required" dc:"User ID"`
}

type ProfileInfo struct {
	ID                uint64   `json:"id"`
	CompanyName       string   `json:"company_name"`
	Role              uint8    `json:"role"`
	RoleName          string   `json:"role_name"`
	Introduction      string   `json:"introduction,omitempty"`
	RegisteredAddress string   `json:"registered_address,omitempty"`
	Offices           []string `json:"offices"`
	Factories         []string `json:"factories"`
}

type ProfileRes struct {
	g.Meta `mime:"application/json"`
	ProfileInfo
}
