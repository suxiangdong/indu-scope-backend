package middleware

import (
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/suxiangdong/indu-scope-backend/internal/companyaccess"
	authmodel "github.com/suxiangdong/indu-scope-backend/internal/model/auth"
	model "github.com/suxiangdong/indu-scope-backend/internal/model/bizctx"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

// Frontend user roles. Keep these values aligned with users.role.
const (
	UserRolePending  uint8 = 0
	UserRoleBuyer    uint8 = 1
	UserRoleSupplier uint8 = 2
)

// UserStatusVerified is the only status allowed to access normal business
// APIs. Status 0 means the profile is still being verified and status 2 is
// disabled by an administrator.
const UserStatusVerified uint8 = 1

// Ctx attaches the frontend user session to the business context. It never
// rejects a request; access requirements are applied by the composed
// Require* middleware below.
func Ctx(r *ghttp.Request) {
	businessContext := &model.Context{Session: r.Session}
	ctx := service.Bizctx().Set(r.Context(), businessContext)
	r.SetCtx(ctx)

	if token, ok := bearerToken(r.Header.Get("Authorization")); ok {
		// The frontend bearer token is also the server-side session id.
		if err := r.Session.SetId(token); err == nil {
			if user, err := service.Auth().Me(ctx, token); err == nil {
				businessContext.User = &user
			}
		}
	}

	r.Middleware.Next()
}

// AdminCtx attaches the management-console administrator session to the
// business context. Routes that require management authentication can use
// RequireAdmin.
func AdminCtx(r *ghttp.Request) {
	businessContext := &model.Context{Session: r.Session}
	ctx := service.Bizctx().Set(r.Context(), businessContext)
	r.SetCtx(ctx)

	if token, ok := bearerToken(r.Header.Get("Authorization")); ok {
		// The admin bearer token is also the server-side session id.
		if err := r.Session.SetId(token); err == nil {
			if user, err := service.AdminAuth().Me(ctx, token); err == nil {
				businessContext.AdminUser = &user
			}
		}
	}

	r.Middleware.Next()
}

// RequireAuthenticated requires a logged-in frontend user. It intentionally
// does not check profile status or role, so it can protect verification APIs
// that must remain available to users whose status is not yet verified.
func RequireAuthenticated(r *ghttp.Request) {
	if _, ok := currentUser(r); !ok {
		setUnauthorized(r, "authorization is required")
		return
	}
	r.Middleware.Next()
}

// RequireUser is kept as a compatibility alias for RequireAuthenticated.
func RequireUser(r *ghttp.Request) {
	RequireAuthenticated(r)
}

// RequireStatus returns middleware that requires a logged-in user with the
// given status. It is the generic status building block for route groups.
func RequireStatus(status uint8) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		user, ok := currentUser(r)
		if !ok {
			setUnauthorized(r, "authorization is required")
			return
		}
		if user.Status != status {
			setForbidden(r, statusMessage(status))
			return
		}
		r.Middleware.Next()
	}
}

// RequireVerified requires a logged-in user whose profile status is verified.
// Use it after RequireAuthenticated when composing route permissions. It also
// performs the login check itself so it is safe to use on its own.
func RequireVerified(r *ghttp.Request) {
	RequireStatus(UserStatusVerified)(r)
}

// RequireActiveVip protects member-only operations and company profiles. It must be placed after
// Ctx and RequireAuthenticated; verified users without an active subscription
// receive a 403 instead of an authentication error.
func RequireActiveVip(r *ghttp.Request) {
	user, ok := currentUser(r)
	if !ok {
		setUnauthorized(r, "authorization is required")
		return
	}
	if user.Status != UserStatusVerified {
		setForbidden(r, "verified profile is required")
		return
	}
	allowed, err := companyaccess.Allowed(r.Context(), uint64(user.ID))
	if err != nil {
		r.Response.Status = http.StatusInternalServerError
		r.SetError(gerror.NewCode(gcode.CodeInternalError, "internal server error"))
		return
	}
	if !allowed {
		setForbidden(r, "active VIP membership is required")
		return
	}
	r.Middleware.Next()
}

