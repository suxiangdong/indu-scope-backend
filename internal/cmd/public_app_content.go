package cmd

import (
	"github.com/gogf/gf/v2/net/ghttp"

	appContentController "github.com/suxiangdong/indu-scope-backend/internal/controller/appcontent"
	"github.com/suxiangdong/indu-scope-backend/internal/middleware"
)

func registerPublicAppContentRoutes(s *ghttp.Server) {
	// These read-only pages and login agreements must remain accessible before
	// authentication. The authentication/OCR limiter uses the connection peer,
	// which is shared by users behind a reverse proxy; do not apply it here.
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(ghttp.MiddlewareHandlerResponse, middleware.HTTPErrorStatus, middleware.Ctx)
		group.Bind(appContentController.NewV1())
	})
}
