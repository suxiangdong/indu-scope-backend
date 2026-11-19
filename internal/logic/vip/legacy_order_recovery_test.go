package vip

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/suxiangdong/indu-scope-backend/internal/testutil/dbmock"
	"github.com/suxiangdong/indu-scope-backend/internal/wechatpay"
)

const legacyOrderNo = "VIP2026111812345600112233445566778899aabb"

func legacyOrder(status string) orderRow {
	return orderRow{ID: 10, UserID: 2, OrderNo: legacyOrderNo, Status: status,
		AmountFen: 20000, Currency: "CNY", DurationDaysSnapshot: 30, PlanTypeSnapshot: "normal",
		Provider: "wechatpay", PaymentMethod: "wechat_miniapp", PaymentAppID: "wx_test", PaymentMchID: "merchant_test"}
}

func legacyRecoveryFixture(t *testing.T) (*sVip, sqlmock.Sqlmock) {
	t.Helper()
	db, mock := dbmock.New(t, "id user_id order_no status amount_fen currency duration_days_snapshot plan_type_snapshot provider payment_method payment_app_id payment_mch_id prepay_id provider_transaction_id paid_at subscription_id entitlement_start_at entitlement_end_at entitlement_revoked_at closed_at updated_at")
	return &sVip{paymentDatabase: db}, mock
}

func legacyOrderRows(row orderRow) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "user_id", "order_no", "status", "amount_fen", "currency", "duration_days_snapshot", "plan_type_snapshot", "provider", "payment_method", "payment_app_id", "payment_mch_id", "prepay_id", "provider_transaction_id", "paid_at", "subscription_id", "entitlement_start_at", "entitlement_end_at", "entitlement_revoked_at"}).
		AddRow(row.ID, row.UserID, row.OrderNo, row.Status, row.AmountFen, row.Currency, row.DurationDaysSnapshot, row.PlanTypeSnapshot, row.Provider, row.PaymentMethod, row.PaymentAppID, row.PaymentMchID, row.PrepayID, row.ProviderTransactionID, row.PaidAt, row.SubscriptionID, row.EntitlementStartAt, row.EntitlementEndAt, row.EntitlementRevokedAt)
}

func expectLegacyRead(mock sqlmock.Sqlmock, row orderRow) {
	mock.ExpectQuery("SELECT (.+) FROM `payment_orders` WHERE (.+)").WithArgs(uint64(10), uint64(2)).WillReturnRows(legacyOrderRows(row))
}

func expectLegacyCancelLock(mock sqlmock.Sqlmock, row orderRow) {
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM `users`(.+)FOR UPDATE").WithArgs(uint64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectQuery("SELECT (.+) FROM `payment_orders`(.+)FOR UPDATE").WithArgs(uint64(10), uint64(2)).WillReturnRows(legacyOrderRows(row))
}

