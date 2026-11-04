package v1

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type AdminUserInfo struct {
	ID          int64      `json:"id" dc:"Admin user ID"`
	Username    string     `json:"username" dc:"Admin username"`
	Name        string     `json:"name" dc:"Admin name"`
	Phone       string     `json:"phone" dc:"Phone number"`
	Status      uint8      `json:"status" dc:"0 disabled, 1 active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" dc:"Last login time"`
	CreatedAt   time.Time  `json:"created_at" dc:"Created time"`
	UpdatedAt   time.Time  `json:"updated_at" dc:"Updated time"`
}

type ListReq struct {
	g.Meta `path:"/api/v1/admin/users" tags:"AdminUsers" method:"get" summary:"List admin users"`
}

type ListRes struct {
	g.Meta `mime:"application/json"`
	Items  []AdminUserInfo `json:"items"`
}

type CreateReq struct {
	g.Meta   `path:"/api/v1/admin/users/create" tags:"AdminUsers" method:"post" summary:"Create admin user"`
	Username string `json:"username" v:"required" dc:"Admin username"`
	Name     string `json:"name" v:"required" dc:"Admin name"`
	Phone    string `json:"phone" v:"required" dc:"Phone number"`
	Password string `json:"password" v:"required" dc:"Password"`
	Status   *uint  `json:"status" dc:"0 disabled, 1 active"`
}

type CreateRes struct {
	g.Meta `mime:"application/json"`
	AdminUserInfo
}

type UpdateReq struct {
	g.Meta   `path:"/api/v1/admin/users/update" tags:"AdminUsers" method:"post" summary:"Update admin user"`
	ID       uint64 `json:"id" v:"required" dc:"Admin user ID"`
	Username string `json:"username" v:"required" dc:"Admin username"`
	Name     string `json:"name" v:"required" dc:"Admin name"`
	Phone    string `json:"phone" v:"required" dc:"Phone number"`
	Password string `json:"password" dc:"New password, optional"`
	Status   *uint  `json:"status" dc:"0 disabled, 1 active"`
}

type UpdateRes struct {
	g.Meta `mime:"application/json"`
	AdminUserInfo
}

type DeleteReq struct {
	g.Meta `path:"/api/v1/admin/users/delete" tags:"AdminUsers" method:"post" summary:"Delete admin user"`
	ID     uint64 `json:"id" v:"required" dc:"Admin user ID"`
}

type DeleteRes struct {
	g.Meta  `mime:"application/json"`
	Success bool `json:"success"`
}
