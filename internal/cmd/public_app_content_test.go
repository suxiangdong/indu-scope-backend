package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"
	_ "github.com/suxiangdong/indu-scope-backend/internal/logic/appcontent"
	_ "github.com/suxiangdong/indu-scope-backend/internal/logic/bizctx"
	_ "github.com/suxiangdong/indu-scope-backend/internal/logic/vip"
	"github.com/suxiangdong/indu-scope-backend/internal/middleware"
	appcontentModel "github.com/suxiangdong/indu-scope-backend/internal/model/appcontent"
	vipModel "github.com/suxiangdong/indu-scope-backend/internal/model/vip"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

type publicContentVipService struct{ service.IVip }

func (publicContentVipService) ListAgreements(ctx context.Context, activeOnly bool) ([]vipModel.Agreement, error) {
	if value := service.Bizctx().Get(ctx); value == nil || value.User != nil || !activeOnly {
		return nil, fmt.Errorf("expected anonymous request for active agreements")
	}
	return []vipModel.Agreement{
		{ID: 3, Key: "user", Title: "用户协议", Content: "用户正文", Status: 1, Version: 1},
		{ID: 4, Key: "privacy", Title: "隐私政策", Content: "隐私正文", Status: 1, Version: 1},
	}, nil
}

type publicContentPageService struct{ service.IAppContent }

func (publicContentPageService) Get(ctx context.Context, pageKey string) (appcontentModel.PageContent, error) {
	if value := service.Bizctx().Get(ctx); value == nil || value.User != nil {
		return appcontentModel.PageContent{}, fmt.Errorf("expected anonymous page request")
	}
	return appcontentModel.PageContent{ID: 1, PageKey: pageKey, Title: "联系我们", Body: "公开联系内容", Status: 1}, nil
}

func TestPublicAppContentDoesNotShareAuthenticationRateLimit(t *testing.T) {
	previousVip, previousAppContent := service.Vip(), service.AppContent()
	service.RegisterVip(publicContentVipService{})
	service.RegisterAppContent(publicContentPageService{})
	t.Cleanup(func() {
		service.RegisterVip(previousVip)
		service.RegisterAppContent(previousAppContent)
	})

	s := ghttp.GetServer(guid.S())
	s.SetAddr("127.0.0.1:0")
	s.SetDumpRouterMap(false)
	s.SetAccessLogEnabled(false)
	s.SetErrorLogEnabled(false)
	s.SetSessionCookieOutput(false)
	registerPublicAppContentRoutes(s)

	// An adjacent route keeps the same middleware and quota as authentication
	// endpoints. All local requests have the same connection-peer IP, as they
	// would when reaching the application through one reverse proxy.
	limitedPath := "/limited-" + guid.S()
	var limitedCalls atomic.Int32
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(ghttp.MiddlewareHandlerResponse, middleware.HTTPErrorStatus, middleware.RateLimit(time.Minute, 30), middleware.Ctx)
		group.GET(limitedPath, func(r *ghttp.Request) { limitedCalls.Add(1) })
	})
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	client := &http.Client{Timeout: 3 * time.Second}
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", s.GetListenedPort())
	for _, path := range []string{"/api/v1/app-agreements", "/api/v1/app-pages/contact", limitedPath} {
		for i := 1; i <= 35; i++ {
			// No credentials or forwarded-IP headers are supplied.
			response, err := client.Get(baseURL + path)
			if err != nil {
				t.Fatal(err)
			}
			var payload struct {
				Code int `json:"code"`
				Data struct {
					Items []struct {
						Key string `json:"key"`
					} `json:"items"`
					PageKey string `json:"page_key"`
				} `json:"data"`
			}
			err = json.NewDecoder(response.Body).Decode(&payload)
			_ = response.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
			if path == limitedPath && i > 30 {
				if response.StatusCode != http.StatusTooManyRequests || payload.Code != 63 {
					t.Fatalf("limited request %d: HTTP %d, code %d", i, response.StatusCode, payload.Code)
				}
				continue
			}
			if response.StatusCode != http.StatusOK || payload.Code != 0 {
				t.Fatalf("%s request %d: HTTP %d, code %d", path, i, response.StatusCode, payload.Code)
			}
			if path == "/api/v1/app-agreements" && (len(payload.Data.Items) != 2 || payload.Data.Items[0].Key != "user" || payload.Data.Items[1].Key != "privacy") {
				t.Fatalf("agreement request %d: missing public agreements", i)
			}
			if path == "/api/v1/app-pages/contact" && payload.Data.PageKey != "contact" {
				t.Fatalf("page request %d: missing public page", i)
			}
		}
	}
	if calls := limitedCalls.Load(); calls != 30 {
		t.Fatalf("limited handler ran %d times; expected 30", calls)
	}
}
