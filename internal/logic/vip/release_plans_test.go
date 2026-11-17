package vip

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestReleaseCatalogKeepsRecurringPlansForAdministratorsOnly(t *testing.T) {
	s, mock := paymentServiceFixture(t)
	columns := []string{"id", "name", "plan_type", "status"}
	mock.ExpectQuery("SELECT (.+) FROM `vip_plans` WHERE (.+)status(.+)plan_type(.+)").WithArgs(1, "normal").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, "普通套餐", "normal", 1))
	public, err := s.ListPlans(context.Background(), true)
	if err != nil || len(public) != 1 || public[0].PlanType != "normal" {
		t.Fatalf("public=%+v err=%v", public, err)
	}
	mock.ExpectQuery("SELECT (.+) FROM `vip_plans` ORDER BY").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(2, "连续包月", "auto_renew", 1).AddRow(3, "已停用", "normal", 0))
	admin, err := s.ListPlans(context.Background(), false)
	if err != nil || len(admin) != 2 || admin[0].PlanType != "auto_renew" || admin[1].Status != 0 {
		t.Fatalf("admin=%+v err=%v", admin, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestNewRecurringCheckoutCannotActivateEvenWhenFree(t *testing.T) {
	for _, price := range []uint{0, 20000} {
		t.Run(fmt.Sprint(price), func(t *testing.T) {
			s, mock := paymentServiceFixture(t)
			expectCheckoutUserLock(mock)
			mock.ExpectQuery("SELECT (.+) FROM `vip_plans`(.+)").WithArgs(uint64(3), 1).
				WillReturnRows(sqlmock.NewRows([]string{"id", "plan_type", "duration_days", "price_fen"}).AddRow(3, "auto_renew", 30, price))
			mock.ExpectRollback()
			if _, _, err := s.CreateOrder(context.Background(), 2, 3, true, ""); err == nil || !strings.Contains(err.Error(), "连续包月签约尚未开通") {
				t.Fatalf("unexpected checkout error: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestLegacyRecurringCheckoutRetryUsesItsExistingOrder(t *testing.T) {
	s, mock := paymentServiceFixture(t)
	expectCheckoutUserLock(mock)
	mock.ExpectQuery("SELECT (.+) FROM `payment_orders`(.+)idempotency_key(.+)").WithArgs(uint64(2), "legacy-checkout").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "plan_id", "plan_type_snapshot", "status"}).AddRow(9, 2, 3, "auto_renew", "paid"))
	mock.ExpectCommit()
	order, reused, err := s.CreateOrder(context.Background(), 2, 3, true, "legacy-checkout")
	if err != nil || !reused || order.ID != 9 || order.Status != "paid" {
		t.Fatalf("order=%+v reused=%t err=%v", order, reused, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
