package cmd

import (
	"context"
	"html"
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
	if err == nil && value != nil {
		switch strings.ToLower(strings.TrimSpace(value.String())) {
		case "0", "false", "no", "off":
			server.SetSwaggerPath("")
			server.SetOpenApiPath("")
			return
		}
	}
	path, err := g.Cfg().Get(ctx, "server.swaggerPath")
	if err != nil || path == nil || strings.TrimSpace(path.String()) == "" {
		return
	}
	if custom, err := g.Cfg().Get(ctx, "server.swaggerUITemplate"); err == nil && custom != nil && custom.String() != "" {
		return
	}
	server.SetSwaggerUITemplate(localSwaggerTemplate(path.String()))
}

// GoFrame already embeds a compatible ReDoc bundle under SwaggerPath. Serving
// that version avoids the moving external "latest" build, which can introduce
// module-only syntax into a page that loads a classic script.
func localSwaggerTemplate(swaggerPath string) string {
	assetURL := html.EscapeString(strings.TrimRight(swaggerPath, "/") + "/redoc.standalone.js")
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>源仓 API 文档</title>
  <style>body{margin:0;padding:0}#docs-error{margin:24px;font:16px/1.6 sans-serif;color:#9b2c2c}</style>
</head>
<body>
  <redoc spec-url="{SwaggerUIDocUrl}" show-object-schema-examples="true"></redoc>
  <p id="docs-error" hidden>接口文档脚本加载失败，请刷新页面。也可<a href="{SwaggerUIDocUrl}">查看 OpenAPI 定义</a>。</p>
  <noscript>请启用 JavaScript 后查看接口文档。</noscript>
  <script src="` + assetURL + `" onerror="document.getElementById('docs-error').hidden=false"></script>
</body>
</html>`
}
