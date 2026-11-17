package middleware

import (
	"net/http"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// HTTPErrorStatus keeps transport errors separate from expected business
// validation errors. Business errors are returned as HTTP 200 with a non-zero
// response code and message so web and mini-program clients can display the
// server-provided message instead of replacing it with a generic HTTP error.
func HTTPErrorStatus(r *ghttp.Request) {
	r.Middleware.Next()
	if r.Response.Status != 0 && r.Response.Status != http.StatusOK {
		return
	}
	code := gerror.Code(r.GetError())
	status := 0
	switch code {
	case gcode.CodeNotAuthorized:
		status = http.StatusUnauthorized
	case gcode.CodeNotFound:
		status = http.StatusNotFound
	case gcode.CodeServerBusy:
		status = http.StatusTooManyRequests
	case gcode.CodeInternalError, gcode.CodeDbOperationError, gcode.CodeInternalPanic:
		status = http.StatusInternalServerError
	}
	if status != 0 {
		r.Response.Status = status
	}
}
