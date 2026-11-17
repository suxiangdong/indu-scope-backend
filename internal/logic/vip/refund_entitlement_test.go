package vip

import (
	"testing"
	"time"
)

func TestEntitlementRefundPreservesOtherPurchases(t *testing.T) {
	base := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	days := func(n int) time.Time { return base.Add(time.Duration(n) * 24 * time.Hour) }
	grant := func(id uint64, start, end int) orderRow {
		a, b := days(start), days(end)
		return orderRow{ID: id, PlanID: id, DurationDaysSnapshot: uint(end - start), EntitlementStartAt: &a, EntitlementEndAt: &b}
	}
	for _, test := range []struct {
		name       string
		now        int
		target     uint64
		wantDays   int
		wantEnd    int
		wantShifts int
	}{
		{"current partly consumed", 10, 1, 20, 40, 1},
		{"future renewal", 10, 2, 30, 30, 0},
		{"spent original order", 40, 1, 0, 60, 0},
		{"last partly consumed", 40, 2, 20, 40, 0},
		{"all spent", 80, 2, 0, 60, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			plan, err := planEntitlementReclamation(days(test.now), days(60), []orderRow{grant(2, 30, 60), grant(1, 0, 30)}, test.target)
			if err != nil {
				t.Fatal(err)
			}
			if plan.Duration != time.Duration(test.wantDays)*24*time.Hour || !plan.EndAt.Equal(days(test.wantEnd)) || len(plan.Shifts) != test.wantShifts {
				t.Fatalf("incorrect reclamation %+v", plan)
			}
			for _, shift := range plan.Shifts {
				if shift.EndAt.Sub(shift.StartAt) != 30*24*time.Hour {
					t.Fatal("another order lost purchased time")
				}
			}
		})
	}
	// Refunding one purchase and then another must use the shifted allocations.
	p, err := planEntitlementReclamation(days(10), days(60), []orderRow{grant(1, 0, 30), grant(2, 30, 60)}, 1)
	if err != nil {
		t.Fatal(err)
	}
	shifted := grant(2, 10, 40)
	p2, err := planEntitlementReclamation(days(20), p.EndAt, []orderRow{shifted}, 2)
	if err != nil || p2.Duration != 20*24*time.Hour || !p2.EndAt.Equal(days(20)) {
		t.Fatalf("sequential refund=%+v err=%v", p2, err)
	}
}

func TestUnknownOrModifiedEntitlementRequiresManualReview(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	end := now.Add(30 * 24 * time.Hour)
	known := orderRow{ID: 1, DurationDaysSnapshot: 30, EntitlementStartAt: &now, EntitlementEndAt: &end}
	for _, grants := range [][]orderRow{nil, {{ID: 1}}, {known, known}} {
		if _, err := planEntitlementReclamation(now, end, grants, 1); err == nil {
			t.Fatal("unverifiable allocations were automatically reclaimed")
		}
	}
	if _, err := planEntitlementReclamation(now, end.Add(time.Hour), []orderRow{known}, 1); err == nil {
		t.Fatal("manually adjusted subscription was silently changed")
	}
}

func TestEntitlementRestoresRemainingOrderTerms(t *testing.T) {
	base := time.Now().Truncate(time.Millisecond)
	firstEnd, lastEnd := base.Add(30*24*time.Hour), base.Add(60*24*time.Hour)
	first := orderRow{ID: 1, PlanID: 3, PlanTypeSnapshot: "normal", Currency: "CNY", DurationDaysSnapshot: 30, EntitlementStartAt: &base, EntitlementEndAt: &firstEnd}
	last := orderRow{ID: 2, PlanID: 4, PlanTypeSnapshot: "normal", Currency: "CNY", DurationDaysSnapshot: 30, EntitlementStartAt: &firstEnd, EntitlementEndAt: &lastEnd}
	plan, err := planEntitlementReclamation(base.Add(5*24*time.Hour), lastEnd, []orderRow{first, last}, last.ID)
	if err != nil || plan.PlanID != first.PlanID || plan.PlanType != first.PlanTypeSnapshot || plan.Currency != first.Currency {
		t.Fatalf("remaining terms=%+v err=%v", plan, err)
	}
}
