package vip

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	authModel "github.com/suxiangdong/indu-scope-backend/internal/model/auth"
	bizctxModel "github.com/suxiangdong/indu-scope-backend/internal/model/bizctx"
	"github.com/suxiangdong/indu-scope-backend/internal/wechatpay"
)

func refundTestOperator(id int64, super bool, status uint8) context.Context {
	return context.WithValue(context.Background(), consts.BizctxContextKey, &bizctxModel.Context{AdminUser: &authModel.AdminUser{ID: id, IsSuperAdmin: super, Status: status}})
}

func refundFixtureOrder() orderRow {
	r := boundOrder()
	r.Status = "paid"
	r.ProviderTransactionID = "WXTRANSACTION"
	r.SubscriptionID = uint64Ptr(7)
	now := time.Now().Truncate(time.Millisecond)
	r.PaidAt = &now
	return r
}
func refundResultFixture(status string) *wechatpay.RefundResult {
	now := time.Now().Truncate(time.Millisecond)
	return &wechatpay.RefundResult{MchID: "1005166565", OrderNo: "VIPTEST", TransactionID: "WXTRANSACTION", RefundNo: "RFTEST", RefundID: "WXREFUND", Status: status, Currency: "CNY", TotalFen: 20000, RefundFen: 20000, SuccessAt: &now}
}
func refundSQLRows(status string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "refund_no", "order_id", "order_no", "user_id", "amount_fen", "total_fen", "currency", "status", "provider_refund_id", "reason"}).AddRow(20, "RFTEST", 10, "VIPTEST", 2, 20000, 20000, "CNY", status, "WXREFUND", "用户申请退款")
}
func refundOrderSQLRows(r orderRow) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "user_id", "plan_id", "order_no", "status", "amount_fen", "currency", "payment_method", "provider", "payment_app_id", "payment_mch_id", "payer_openid", "plan_type_snapshot", "duration_days_snapshot", "provider_transaction_id", "subscription_id", "paid_at", "entitlement_start_at", "entitlement_end_at"}).
		AddRow(r.ID, r.UserID, r.PlanID, r.OrderNo, r.Status, r.AmountFen, r.Currency, r.PaymentMethod, r.Provider, r.PaymentAppID, r.PaymentMchID, r.PayerOpenID, r.PlanTypeSnapshot, r.DurationDaysSnapshot, r.ProviderTransactionID, r.SubscriptionID, r.PaidAt, r.EntitlementStartAt, r.EntitlementEndAt)
}
func expectRefundBegin(mock sqlmock.Sqlmock, status string, order orderRow) {
	mock.ExpectExec("INSERT INTO `payment_events`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT (.+) FROM `payment_refunds`").WillReturnRows(sqlmock.NewRows([]string{"id", "order_id", "user_id"}).AddRow(20, 10, 2))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM `users`(.+)FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectQuery("SELECT (.+) FROM `payment_orders`(.+)FOR UPDATE").WillReturnRows(refundOrderSQLRows(order))
	mock.ExpectQuery("SELECT (.+) FROM `payment_refunds`(.+)FOR UPDATE").WillReturnRows(refundSQLRows(status))
}

