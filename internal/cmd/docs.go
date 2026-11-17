package cmd

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// configureAPIDocs applies the application-level Swagger switch after
// GoFrame loads its server configuration. Keeping the switch outside
// ghttp.ServerConfig lets production disable both the Swagger UI and the
// OpenAPI document without changing the framework configuration type.
func configureAPIDocs(ctx context.Context, server *ghttp.Server) {
	value, err := g.Cfg().Get(ctx, "server.swagger_enabled")
	if err != nil || value == nil {
		return // preserve the existing enabled-by-path default
	}
	switch strings.ToLower(strings.TrimSpace(value.String())) {
	case "0", "false", "no", "off":
		server.SetSwaggerPath("")
		server.SetOpenApiPath("")
	}
}
