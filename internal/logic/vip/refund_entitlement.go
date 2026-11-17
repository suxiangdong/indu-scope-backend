package vip

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

type entitlementShift struct {
	ID             uint64
	StartAt, EndAt time.Time
}
type entitlementReclamation struct {
	Duration time.Duration
	EndAt    time.Time
	PlanID   uint64
	PlanType string
	Currency string
	Shifts   []entitlementShift
}

// Remove only the refunded grant's FUTURE interval. Later grants move earlier,
// retaining their complete duration; already consumed time is never taken from
// another purchase. Missing/inconsistent historical allocations need review.
func planEntitlementReclamation(now, subscriptionEnd time.Time, grants []orderRow, targetID uint64) (*entitlementReclamation, error) {
	rows := append([]orderRow(nil), grants...)
	if len(rows) == 0 {
		return nil, errors.New("missing entitlement grants")
	}
	for _, r := range rows {
		if r.EntitlementStartAt == nil || r.EntitlementEndAt == nil || !r.EntitlementEndAt.After(*r.EntitlementStartAt) {
			return nil, errors.New("unknown entitlement interval")
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].EntitlementStartAt.Before(*rows[j].EntitlementStartAt) })
	var target *orderRow
	for i := range rows {
		if i > 0 && rows[i].EntitlementStartAt.Before(*rows[i-1].EntitlementEndAt) {
			return nil, errors.New("overlapping entitlement intervals")
		}
		if rows[i].ID == targetID {
			target = &rows[i]
		}
	}
	if target == nil || !rows[len(rows)-1].EntitlementEndAt.Equal(subscriptionEnd) {
		return nil, errors.New("subscription and order allocations disagree")
	}
	if target.DurationDaysSnapshot == 0 || target.EntitlementEndAt.Sub(*target.EntitlementStartAt) != time.Duration(target.DurationDaysSnapshot)*24*time.Hour {
		return nil, errors.New("entitlement snapshot mismatch")
	}
	start := *target.EntitlementStartAt
	if now.After(start) {
		start = now
	}
	remove := target.EntitlementEndAt.Sub(start)
	if remove < 0 {
		remove = 0
	}
	plan := &entitlementReclamation{Duration: remove, EndAt: subscriptionEnd.Add(-remove)}
	for _, row := range rows {
		if row.ID == targetID {
			continue
		}
		if row.EntitlementStartAt.Before(*target.EntitlementEndAt) || remove == 0 {
			plan.PlanID = row.PlanID
			plan.PlanType, plan.Currency = row.PlanTypeSnapshot, row.Currency
			continue
		}
		plan.Shifts = append(plan.Shifts, entitlementShift{ID: row.ID, StartAt: row.EntitlementStartAt.Add(-remove), EndAt: row.EntitlementEndAt.Add(-remove)})
		plan.PlanID = row.PlanID
		plan.PlanType, plan.Currency = row.PlanTypeSnapshot, row.Currency
	}
	return plan, nil
}

func reclaimOrderEntitlement(ctx context.Context, tx gdb.TX, order orderRow, now time.Time) (string, uint64, error) {
	if order.EntitlementRevokedAt != nil {
		return "none", 0, nil
	}
	if order.SubscriptionID == nil || order.EntitlementStartAt == nil || order.EntitlementEndAt == nil {
		return "manual", 0, nil
	}
	var subscription *subscriptionRow
	err := tx.Model("vip_subscriptions").Where("id", *order.SubscriptionID).Where("user_id", order.UserID).Scan(&subscription)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", 0, err
	}
	if subscription == nil || subscription.EndAt == nil {
		return "manual", 0, nil
	}
	var grants []orderRow
	if err := tx.Model("payment_orders").Where("subscription_id", *order.SubscriptionID).Where("status", "paid").Order("id").Scan(&grants); err != nil {
		return "", 0, err
	}
	plan, err := planEntitlementReclamation(now, *subscription.EndAt, grants, order.ID)
	if err != nil {
		return "manual", 0, nil
	}
	for _, shift := range plan.Shifts {
		if _, err := tx.Model("payment_orders").Where("id", shift.ID).Data(g.Map{"entitlement_start_at": shift.StartAt, "entitlement_end_at": shift.EndAt}).Update(); err != nil {
			return "", 0, err
		}
	}
	if plan.Duration > 0 {
		data := g.Map{"end_at": plan.EndAt}
		if plan.PlanID > 0 {
			data["plan_id"] = plan.PlanID
			// Old grants without trustworthy terms retain the frozen subscription
			// values rather than reloading a mutable catalog row.
			if plan.PlanType == "normal" || plan.PlanType == "auto_renew" {
				data["plan_type_snapshot"] = plan.PlanType
			}
			if plan.Currency != "" {
				data["currency_snapshot"] = plan.Currency
			}
		}
		if !plan.EndAt.After(now) {
			data["status"] = "expired"
			data["auto_renew"] = 0
			data["next_billing_at"] = nil
		}
		if _, err := tx.Model("vip_subscriptions").Where("id", subscription.ID).Data(data).Update(); err != nil {
			return "", 0, err
		}
	}
	if _, err := tx.Model("payment_orders").Where("id", order.ID).Data(g.Map{"entitlement_revoked_at": now}).Update(); err != nil {
		return "", 0, err
	}
	status := "none"
	if plan.Duration > 0 {
		status = "reclaimed"
	}
	return status, uint64(plan.Duration / time.Second), nil
}