func TestRefundProcessingDoesNotChangeMembership(t *testing.T) {
	s, mock := paymentServiceFixture(t)
	expectRefundBegin(mock, "created", refundFixtureOrder())
	mock.ExpectExec("UPDATE `payment_refunds`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE `payment_events`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := s.applyRefundResult(context.Background(), "processing", "REFUND.CREATE.PROCESSING", nil, refundResultFixture("PROCESSING")); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRefundSuccessReclaimsOnceAndProtectsOtherOrders(t *testing.T) {
	s, mock := paymentServiceFixture(t)
	now := time.Now().Truncate(time.Millisecond)
	start, end, final := now.Add(-10*24*time.Hour), now.Add(20*24*time.Hour), now.Add(50*24*time.Hour)
	order := refundFixtureOrder()
	order.EntitlementStartAt, order.EntitlementEndAt = &start, &end
	expectRefundBegin(mock, "processing", order)
	mock.ExpectQuery("SELECT (.+) FROM `vip_subscriptions`(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "plan_id", "end_at"}).AddRow(7, 2, 4, final))
	mock.ExpectQuery("SELECT (.+) FROM `payment_orders`(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "plan_type_snapshot", "currency", "duration_days_snapshot", "entitlement_start_at", "entitlement_end_at"}).AddRow(10, 3, "normal", "CNY", 30, start, end).AddRow(11, 4, "normal", "CNY", 30, end, final))
	mock.ExpectExec("UPDATE `payment_orders` SET .*`entitlement_start_at`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE `vip_subscriptions` SET (.+)(currency_snapshot(.+)plan_type_snapshot|plan_type_snapshot(.+)currency_snapshot)(.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE `payment_orders` SET `entitlement_revoked_at`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE `payment_orders` SET `status`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE `payment_refunds`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE `payment_events`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := s.applyRefundResult(context.Background(), "refund-callback", "REFUND.SUCCESS", nil, refundResultFixture("SUCCESS")); err != nil {
		t.Fatal(err)
	}
	// A second success (active query) or stale processing response cannot remove
	// time again or regress a completed refund.
	order.Status = "refunded"
	for _, state := range []string{"SUCCESS", "PROCESSING"} {
		expectRefundBegin(mock, "success", order)
		mock.ExpectExec("UPDATE `payment_events`").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()
		if err := s.applyRefundResult(context.Background(), "repeat:"+state, "REFUND.QUERY."+state, nil, refundResultFixture(state)); err != nil {
			t.Fatal(err)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRefundLegacyEntitlementAndAmountMismatch(t *testing.T) {
	for _, wrongAmount := range []bool{false, true} {
		t.Run(map[bool]string{false: "manual entitlement review", true: "reject amount"}[wrongAmount], func(t *testing.T) {
			s, mock := paymentServiceFixture(t)
			expectRefundBegin(mock, "processing", refundFixtureOrder())
			result := refundResultFixture("SUCCESS")
			if wrongAmount {
				result.RefundFen = 1
				mock.ExpectRollback()
				mock.ExpectExec("UPDATE `payment_events`").WillReturnResult(sqlmock.NewResult(0, 1))
			} else {
				// Missing historical allocations do not revoke another order's VIP.
				mock.ExpectExec("UPDATE `payment_orders`").WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec("UPDATE `payment_refunds` SET .*`entitlement_status`").WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec("UPDATE `payment_events`").WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			}
			err := s.applyRefundResult(context.Background(), "refund-result", "REFUND.SUCCESS", nil, result)
			if (err != nil) != wrongAmount {
				t.Fatalf("refund err=%v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRefundSettlementRollsBackWhenEntitlementWriteFails(t *testing.T) {
	s, mock := paymentServiceFixture(t)
	now := time.Now().Truncate(time.Millisecond)
	start, end := now.Add(-24*time.Hour), now.Add(29*24*time.Hour)
	order := refundFixtureOrder()
	order.EntitlementStartAt, order.EntitlementEndAt = &start, &end
	expectRefundBegin(mock, "processing", order)
	mock.ExpectQuery("SELECT (.+) FROM `vip_subscriptions`(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "end_at"}).AddRow(7, 2, end))
	mock.ExpectQuery("SELECT (.+) FROM `payment_orders`(.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "duration_days_snapshot", "entitlement_start_at", "entitlement_end_at"}).AddRow(10, 3, 30, start, end))
	mock.ExpectExec("UPDATE `vip_subscriptions`").WillReturnError(errors.New("database write unavailable"))
	mock.ExpectRollback()
	mock.ExpectExec("UPDATE `payment_events`").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := s.applyRefundResult(context.Background(), "refund-rollback", "REFUND.SUCCESS", nil, refundResultFixture("SUCCESS")); err == nil {
		t.Fatal("failed entitlement transaction was acknowledged")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

type failingRefundGateway struct{ calls []wechatpay.RefundInput }

func (f *failingRefundGateway) CreateRefund(_ context.Context, in wechatpay.RefundInput) (*wechatpay.RefundResult, error) {
	f.calls = append(f.calls, in)
	return nil, errors.New("network timeout after request may have reached WeChat")
}
func (f *failingRefundGateway) QueryRefund(context.Context, string) (*wechatpay.RefundResult, error) {
	return nil, errors.New("not found")
}
func (f *failingRefundGateway) RefundCallbackURL() (string, error) {
	return "https://api.example.com/api/v1/payment/wechat/refund/notify", nil
}

func TestRefundRetryUsesSameFrozenRefundNumber(t *testing.T) {
	s, mock := paymentServiceFixture(t)
	gateway := &failingRefundGateway{}
	s.refundGateway = gateway
	s.refundConfig = wechatpay.Config{AppID: "wx_test", MchID: "1005166565"}
	for i := 0; i < 2; i++ {
		mock.ExpectQuery("SELECT (.+) FROM `payment_refunds`").WillReturnRows(refundSQLRows("created"))
		mock.ExpectQuery("SELECT (.+) FROM `payment_orders`").WillReturnRows(refundOrderSQLRows(refundFixtureOrder()))
		mock.ExpectExec("UPDATE `payment_refunds`").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery("SELECT (.+) FROM `payment_refunds`").WillReturnRows(refundSQLRows("created"))
		if _, err := s.RetryRefund(refundTestOperator(1, true, 1), 20); err != nil {
			t.Fatal(err)
		}
	}
	if len(gateway.calls) != 2 || gateway.calls[0].RefundNo != "RFTEST" || gateway.calls[0] != gateway.calls[1] {
		t.Fatal("network retry changed the refund identity or amount")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRefundDuplicateCreateReturnsExistingProcessingRefund(t *testing.T) {
	s, mock := paymentServiceFixture(t)
	gateway := &failingRefundGateway{}
	s.refundGateway = gateway
	mock.ExpectQuery("SELECT (.+) FROM `payment_orders`").WillReturnRows(refundOrderSQLRows(refundFixtureOrder()))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM `users`(.+)FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectQuery("SELECT (.+) FROM `payment_orders`(.+)FOR UPDATE").WillReturnRows(refundOrderSQLRows(refundFixtureOrder()))
	mock.ExpectQuery("SELECT (.+) FROM `payment_refunds`(.+)order_id").WillReturnRows(refundSQLRows("processing"))
	mock.ExpectCommit()
	mock.ExpectQuery("SELECT (.+) FROM `payment_refunds`").WillReturnRows(refundSQLRows("processing"))
	r, err := s.CreateRefund(refundTestOperator(1, true, 1), 10, 1, "重复点击退款")
	if err != nil || r.ID != 20 || len(gateway.calls) != 0 {
		t.Fatalf("duplicate refund result=%v err=%v", r, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRefundOperationsRequireActiveSuperAdmin(t *testing.T) {
	for _, test := range []struct {
		name string
		ctx  context.Context
	}{
		{"unauthenticated", context.Background()},
		{"ordinary administrator", refundTestOperator(1, false, 1)},
		{"disabled super administrator", refundTestOperator(1, true, 0)},
	} {
		t.Run(test.name, func(t *testing.T) {
			s, mock := paymentServiceFixture(t)
			gateway := &failingRefundGateway{}
			s.refundGateway = gateway
			_, createErr := s.CreateRefund(test.ctx, 10, 1, "退款权限测试")
			_, retryErr := s.RetryRefund(test.ctx, 20)
			if gerror.Code(createErr) != gcode.CodeNotAuthorized || gerror.Code(retryErr) != gcode.CodeNotAuthorized {
				t.Fatalf("authorization errors: create=%v retry=%v", createErr, retryErr)
			}
			if len(gateway.calls) != 0 {
				t.Fatal("unauthorized operation reached WeChat")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRefundOperatorCannotBeSpoofed(t *testing.T) {
	s, mock := paymentServiceFixture(t)
	_, err := s.CreateRefund(refundTestOperator(2, true, 1), 10, 1, "冒用其他管理员")
	if gerror.Code(err) != gcode.CodeNotAuthorized {
		t.Fatalf("spoofed administrator was accepted: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRefundableOrderChecksAndIdentity(t *testing.T) {
	for _, name := range []string{"valid", "free", "unpaid", "expired", "wrong merchant", "other method", "other transaction", "amount", "refund number", "missing success time"} {
		t.Run(name, func(t *testing.T) {
			order := refundFixtureOrder()
			row := refundRow{ID: 20, RefundNo: "RFTEST", OrderID: 10, OrderNo: "VIPTEST", UserID: 2, AmountFen: 20000, TotalFen: 20000, Currency: "CNY"}
			result := refundResultFixture("SUCCESS")
			switch name {
			case "free":
				order.AmountFen = 0
			case "unpaid":
				order.Status = "pending"
			case "expired":
				old := time.Now().Add(-366 * 24 * time.Hour)
				order.PaidAt = &old
			case "wrong merchant":
				order.PaymentMchID = "other"
			case "other method":
				order.PaymentMethod = "free"
			case "other transaction":
				result.TransactionID = "OTHER"
			case "amount":
				result.RefundFen = 1
			case "refund number":
				result.RefundNo = "OTHER"
			case "missing success time":
				result.SuccessAt = nil
			}
			err := validateRefundableOrder(order, wechatpay.Config{AppID: "wx_test", MchID: "1005166565"}, time.Now())
			if err == nil {
				err = validateRefundResult(order, row, result)
			}
			if (err == nil) != (name == "valid") {
				t.Fatalf("validation err=%v", err)
			}
		})
	}
}
