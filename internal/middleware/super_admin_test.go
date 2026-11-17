package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	_ "github.com/suxiangdong/indu-scope-backend/internal/logic/bizctx"
	"github.com/suxiangdong/indu-scope-backend/internal/middleware"
	authModel "github.com/suxiangdong/indu-scope-backend/internal/model/auth"
	bizctxModel "github.com/suxiangdong/indu-scope-backend/internal/model/bizctx"
)

func TestSuperAdminMiddlewareDeniesRefundMutations(t *testing.T) {
	for _, path := range []string{"/api/v1/admin/vip/orders/10/refund", "/api/v1/admin/vip/refunds/20/retry"} {
		for _, loggedIn := range []bool{false, true} {
			ctx := context.Background()
			want := http.StatusUnauthorized
			if loggedIn {
				ctx = context.WithValue(ctx, consts.BizctxContextKey, &bizctxModel.Context{AdminUser: &authModel.AdminUser{ID: 1, Status: 1, IsSuperAdmin: false}})
				want = http.StatusForbidden
			}
			r := &ghttp.Request{Request: httptest.NewRequest(http.MethodPost, path, nil).WithContext(ctx), Response: &ghttp.Response{}}
			// Only the buffered status is needed for a denied request. Initialize
			// the exposed embedded pointer without opening a listening socket.
			r.Response.BufferWriter = allocatePointer(r.Response.BufferWriter)
			middleware.RequireSuperAdmin(r)
			if r.Response.Status != want || r.GetError() == nil {
				t.Fatalf("path %s loggedIn=%t returned status=%d error=%v", path, loggedIn, r.Response.Status, r.GetError())
			}
		}
	}
}

func allocatePointer[T any](_ *T) *T { return new(T) }