func expectLegacyLocalClose(mock sqlmock.Sqlmock) {
	// Only the closure markers and normal update timestamp may change. In
	// particular, the order number, snapshots and membership remain intact.
	mock.ExpectExec("UPDATE `payment_orders` SET `status`=\\?,`closed_at`=\\?,`updated_at`=\\? WHERE (.+)").
		WithArgs("closed", sqlmock.AnyArg(), sqlmock.AnyArg(), uint64(10)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
}

func TestCancelLegacyUnpaidOrderWithoutPaymentConfiguration(t *testing.T) {
	for _, status := range []string{"pending", "paying"} {
		for _, bound := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/bound=%t", status, bound), func(t *testing.T) {
				s, mock := legacyRecoveryFixture(t)
				row := legacyOrder(status)
				if !bound {
					row.PaymentAppID, row.PaymentMchID = "", ""
				}
				expectLegacyRead(mock, row)
				expectLegacyCancelLock(mock, row)
				expectLegacyLocalClose(mock)
				order, err := s.CancelOrder(context.Background(), 2, 10)
				if err != nil || order.Status != "closed" || order.OrderNo != legacyOrderNo {
					t.Fatalf("legacy cancellation: status=%s number=%s err=%v", order.Status, order.OrderNo, err)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestLegacyRecoveryRejectsPaymentOrEntitlementEvidence(t *testing.T) {
	now := time.Now()
	subscriptionID := uint64(7)
	cases := map[string]func(*orderRow){
		"prepay":       func(row *orderRow) { row.PrepayID = "prepay_exists" },
		"transaction":  func(row *orderRow) { row.ProviderTransactionID = "transaction_exists" },
		"paid_at":      func(row *orderRow) { row.PaidAt = &now },
		"subscription": func(row *orderRow) { row.SubscriptionID = &subscriptionID },
		"grant_start":  func(row *orderRow) { row.EntitlementStartAt = &now },
		"grant_end":    func(row *orderRow) { row.EntitlementEndAt = &now },
		"revocation":   func(row *orderRow) { row.EntitlementRevokedAt = &now },
		"provider":     func(row *orderRow) { row.Provider = "manual" },
		"method":       func(row *orderRow) { row.PaymentMethod = "free" },
		"failed":       func(row *orderRow) { row.Status = "failed" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			s, mock := legacyRecoveryFixture(t)
			row := legacyOrder("paying")
			// Even an inconsistent unbound record with payment evidence must not
			// fall through to the ordinary local-cancel path.
			row.PaymentAppID, row.PaymentMchID = "", ""
			mutate(&row)
			// The initial snapshot looked safe; only the locked row reveals
			// evidence written by a concurrent payment/settlement operation.
			expectLegacyRead(mock, legacyOrder("paying"))
			expectLegacyCancelLock(mock, row)
			mock.ExpectRollback()
			if _, err := s.CancelOrder(context.Background(), 2, 10); !errors.Is(err, wechatpay.ErrInvalidOrderNo) {
				t.Fatalf("legacy order with %s evidence was changed: %v", name, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestLegacyRecoveryUsesLockedStateAndPreservesEndedOrders(t *testing.T) {
	for _, ended := range []string{"paid", "closed", "refunded"} {
		t.Run(ended, func(t *testing.T) {
			s, mock := legacyRecoveryFixture(t)
			expectLegacyRead(mock, legacyOrder("paying"))
			expectLegacyCancelLock(mock, legacyOrder(ended))
			mock.ExpectCommit()
			order, err := s.CancelOrder(context.Background(), 2, 10)
			if err != nil || order.Status != ended || order.OrderNo != legacyOrderNo {
				t.Fatalf("locked terminal order: status=%s err=%v", order.Status, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSyncLegacyOrderDelegatesToLockedCancellation(t *testing.T) {
	s, mock := legacyRecoveryFixture(t)
	row := legacyOrder("paying")
	expectLegacyRead(mock, row) // SyncOrder's initial read.
	expectLegacyRead(mock, row) // CancelOrder checks ownership again.
	expectLegacyCancelLock(mock, row)
	expectLegacyLocalClose(mock)
	order, err := s.SyncOrder(context.Background(), 2, 10)
	if err != nil || order.Status != "closed" || order.OrderNo != legacyOrderNo {
		t.Fatalf("legacy synchronization: status=%s err=%v", order.Status, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSyncLegacyEndedOrderDoesNotAttemptRecovery(t *testing.T) {
	for _, status := range []string{"paid", "closed", "refunded"} {
		t.Run(status, func(t *testing.T) {
			s, mock := legacyRecoveryFixture(t)
			expectLegacyRead(mock, legacyOrder(status))
			order, err := s.SyncOrder(context.Background(), 2, 10)
			if err != nil || order.Status != status {
				t.Fatalf("ended order synchronization: status=%s err=%v", order.Status, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestLegacyRecoveryOnlyRecognizesTheHistoricalFormat(t *testing.T) {
	for _, number := range []string{strings.Repeat("A", 41), strings.ToUpper(legacyOrderNo), legacyOrderNo[:40], legacyOrderNo + "0", "VIPTEST"} {
		row := legacyOrder("pending")
		row.OrderNo = number
		if canCloseLegacyUnpayableOrder(row) {
			t.Fatalf("unrecognized legacy order format: %s", number)
		}
	}
}

func TestPreparePaymentRejectsInvalidOrderNoBeforeLoadingPaymentClient(t *testing.T) {
	s, mock := legacyRecoveryFixture(t)
	expectLegacyRead(mock, legacyOrder("paying"))
	if _, err := s.PreparePayment(context.Background(), 2, 10); !errors.Is(err, wechatpay.ErrInvalidOrderNo) {
		t.Fatalf("invalid merchant order number reached payment client: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPreparePaidLegacyOrderStillReportsActivation(t *testing.T) {
	s, mock := legacyRecoveryFixture(t)
	expectLegacyRead(mock, legacyOrder("paid"))
	checkout, err := s.PreparePayment(context.Background(), 2, 10)
	if err != nil || checkout == nil || checkout.Order.Status != "paid" || checkout.Payment != nil {
		t.Fatalf("paid legacy checkout: %+v err=%v", checkout, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
