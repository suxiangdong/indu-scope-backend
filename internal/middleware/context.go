package middleware

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	model "github.com/suxiangdong/indu-scope-backend/internal/model/bizctx"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

// Ctx attaches the frontend user session to the business context. Routes that
// require frontend authentication can use RequireUser.
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

// RequireUser isLogin reports whether the current request session contains an authenticated
// user. The token itself is only an identifier; the session data is the source
// of truth for whether the user is still logged in.
// RequireUser stops the request unless Ctx loaded an active user session.
func RequireUser(r *ghttp.Request) {
	if value := service.Bizctx().Get(r.Context()); value == nil || value.User == nil {
		r.SetError(gerror.NewCode(gcode.CodeNotAuthorized, "authorization is required"))
		return
	}
	r.Middleware.Next()
}

// RequireAdmin stops the request unless Ctx loaded an active administrator session.
func RequireAdmin(r *ghttp.Request) {
	if value := service.Bizctx().Get(r.Context()); value == nil || value.AdminUser == nil {
		r.SetError(gerror.NewCode(gcode.CodeNotAuthorized, "admin authorization is required"))
		return
	}
	r.Middleware.Next()
}

func bearerToken(authorization string) (string, bool) {
	parts := strings.Fields(authorization)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}
