package companyaccess

import (
	"testing"
	"time"

	vipModel "github.com/suxiangdong/indu-scope-backend/internal/model/vip"
)

func TestCompanyAccessRequiresCurrentActiveMembership(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	past, future := now.Add(-time.Hour), now.Add(time.Hour)
	for _, test := range []struct {
		name         string
		subscription *vipModel.Subscription
		allowed      bool
	}{
		{name: "nonmember"},
		{name: "active", subscription: &vipModel.Subscription{Status: "active", StartAt: &past, EndAt: &future}, allowed: true},
		{name: "lifetime", subscription: &vipModel.Subscription{Status: "active", StartAt: &past}, allowed: true},
		{name: "expired", subscription: &vipModel.Subscription{Status: "active", EndAt: &past}},
		{name: "expiry boundary", subscription: &vipModel.Subscription{Status: "active", EndAt: &now}},
		{name: "not started", subscription: &vipModel.Subscription{Status: "active", StartAt: &future, EndAt: &future}},
		{name: "start boundary", subscription: &vipModel.Subscription{Status: "active", StartAt: &now, EndAt: &future}, allowed: true},
		{name: "revoked", subscription: &vipModel.Subscription{Status: "cancelled", EndAt: &future}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := activeAt(test.subscription, now); got != test.allowed {
				t.Fatalf("company access=%t, want=%t", got, test.allowed)
			}
		})
	}
}

func TestLockedCompanySummaryContainsNoIdentity(t *testing.T) {
	if summary := Summary(123, "测试企业", false); summary != nil {
		t.Fatalf("locked company identity escaped: %+v", summary)
	}
	if summary := Summary(0, "测试企业", true); summary != nil {
		t.Fatalf("missing company must not produce a navigable summary: %+v", summary)
	}
	summary := Summary(123, " 测试企业 ", true)
	if summary == nil || summary.ID != 123 || summary.CompanyName != "测试企业" {
		t.Fatalf("unexpected member company summary: %+v", summary)
	}
}
