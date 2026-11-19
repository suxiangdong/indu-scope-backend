package userprofile

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"
	_ "github.com/suxiangdong/indu-scope-backend/internal/logic/bizctx"
	_ "github.com/suxiangdong/indu-scope-backend/internal/logic/userprofile"
	_ "github.com/suxiangdong/indu-scope-backend/internal/logic/vip"
	"github.com/suxiangdong/indu-scope-backend/internal/middleware"
	authModel "github.com/suxiangdong/indu-scope-backend/internal/model/auth"
	bizctxModel "github.com/suxiangdong/indu-scope-backend/internal/model/bizctx"
	userprofileModel "github.com/suxiangdong/indu-scope-backend/internal/model/userprofile"
	vipModel "github.com/suxiangdong/indu-scope-backend/internal/model/vip"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

type membershipStub struct {
	service.IVip
	subscription *vipModel.Subscription
	err          error
}

func (s membershipStub) GetSubscription(context.Context, uint64) (*vipModel.Subscription, error) {
	return s.subscription, s.err
}

type companyProfileStub struct {
	service.IUserProfile
	calls atomic.Int32
}

func (s *companyProfileStub) Profile(_ context.Context, id uint64) (userprofileModel.Profile, error) {
	s.calls.Add(1)
	return userprofileModel.Profile{
		ID: id, Role: 2, CompanyName: "PRIVATE_COMPANY_NAME", Introduction: "PRIVATE_INTRODUCTION",
		RegisteredAddress: "PRIVATE_ADDRESS", Offices: []string{}, Factories: []string{},
	}, nil
}

func TestCompanyProfileRouteCannotBypassMembership(t *testing.T) {
	past, future := time.Now().Add(-time.Hour), time.Now().Add(time.Hour)
	for _, test := range []struct {
		name         string
		subscription *vipModel.Subscription
		err          error
		status       int
		profileCalls int32
	}{
		{name: "nonmember", status: http.StatusForbidden},
		{name: "expired member", subscription: &vipModel.Subscription{Status: "active", EndAt: &past}, status: http.StatusForbidden},
		{name: "future member", subscription: &vipModel.Subscription{Status: "active", StartAt: &future}, status: http.StatusForbidden},
		{name: "active member", subscription: &vipModel.Subscription{Status: "active", EndAt: &future}, status: http.StatusOK, profileCalls: 1},
		{name: "lifetime member", subscription: &vipModel.Subscription{Status: "active"}, status: http.StatusOK, profileCalls: 1},
		{name: "membership lookup fails", err: errors.New("membership unavailable"), status: http.StatusInternalServerError},
	} {
		t.Run(test.name, func(t *testing.T) {
			previousVip, previousProfile := service.Vip(), service.UserProfile()
			profile := &companyProfileStub{}
			service.RegisterVip(membershipStub{subscription: test.subscription, err: test.err})
			service.RegisterUserProfile(profile)
			t.Cleanup(func() {
				service.RegisterVip(previousVip)
				service.RegisterUserProfile(previousProfile)
			})

			s := ghttp.GetServer(guid.S())
			s.SetAddr("127.0.0.1:0")
			s.SetDumpRouterMap(false)
			s.SetAccessLogEnabled(false)
			s.SetErrorLogEnabled(false)
			s.SetSessionCookieOutput(false)
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse, middleware.HTTPErrorStatus, func(r *ghttp.Request) {
					r.SetCtx(service.Bizctx().Set(r.Context(), &bizctxModel.Context{User: &authModel.User{ID: 10, Status: 1, Role: 1}}))
					r.Middleware.Next()
				}, middleware.RequireAuthenticated, middleware.RequireVerified, middleware.RequireActiveVip)
				group.Bind(NewV1())
			})
			if err := s.Start(); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = s.Shutdown() })
			client := &http.Client{Timeout: 3 * time.Second}
			response, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/api/v1/user-profiles/profile?id=88", s.GetListenedPort()))
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != test.status || profile.calls.Load() != test.profileCalls {
				t.Fatalf("status=%d calls=%d body=%s", response.StatusCode, profile.calls.Load(), body)
			}
			exposesCompany := strings.Contains(string(body), "PRIVATE_COMPANY_NAME")
			if exposesCompany != (test.profileCalls == 1) {
				t.Fatalf("unexpected company visibility: %s", body)
			}
			if test.profileCalls == 0 && (strings.Contains(string(body), "PRIVATE_ADDRESS") || strings.Contains(string(body), "PRIVATE_INTRODUCTION")) {
				t.Fatalf("blocked profile leaked company data: %s", body)
			}
		})
	}
}
