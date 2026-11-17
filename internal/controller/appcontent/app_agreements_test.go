package appcontent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"
	v1 "github.com/suxiangdong/indu-scope-backend/api/appcontent/v1"
	_ "github.com/suxiangdong/indu-scope-backend/internal/logic/vip"
	vipModel "github.com/suxiangdong/indu-scope-backend/internal/model/vip"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

type publicAgreementService struct{ service.IVip }

func (publicAgreementService) ListAgreements(_ context.Context, activeOnly bool) ([]vipModel.Agreement, error) {
	if !activeOnly {
		return nil, fmt.Errorf("public request must only request active agreements")
	}
	return []vipModel.Agreement{
		{ID: 1, Key: "service", Title: "会员协议", Content: "会员正文", Status: 1, Version: 2},
		{ID: 2, Key: "auto-renew", Title: "续费协议", Content: "续费正文", Status: 1, Version: 1},
		{ID: 3, Key: "user", Title: "用户协议", Content: "用户正文", Status: 1, Version: 3},
		{ID: 4, Key: "privacy", Title: "隐私政策", Content: "隐私正文", Status: 1, Version: 4},
	}, nil
}

func TestAppAgreementsCanBeReadWithoutLogin(t *testing.T) {
	previous := service.Vip()
	service.RegisterVip(publicAgreementService{})
	t.Cleanup(func() { service.RegisterVip(previous) })
	s := ghttp.GetServer(guid.S())
	s.SetAddr("127.0.0.1:0")
	s.SetDumpRouterMap(false)
	s.SetAccessLogEnabled(false)
	s.SetErrorLogEnabled(false)
	s.SetSessionCookieOutput(false)
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(ghttp.MiddlewareHandlerResponse)
		group.Bind(NewV1())
	})
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Shutdown() })
	client := &http.Client{Timeout: 3 * time.Second}
	// No Authorization header, cookie, user role, or verified/VIP state.
	response, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/api/v1/app-agreements", s.GetListenedPort()))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var payload struct {
		Code int              `json:"code"`
		Data v1.AgreementsRes `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || payload.Code != 0 || len(payload.Data.Items) != 2 ||
		payload.Data.Items[0].Key != "user" || payload.Data.Items[1].Key != "privacy" || payload.Data.Items[1].Version != 4 {
		t.Fatalf("status=%d payload=%+v", response.StatusCode, payload)
	}
}
