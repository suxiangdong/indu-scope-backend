package vip

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	mysqldriver "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
)

var paymentTestDBID atomic.Int64

// The SQL driver is isolated from g.DB's configured production/development
// connection. These tests cannot contact the user's actual MySQL database.
type paymentTestDriver struct {
	*mysqldriver.Driver
	db *sql.DB
}

func (d *paymentTestDriver) New(c *gdb.Core, _ *gdb.ConfigNode) (gdb.DB, error) {
	return &paymentTestDriver{Driver: &mysqldriver.Driver{Core: c}, db: d.db}, nil
}
func (d *paymentTestDriver) Open(_ *gdb.ConfigNode) (*sql.DB, error) { return d.db, nil }
func (d *paymentTestDriver) TableFields(_ context.Context, table string, _ ...string) (map[string]*gdb.TableField, error) {
	fields := map[string]*gdb.TableField{}
	columns := strings.Fields("id user_id plan_id order_no status amount_fen currency payment_method provider provider_transaction_id payment_app_id payment_mch_id payer_openid plan_type_snapshot duration_days_snapshot prepay_id expires_at subscription_id paid_at event_id event_type payload process_status processed_at end_at start_at auto_renew wechat_openid entitlement_start_at entitlement_end_at entitlement_revoked_at refund_no order_id requested_by_admin_id reason total_fen provider_refund_id payer_refund_fen received_account entitlement_status reclaimed_seconds last_error success_at next_billing_at code name description billing_cycle plan_type duration_days price_fen original_price_fen auto_renew_enabled sort_order created_at updated_at agreement_key title content version idempotency_key plan_code_snapshot plan_name_snapshot plan_description_snapshot billing_cycle_snapshot auto_renew_enabled_snapshot original_price_fen_snapshot service_agreement_id service_agreement_version auto_renew_agreement_id auto_renew_agreement_version currency_snapshot")
	for i, name := range columns {
		kind := "varchar(128)"
		if name == "id" || strings.HasSuffix(name, "_id") && name != "payment_app_id" && name != "payment_mch_id" && name != "provider_transaction_id" && name != "provider_refund_id" && name != "prepay_id" && name != "event_id" {
			kind = "bigint unsigned"
		}
		if name == "amount_fen" || name == "duration_days_snapshot" {
			kind = "int unsigned"
		}
		if strings.HasSuffix(name, "_at") {
			kind = "datetime(3)"
		}
		fields[name] = &gdb.TableField{Index: i, Name: name, Type: kind}
	}
	return fields, nil
}

func paymentServiceFixture(t *testing.T) (*sVip, sqlmock.Sqlmock) {
	t.Helper()
	raw, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherFunc(func(expected, actual string) error {
		// Ownership is serialized by an existing users PK record. Locking
		// missing child ranges would let different users share InnoDB gaps
		// and deadlock while each inserts its first subscription/order/refund.
		if strings.Contains(actual, "FOR UPDATE") && (strings.Contains(actual, "`vip_subscriptions`") ||
			strings.Contains(actual, "AND (`idempotency_key`") ||
			(strings.Contains(actual, "`payment_refunds`") && strings.Contains(actual, "WHERE (`order_id`")) ||
			(strings.Contains(actual, "`payment_orders`") && strings.Contains(actual, "`subscription_id`") && strings.Contains(actual, "WHERE (`subscription_id`"))) {
			return fmt.Errorf("child range must use owner-serialized consistent read: %s", actual)
		}
		return sqlmock.QueryMatcherRegexp.Match(expected, actual)
	})))
	if err != nil {
		t.Fatal(err)
	}
	name := fmt.Sprintf("vip_payment_mock_%d", paymentTestDBID.Add(1))
	if err := gdb.Register(name, &paymentTestDriver{Driver: &mysqldriver.Driver{}, db: raw}); err != nil {
		t.Fatal(err)
	}
	db, err := gdb.New(gdb.ConfigNode{Type: name, Name: "vip_test"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = raw.Close() })
	return &sVip{paymentDatabase: db}, mock
}

func successfulTransaction() *payments.Transaction {
	return &payments.Transaction{Appid: core.String("wx_test"), Mchid: core.String("1005166565"), OutTradeNo: core.String("VIPTEST"),
		TradeType: core.String("JSAPI"), TradeState: core.String("SUCCESS"), TransactionId: core.String("WXTRANSACTION"), SuccessTime: core.String("2026-11-17T10:00:00+08:00"),
		Payer: &payments.TransactionPayer{Openid: core.String("openid_test")}, Amount: &payments.TransactionAmount{Total: core.Int64(20000), Currency: core.String("CNY")}}
}

func boundOrder() orderRow {
	return orderRow{ID: 10, UserID: 2, PlanID: 3, OrderNo: "VIPTEST", Status: "paying", AmountFen: 20000, Currency: "CNY", PaymentMethod: "wechat_miniapp", Provider: "wechatpay", PaymentAppID: "wx_test", PaymentMchID: "1005166565", PayerOpenID: "openid_test", PlanTypeSnapshot: "normal", DurationDaysSnapshot: 30}
}

