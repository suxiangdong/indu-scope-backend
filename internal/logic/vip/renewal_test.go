package vip

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	vipModel "github.com/suxiangdong/indu-scope-backend/internal/model/vip"
)

func TestNormalRenewalOptions(t *testing.T) {
	for _, name := range []string{"same plan", "quarter", "year", "free membership to paid", "free target", "recurring current", "recurring target", "other currency", "unlimited"} {
		t.Run(name, func(t *testing.T) {
			current := vipModel.Plan{ID: 1, PlanType: "normal", Currency: "CNY", PriceFen: 100, DurationDays: 30}
			target := current
			future := time.Now().Add(10 * 24 * time.Hour)
			endAt := &future
			wantErr := false
			switch name {
			case "quarter":
				target.ID, target.DurationDays = 2, 90
			case "year":
				target.ID, target.DurationDays = 3, 365
			case "free membership to paid":
				current.PriceFen, target.ID = 0, 2
			case "free target":
				target.PriceFen, wantErr = 0, true
			case "recurring current":
				current.PlanType, wantErr = "auto_renew", true
			case "recurring target":
				target.PlanType, wantErr = "auto_renew", true
			case "other currency":
				target.Currency, wantErr = "USD", true
			case "unlimited":
				endAt, wantErr = nil, true
			}
			if err := validateNormalRenewal(current, endAt, target); (err != nil) != wantErr {
				t.Fatalf("renewal validation err=%v, wantErr=%t", err, wantErr)
			}
		})
	}
}

// A callback must settle the frozen order without reloading the editable
// catalog. The subscription query uses its frozen type/currency instead.
func TestPaymentSettlementDoesNotReadCatalogTerms(t *testing.T) {
	s, mock := paymentServiceFixture(t)
	expectPaymentBegin(mock, "paying")
	mock.ExpectQuery("SELECT (.+) FROM `vip_subscriptions` s WHERE(.+)plan_type_snapshot(.+)currency_snapshot(.+)").
		WithArgs(uint64(2), "active", "normal", "CNY").
		WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "end_at", "plan_plan_type", "plan_currency"}).AddRow(7, 4, time.Now().Add(10*24*time.Hour), "normal", "CNY"))
	mock.ExpectExec("UPDATE `vip_subscriptions` SET (.+)plan_type_snapshot(.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE `payment_orders`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE `payment_events`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := s.applyPayment(context.Background(), "catalog-independent", "TRANSACTION.SUCCESS", []byte(`{}`), successfulTransaction()); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestLatePaymentCreatesGrantWithoutChangingSpecialMembership(t *testing.T) {
	s, mock := paymentServiceFixture(t)
	expectPaymentBegin(mock, "paying")
	// There may be an active lifetime/recurring membership. Only finite
	// ordinary same-currency subscriptions can be selected for extension.
	mock.ExpectQuery("SELECT (.+) FROM `vip_subscriptions` s WHERE(.+)plan_type_snapshot(.+)currency_snapshot(.+)end_at IS NOT NULL(.+)").
		WithArgs(uint64(2), "active", "normal", "CNY").
		WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "end_at"}))
	mock.ExpectExec("INSERT INTO `vip_subscriptions`(.+)(currency_snapshot(.+)plan_type_snapshot|plan_type_snapshot(.+)currency_snapshot)(.+)").WillReturnResult(sqlmock.NewResult(8, 1))
	mock.ExpectExec("UPDATE `payment_orders`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE `payment_events`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := s.applyPayment(context.Background(), "late-payment", "TRANSACTION.SUCCESS", []byte(`{}`), successfulTransaction()); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