// RequireRole returns middleware that requires the given frontend user role.
// Role checks are deliberately independent from status checks so routes can
// compose RequireAuthenticated, RequireVerified and RequireRole as needed.
func RequireRole(role uint8) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		user, ok := currentUser(r)
		if !ok {
			setUnauthorized(r, "authorization is required")
			return
		}
		if user.Role != role {
			setForbidden(r, roleMessage(role))
			return
		}
		r.Middleware.Next()
	}
}

// RequireBuyer is a convenience middleware for routes that need both a
// verified profile and the buyer role. New route groups should prefer the
// explicit composition RequireAuthenticated, RequireVerified,
// RequireRole(UserRoleBuyer).
func RequireBuyer(r *ghttp.Request) {
	requireVerifiedRole(r, UserRoleBuyer)
}

// RequireSupplier is the supplier equivalent of RequireBuyer.
func RequireSupplier(r *ghttp.Request) {
	requireVerifiedRole(r, UserRoleSupplier)
}

// RequireAdmin stops the request unless AdminCtx loaded an active administrator session.
func RequireAdmin(r *ghttp.Request) {
	if value := service.Bizctx().Get(r.Context()); value == nil || value.AdminUser == nil {
		setUnauthorized(r, "admin authorization is required")
		return
	}
	r.Middleware.Next()
}

// RequireSuperAdmin requires an active administrator with the super-admin flag.
// Keep this separate from RequireAdmin so sensitive routes can compose the two
// checks explicitly and ordinary administrators receive a 403 response.
func RequireSuperAdmin(r *ghttp.Request) {
	value := service.Bizctx().Get(r.Context())
	if value == nil || value.AdminUser == nil {
		setUnauthorized(r, "admin authorization is required")
		return
	}
	if !value.AdminUser.IsSuperAdmin {
		setForbidden(r, "super administrator authorization is required")
		return
	}
	r.Middleware.Next()
}

func currentUser(r *ghttp.Request) (*authmodel.User, bool) {
	value := service.Bizctx().Get(r.Context())
	if value == nil || value.User == nil {
		return nil, false
	}
	return value.User, true
}

func requireVerifiedRole(r *ghttp.Request, role uint8) {
	user, ok := currentUser(r)
	if !ok {
		setUnauthorized(r, "authorization is required")
		return
	}
	if user.Status != UserStatusVerified {
		setForbidden(r, "verified profile is required")
		return
	}
	if user.Role != role {
		setForbidden(r, roleMessage(role))
		return
	}
	r.Middleware.Next()
}

// setUnauthorized keeps the API envelope compatible with existing clients
// while exposing the standard HTTP status used by clients to invalidate a
// cached login session.
func setUnauthorized(r *ghttp.Request, message string) {
	r.Response.Status = http.StatusUnauthorized
	r.SetError(gerror.NewCode(gcode.CodeNotAuthorized, message))
}

// setForbidden distinguishes a valid session without the required profile
// status/role from an expired or missing session. Clients must not clear a
// valid login when an endpoint is simply outside that user's permissions.
func setForbidden(r *ghttp.Request, message string) {
	r.Response.Status = http.StatusForbidden
	r.SetError(gerror.NewCode(gcode.CodeNotAuthorized, message))
}

func roleMessage(role uint8) string {
	switch role {
	case UserRoleBuyer:
		return "buyer authorization is required"
	case UserRoleSupplier:
		return "supplier authorization is required"
	default:
		return "required role authorization is required"
	}
}

func statusMessage(status uint8) string {
	if status == UserStatusVerified {
		return "verified profile is required"
	}
	return "required user status is not available"
}

func bearerToken(authorization string) (string, bool) {
	parts := strings.Fields(authorization)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}
