package bizctx

import (
	"github.com/gogf/gf/v2/net/ghttp"

	authmodel "github.com/suxiangdong/indu-scope-backend/internal/model/auth"
)

// ContextUser is the user data serialized in the request session.
type ContextUser struct {
	ID           int64  `json:"id"`
	Phone        string `json:"phone"`
	CompanyName  string `json:"company_name"`
	Username     string `json:"username"`
	Name         string `json:"name"`
	Role         uint8  `json:"role"`
	UserType     uint8  `json:"user_type"`
	Status       uint8  `json:"status"`
	IsSuperAdmin bool   `json:"is_super_admin"`
}

// Context is the request-scoped business context. Session is the GoFrame
// request session; User or AdminUser is populated by the context middleware.
type Context struct {
	Session   *ghttp.Session
	User      *authmodel.User
	AdminUser *authmodel.AdminUser
}
