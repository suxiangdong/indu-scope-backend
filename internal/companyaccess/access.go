// Package companyaccess keeps company-detail visibility aligned with the
// active membership required by the full company-profile endpoint.
package companyaccess

import (
	"context"
	"strings"
	"time"

	userprofilesv1 "github.com/suxiangdong/indu-scope-backend/api/userprofiles/v1"
	vipModel "github.com/suxiangdong/indu-scope-backend/internal/model/vip"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

// Allowed always reads the viewer's current subscription. Do not cache its
// result with a product, requirement, or company row shared by other viewers.
func Allowed(ctx context.Context, userID uint64) (bool, error) {
	subscription, err := service.Vip().GetSubscription(ctx, userID)
	if err != nil {
		return false, err
	}
	return activeAt(subscription, time.Now()), nil
}

func activeAt(subscription *vipModel.Subscription, now time.Time) bool {
	if subscription == nil || subscription.Status != "active" {
		return false
	}
	if subscription.StartAt != nil && subscription.StartAt.After(now) {
		return false
	}
	return subscription.EndAt == nil || subscription.EndAt.After(now)
}

// Summary never includes a company identifier or name when access is locked.
func Summary(userID uint64, name string, allowed bool) *userprofilesv1.CompanySummary {
	if !allowed || userID == 0 {
		return nil
	}
	return &userprofilesv1.CompanySummary{ID: userID, CompanyName: strings.TrimSpace(name)}
}
