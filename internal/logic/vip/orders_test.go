package vip

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	vipModel "github.com/suxiangdong/indu-scope-backend/internal/model/vip"
)

func expectCheckoutUserLock(mock sqlmock.Sqlmock) {
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM `users`(.+)FOR UPDATE").WithArgs(uint64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
}

func expectCheckoutTerms(mock sqlmock.Sqlmock, price uint) {
	mock.ExpectQuery("SELECT (.+) FROM `payment_orders`(.+)idempotency_key(.+)").WithArgs(uint64(2), "same-checkout").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery("SELECT (.+) FROM `vip_plans`(.+)").WithArgs(uint64(3), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "plan_type", "billing_cycle", "duration_days", "price_fen", "currency"}).
			AddRow(3, "ordinary", "月度会员", "normal", "month", 30, price, "CNY"))
	mock.ExpectQuery("SELECT (.+) FROM `vip_agreements`(.+)").WithArgs("service").
		WillReturnRows(sqlmock.NewRows([]string{"id", "agreement_key", "title", "content", "status", "version"}).
			AddRow(1, "service", "服务协议", "正文", 1, 2))
	mock.ExpectQuery("SELECT (.+) FROM `vip_subscriptions`(.+)").WithArgs(uint64(2), "active").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
}

// Both free activation and paid checkout use the same transaction boundary.
// A retry waiting on the user's row sees the committed key before checking
// membership eligibility or mutable catalog/agreement availability.
func TestCheckoutRetryAfterSerializedCreation(t *testing.T) {
	for _, price := range []uint{0, 20000} {
		t.Run(fmt.Sprint(price), func(t *testing.T) {
			s, mock := paymentServiceFixture(t)
			expectCheckoutUserLock(mock)
			expectCheckoutTerms(mock, price)
			wantStatus, wantMethod := "pending", "wechat_miniapp"
			if price == 0 {
				mock.ExpectQuery("SELECT COUNT(.+) FROM `payment_orders`(.+)").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectExec("INSERT INTO `vip_subscriptions`").WillReturnResult(sqlmock.NewResult(7, 1))
				wantStatus, wantMethod = "paid", "free"
			}
			mock.ExpectExec("INSERT INTO `payment_orders`(.+)idempotency_key(.+)").WillReturnResult(sqlmock.NewResult(10, 1))
			mock.ExpectCommit()
			first, reused, err := s.CreateOrder(context.Background(), 2, 3, false, "same-checkout")
			if err != nil || reused || first.ID != 10 || first.Status != wantStatus || first.AmountFen != price || first.DurationDaysSnapshot != 30 {
				t.Fatalf("first=%+v reused=%t err=%v", first, reused, err)
			}
			if !regexp.MustCompile(`^[A-Za-z0-9_-]{6,32}$`).MatchString(first.OrderNo) {
				t.Fatalf("generated merchant order number violates WeChat Pay's 6-32 character contract: length=%d", len(first.OrderNo))
			}
			expectCheckoutUserLock(mock)
			mock.ExpectQuery("SELECT (.+) FROM `payment_orders`(.+)idempotency_key(.+)").WithArgs(uint64(2), "same-checkout").
				WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "plan_id", "order_no", "amount_fen", "status", "payment_method", "plan_type_snapshot"}).
					AddRow(first.ID, 2, 3, first.OrderNo, price, wantStatus, wantMethod, "normal"))
			mock.ExpectCommit()
			retry, reused, err := s.CreateOrder(context.Background(), 2, 3, false, "same-checkout")
			if err != nil || !reused || retry.ID != first.ID || retry.Status != wantStatus || retry.AmountFen != price {
				t.Fatalf("retry=%+v reused=%t err=%v", retry, reused, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestFreeActivationRollsBackWhenOrderCannotBeStored(t *testing.T) {
	s, mock := paymentServiceFixture(t)
	expectCheckoutUserLock(mock)
	expectCheckoutTerms(mock, 0)
	mock.ExpectQuery("SELECT COUNT(.+) FROM `payment_orders`(.+)").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec("INSERT INTO `vip_subscriptions`").WillReturnResult(sqlmock.NewResult(7, 1))
	mock.ExpectExec("INSERT INTO `payment_orders`").WillReturnError(errors.New("order write failed"))
	mock.ExpectRollback()
	if _, _, err := s.CreateOrder(context.Background(), 2, 3, false, "same-checkout"); err == nil {
		t.Fatal("free activation succeeded without its audit order")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMissingSubscriptionIsNormalIncludingExpiryBoundary(t *testing.T) {
	for _, noRowsError := range []bool{false, true} {
		t.Run(fmt.Sprint(noRowsError), func(t *testing.T) {
			s, mock := paymentServiceFixture(t)
			query := mock.ExpectQuery("SELECT (.+) FROM `vip_subscriptions`(.+)end_at. IS NULL DESC(.+)end_at. DESC(.+)").WithArgs(uint64(2), "active")
			if noRowsError {
				query.WillReturnError(sql.ErrNoRows)
			} else {
				query.WillReturnRows(sqlmock.NewRows([]string{"id"}))
			}
			result, err := s.GetSubscription(context.Background(), 2)
			if result != nil || err != nil {
				t.Fatalf("missing membership result=%+v err=%v", result, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSubscriptionUsesFrozenTypeAndPrefersLifetimeMembership(t *testing.T) {
	s, mock := paymentServiceFixture(t)
	mock.ExpectQuery("SELECT (.+)s.plan_type_snapshot AS plan_plan_type(.+)s.currency_snapshot AS plan_currency(.+)ORDER BY (.+)end_at. IS NULL DESC,(.+)end_at. DESC,(.+)id. DESC(.+)").
		WithArgs(uint64(2), "active").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "plan_id", "plan_plan_type", "plan_currency", "end_at", "start_at"}).
			AddRow(7, 2, 3, "normal", "CNY", nil, time.Now()))
	result, err := s.GetSubscription(context.Background(), 2)
	if err != nil || result == nil || result.EndAt != nil || result.Plan.PlanType != "normal" || result.Plan.Currency != "CNY" {
		t.Fatalf("subscription=%+v err=%v", result, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPlanAndPaymentRejectUnsettleableDuration(t *testing.T) {
	for _, days := range []uint{0, 1, maxMembershipDurationDays, maxMembershipDurationDays + 1} {
		input := vipModel.PlanInput{Code: "ordinary", Name: "普通会员", PlanType: "normal", BillingCycle: "custom", DurationDays: days}
		err := validatePlanInput(input)
		wantValid := days > 0 && days <= maxMembershipDurationDays
		if (err == nil) != wantValid {
			t.Fatalf("duration %d validation error=%v", days, err)
		}
	}
	s, mock := paymentServiceFixture(t)
	mock.ExpectQuery("SELECT (.+) FROM `payment_orders`(.+)").WithArgs(uint64(10), uint64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "status", "duration_days_snapshot"}).AddRow(10, 2, "pending", maxMembershipDurationDays+1))
	if _, err := s.PreparePayment(context.Background(), 2, 10); err == nil {
		t.Fatal("unsettleable duration reached WeChat prepayment")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