func orderSQLRows(status string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "user_id", "plan_id", "order_no", "status", "amount_fen", "currency", "payment_method", "provider", "payment_app_id", "payment_mch_id", "payer_openid", "plan_type_snapshot", "duration_days_snapshot", "provider_transaction_id"}).
		AddRow(10, 2, 3, "VIPTEST", status, 20000, "CNY", "wechat_miniapp", "wechatpay", "wx_test", "1005166565", "openid_test", "normal", 30, "WXTRANSACTION")
}

func expectPaymentBegin(mock sqlmock.Sqlmock, status string) {
	mock.ExpectExec("INSERT INTO `payment_events`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT (.+) FROM `payment_orders`(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}).AddRow(10, 2))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM `users`(.+)FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectQuery("SELECT (.+) FROM `payment_orders`(.+)FOR UPDATE").WillReturnRows(orderSQLRows(status))
}

func TestPaymentSettlementDoesNotExtendTwice(t *testing.T) {
	s, mock := paymentServiceFixture(t)
	expectPaymentBegin(mock, "paying")
	// A different ordinary billing period can renew this member.
	mock.ExpectQuery("SELECT (.+) FROM `vip_subscriptions`(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "end_at", "plan_plan_type", "plan_currency"}).AddRow(7, 4, time.Now().Add(10*24*time.Hour), "normal", "CNY"))
	mock.ExpectExec("UPDATE `vip_subscriptions` SET .*`plan_id`(.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE `payment_orders`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE `payment_events`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := s.applyPayment(context.Background(), "callback-1", "TRANSACTION.SUCCESS", []byte(`{}`), successfulTransaction()); err != nil {
		t.Fatal(err)
	}
	// Different event IDs are used by active query and callback. The paid
	// order, not only the event ID, prevents a second entitlement update.
	expectPaymentBegin(mock, "paid")
	mock.ExpectExec("UPDATE `payment_events`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := s.applyPayment(context.Background(), "query:WXTRANSACTION", "TRANSACTION.QUERY.SUCCESS", []byte(`{}`), successfulTransaction()); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPaymentSettlementCreatesFirstMembership(t *testing.T) {
	s, mock := paymentServiceFixture(t)
	expectPaymentBegin(mock, "paying")
	mock.ExpectQuery("SELECT (.+) FROM `vip_subscriptions`(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "end_at"}))
	mock.ExpectExec("INSERT INTO `vip_subscriptions`").WillReturnResult(sqlmock.NewResult(8, 1))
	mock.ExpectExec("UPDATE `payment_orders` SET .*`entitlement_start_at`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE `payment_events`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := s.applyPayment(context.Background(), "callback-first", "TRANSACTION.SUCCESS", []byte(`{}`), successfulTransaction()); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPaymentAmountMismatchRollsBack(t *testing.T) {
	s, mock := paymentServiceFixture(t)
	expectPaymentBegin(mock, "paying")
	mock.ExpectRollback()
	mock.ExpectExec("UPDATE `payment_events`").WillReturnResult(sqlmock.NewResult(0, 1))
	transaction := successfulTransaction()
	transaction.Amount.Total = core.Int64(1)
	if err := s.applyPayment(context.Background(), "bad-amount", "TRANSACTION.SUCCESS", []byte(`{}`), transaction); err == nil {
		t.Fatal("wrong payment amount was applied")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPaymentForUnknownOrderRecordsFailure(t *testing.T) {
	s, mock := paymentServiceFixture(t)
	mock.ExpectExec("INSERT INTO `payment_events`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT (.+) FROM `payment_orders`").WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}))
	mock.ExpectExec("UPDATE `payment_events`").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := s.applyPayment(context.Background(), "unknown-order", "TRANSACTION.SUCCESS", []byte(`{}`), successfulTransaction()); err == nil {
		t.Fatal("unknown payment order was applied")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCancelOrderBeforePrepay(t *testing.T) {
	s, mock := paymentServiceFixture(t)
	unbound := func() *sqlmock.Rows {
		return sqlmock.NewRows([]string{"id", "user_id", "order_no", "status"}).AddRow(10, 2, "VIPTEST", "pending")
	}
	mock.ExpectQuery("SELECT (.+) FROM `payment_orders`").WithArgs(uint64(10), uint64(2)).WillReturnRows(unbound())
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM `users`(.+)FOR UPDATE").WithArgs(uint64(2)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectQuery("SELECT (.+) FROM `payment_orders`(.+)FOR UPDATE").WithArgs(uint64(10), uint64(2)).WillReturnRows(unbound())
	mock.ExpectExec("UPDATE `payment_orders`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	order, err := s.CancelOrder(context.Background(), 2, 10)
	if err != nil || order.Status != "closed" {
		t.Fatalf("cancel status=%s, err=%v", order.Status, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCancelPaidOrderDoesNotClose(t *testing.T) {
	s, mock := paymentServiceFixture(t)
	mock.ExpectQuery("SELECT (.+) FROM `payment_orders`").WillReturnRows(orderSQLRows("paying"))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM `users`(.+)FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	// Payment may settle after the initial read but before the cancel lock.
	mock.ExpectQuery("SELECT (.+) FROM `payment_orders`(.+)FOR UPDATE").WillReturnRows(orderSQLRows("paid"))
	mock.ExpectCommit()
	order, err := s.CancelOrder(context.Background(), 2, 10)
	if err != nil || order.Status != "paid" {
		t.Fatalf("cancel status=%s, err=%v", order.Status, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPaidTransactionValidation(t *testing.T) {
	for _, name := range []string{"valid", "app", "merchant", "payer", "amount", "currency", "trade type", "state", "time", "free", "missing snapshot"} {
		t.Run(name, func(t *testing.T) {
			row, tx := boundOrder(), successfulTransaction()
			switch name {
			case "app":
				tx.Appid = core.String("wx_other")
			case "merchant":
				tx.Mchid = core.String("other")
			case "payer":
				tx.Payer.Openid = core.String("other")
			case "amount":
				tx.Amount.Total = core.Int64(1)
			case "currency":
				tx.Amount.Currency = core.String("USD")
			case "trade type":
				tx.TradeType = core.String("NATIVE")
			case "state":
				tx.TradeState = core.String("NOTPAY")
			case "time":
				tx.SuccessTime = nil
			case "free":
				row.AmountFen = 0
			case "missing snapshot":
				row.PaymentAppID = ""
			}
			err := validatePaidTransaction(row, tx)
			if (err == nil) != (name == "valid") {
				t.Fatalf("validation error=%v", err)
			}
		})
	}
}

func TestRenewalUsesRemainingTimeAndSnapshot(t *testing.T) {
	now := time.Date(2026, 11, 17, 0, 0, 0, 0, time.UTC)
	future := now.Add(10 * 24 * time.Hour)
	past := now.Add(-time.Hour)
	row := boundOrder()
	activePlan := func(planID uint64, endAt *time.Time) *subscriptionRow {
		return &subscriptionRow{PlanID: planID, EndAt: endAt, PlanType: "normal", PlanCurrency: "CNY"}
	}
	recurring := activePlan(4, &future)
	recurring.PlanType = "auto_renew"
	for _, test := range []struct {
		name    string
		active  *subscriptionRow
		want    time.Time
		wantErr bool
	}{
		{"first", nil, now.Add(30 * 24 * time.Hour), false},
		{"not expired", activePlan(3, &future), future.Add(30 * 24 * time.Hour), false},
		{"expired", activePlan(3, &past), now.Add(30 * 24 * time.Hour), false},
		{"different plan", activePlan(4, &future), future.Add(30 * 24 * time.Hour), false},
		{"type change cannot reject settlement", recurring, future.Add(30 * 24 * time.Hour), false},
		{"unlimited is not a finite grant", activePlan(3, nil), now.Add(30 * 24 * time.Hour), false},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := membershipEndAt(now, test.active, row)
			if (err != nil) != test.wantErr {
				t.Fatalf("err=%v", err)
			}
			if err == nil && !got.Equal(test.want) {
				t.Fatalf("got %v want %v", got, test.want)
			}
		})
	}
}

func TestOrderLookupIncludesOwner(t *testing.T) {
	s, mock := paymentServiceFixture(t)
	mock.ExpectQuery("SELECT (.+) FROM `payment_orders`(.+)user_id").WithArgs(uint64(10), uint64(99)).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}))
	if _, err := s.loadOrder(context.Background(), 99, 10); err == nil {
		t.Fatal("foreign order was exposed")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestOrderRetryUsesFrozenTermsBeforeCurrentPlan(t *testing.T) {
	for _, planID := range []uint64{3, 99} {
		t.Run(fmt.Sprint(planID), func(t *testing.T) {
			s, mock := paymentServiceFixture(t)
			// The per-user lock precedes key lookup; a concurrent request must
			// recover its predecessor before reading mutable eligibility.
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT (.+) FROM `users`(.+)FOR UPDATE").WithArgs(uint64(2)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
			// No current plan/agreement query is expected: retry must work even
			// when the plan has since changed or been disabled.
			mock.ExpectQuery("SELECT (.+) FROM `payment_orders`(.+)user_id(.+)idempotency_key").WithArgs(uint64(2), "checkout-key").WillReturnRows(orderSQLRows("paying"))
			if planID == 3 {
				mock.ExpectCommit()
			} else {
				mock.ExpectRollback()
			}
			order, reused, err := s.CreateOrder(context.Background(), 2, planID, false, "checkout-key")
			if planID == 3 {
				if err != nil || !reused || order.ID != 10 || order.AmountFen != 20000 {
					t.Fatalf("order=%+v reused=%t err=%v", order, reused, err)
				}
			} else if err == nil {
				t.Fatal("same key accepted for a different plan")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
