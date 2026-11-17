package vip

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"

	vipModel "github.com/suxiangdong/indu-scope-backend/internal/model/vip"
	"github.com/suxiangdong/indu-scope-backend/internal/wechatpay"
)

func (s *sVip) paymentDB() gdb.DB {
	if s.paymentDatabase != nil {
		return s.paymentDatabase
	}
	return g.DB()
}

func (s *sVip) loadOrder(ctx context.Context, userID, orderID uint64) (*orderRow, error) {
	if userID == 0 || orderID == 0 {
		return nil, errors.New("会员订单参数无效")
	}
	var row *orderRow
	err := s.paymentDB().Model("payment_orders").Ctx(ctx).Where("id", orderID).Where("user_id", userID).Scan(&row)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && row == nil) {
		return nil, errors.New("会员订单不存在")
	}
	if err != nil {
		return nil, gerror.Wrap(err, "load vip payment order")
	}
	return row, nil
}

func (s *sVip) PreparePayment(ctx context.Context, userID, orderID uint64) (*vipModel.Checkout, error) {
	row, err := s.loadOrder(ctx, userID, orderID)
	if err != nil {
		return nil, err
	}
	if row.Status == "paid" {
		return &vipModel.Checkout{Order: row.order()}, nil
	}
	if row.DurationDaysSnapshot == 0 || row.DurationDaysSnapshot > maxMembershipDurationDays {
		return nil, errors.New("订单有效天数快照无效，不能发起支付")
	}
	client, err := wechatpay.Load(ctx)
	if errors.Is(err, wechatpay.ErrDisabled) {
		return &vipModel.Checkout{Order: row.order()}, nil
	}
	if err != nil {
		return nil, err
	}
	if row.PlanTypeSnapshot != "normal" {
		return nil, errors.New("连续包月需要独立签约，暂不可使用普通支付")
	}
	if row.AmountFen == 0 || row.Provider != "wechatpay" {
		return nil, errors.New("此订单不支持微信支付")
	}
	if row.Status != "pending" && row.Status != "paying" {
		return nil, errors.New("此订单已不可支付")
	}
	if row.ExpiresAt == nil || !row.ExpiresAt.After(time.Now()) {
		updated, syncErr := s.SyncOrder(ctx, userID, orderID)
		err = syncErr
		if err != nil {
			return nil, err
		}
		if updated.Status == "paid" {
			return &vipModel.Checkout{Order: updated}, nil
		}
		return nil, errors.New("订单已过期，请重新选择套餐")
	}

	// Bind the identity before the network request. A timeout or a process exit
	// can then be recovered with the same order number and the persisted identity.
	err = s.paymentDB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var user struct {
			ID     uint64 `orm:"id"`
			OpenID string `orm:"wechat_openid"`
		}
		if err := tx.Model("users").Fields("id, wechat_openid").Where("id", userID).LockUpdate().Scan(&user); err != nil {
			return err
		}
		if user.ID == 0 || user.OpenID == "" {
			return errors.New("请使用当前微信账号重新登录后支付")
		}
		var locked orderRow
		if err := tx.Model("payment_orders").Where("id", orderID).Where("user_id", userID).LockUpdate().Scan(&locked); err != nil {
			return err
		}
		if locked.Status != "pending" && locked.Status != "paying" {
			return errors.New("订单状态已更新，请刷新会员中心")
		}
		active, err := loadPaymentSubscription(ctx, tx, userID)
		if err != nil {
			return err
		}
		if err := validatePaidRenewal(active, locked); err != nil {
			return err
		}
		// Keep a single in-flight checkout per user. Repeated calls for the same
		// order reuse its prepay_id; a second order must wait until it is closed.
		count, err := tx.Model("payment_orders").Where("user_id", userID).Where("id <> ?", orderID).Where("status", "paying").Where("expires_at > NOW()").Count()
		if err != nil {
			return err
		}
		if count > 0 {
			return errors.New("还有待完成的支付订单，请先完成该订单或等待其过期")
		}
		if locked.PaymentAppID != "" && (locked.PaymentAppID != client.Config.AppID || locked.PaymentMchID != client.Config.MchID || locked.PayerOpenID != user.OpenID) {
			return errors.New("订单支付身份已变化，请重新下单")
		}
		_, err = tx.Model("payment_orders").Where("id", orderID).Data(g.Map{
			"payment_app_id": client.Config.AppID, "payment_mch_id": client.Config.MchID,
			"payer_openid": user.OpenID, "status": "paying",
		}).Update()
		if err != nil {
			return err
		}
		locked.PaymentAppID, locked.PaymentMchID, locked.PayerOpenID = client.Config.AppID, client.Config.MchID, user.OpenID
		locked.Status = "paying"
		row = &locked
		return nil
	})
	if err != nil {
		return nil, err
	}
	var payment *wechatpay.Payment
	// Serialize prepay and close too: otherwise a cancelled order could be
	// created upstream after the close call returned ORDER_NOT_EXIST.
	err = s.paymentDB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var user struct {
			ID uint64 `orm:"id"`
		}
		if err := tx.Model("users").Fields("id").Where("id", userID).LockUpdate().Scan(&user); err != nil {
			return err
		}
		var latest orderRow
		if err := tx.Model("payment_orders").Where("id", orderID).Where("user_id", userID).LockUpdate().Scan(&latest); err != nil {
			return err
		}
		row = &latest
		if latest.Status == "paid" {
			return nil
		}
		if latest.Status != "paying" {
			return errors.New("订单已不可支付")
		}
		if latest.ExpiresAt == nil || !latest.ExpiresAt.After(time.Now()) {
			return errors.New("订单已过期，请刷新订单状态")
		}
		active, err := loadPaymentSubscription(ctx, tx, userID)
		if err != nil {
			return err
		}
		if err := validatePaidRenewal(active, latest); err != nil {
			return err
		}
		if latest.PrepayID != "" {
			payment, err = client.SignPayment(ctx, latest.PrepayID)
			return err
		}
		payment, err = client.Prepay(ctx, wechatpay.PrepayInput{OrderNo: latest.OrderNo, Description: "源仓会员-" + latest.PlanNameSnapshot,
			OpenID: latest.PayerOpenID, Currency: latest.Currency, AmountFen: int64(latest.AmountFen), ExpiresAt: *latest.ExpiresAt})
		if err != nil {
			return err
		}
		_, err = tx.Model("payment_orders").Where("id", orderID).Data(g.Map{"prepay_id": payment.PrepayID}).Update()
		return err
	})
	if err != nil {
		var apiErr *core.APIError
		if errors.As(err, &apiErr) && apiErr.Code == "ORDERPAID" {
			order, syncErr := s.SyncOrder(ctx, userID, orderID)
			if syncErr == nil && order.Status == "paid" {
				return &vipModel.Checkout{Order: order}, nil
			}
		}
		return nil, gerror.Wrap(err, "prepare WeChat Pay checkout")
	}
	if row.Status == "paid" {
		return &vipModel.Checkout{Order: row.order()}, nil
	}
	return &vipModel.Checkout{Order: row.order(), Payment: &vipModel.PaymentParams{TimeStamp: payment.TimeStamp,
		NonceStr: payment.NonceStr, Package: payment.Package, SignType: payment.SignType, PaySign: payment.PaySign}}, nil
}

func (s *sVip) SyncOrder(ctx context.Context, userID, orderID uint64) (vipModel.Order, error) {
	row, err := s.loadOrder(ctx, userID, orderID)
	if err != nil {
		return vipModel.Order{}, err
	}
	if row.Status == "paid" || row.Status == "closed" || row.Status == "refunded" || row.PaymentAppID == "" {
		return row.order(), nil
	}
	client, err := wechatpay.Load(ctx)
	if err != nil {
		return vipModel.Order{}, err
	}
	if row.PaymentAppID != client.Config.AppID || row.PaymentMchID != client.Config.MchID {
		return vipModel.Order{}, errors.New("订单商户配置不匹配")
	}
	transaction, err := client.Query(ctx, row.OrderNo)
	if err != nil {
		var apiErr *core.APIError
		if errors.As(err, &apiErr) && apiErr.Code == "ORDER_NOT_EXIST" {
			// An upstream request may have failed before creating the order.
			if row.ExpiresAt != nil && !row.ExpiresAt.After(time.Now()) {
				_, err = s.paymentDB().Model("payment_orders").Ctx(ctx).Where("id", orderID).Where("status IN ('pending','paying')").Data(g.Map{"status": "closed", "closed_at": time.Now()}).Update()
				if err != nil {
					return vipModel.Order{}, err
				}
				row.Status = "closed"
			}
			return row.order(), nil
		}
		return vipModel.Order{}, err
	}
	if transaction == nil || textValue(transaction.Appid) != row.PaymentAppID || textValue(transaction.Mchid) != row.PaymentMchID || textValue(transaction.OutTradeNo) != row.OrderNo {
		return vipModel.Order{}, errors.New("payment query identity mismatch")
	}
	if textValue(transaction.TradeState) == "SUCCESS" {
		if err := validatePaidTransaction(*row, transaction); err != nil {
			return vipModel.Order{}, err
		}
	}
	switch textValue(transaction.TradeState) {
	case "SUCCESS":
		payload, err := json.Marshal(transaction)
		if err != nil {
			return vipModel.Order{}, err
		}
		if err := s.applyPayment(ctx, "query:"+textValue(transaction.TransactionId), "TRANSACTION.QUERY.SUCCESS", payload, transaction); err != nil {
			return vipModel.Order{}, err
		}
	case "NOTPAY", "USERPAYING":
		if row.ExpiresAt != nil && !row.ExpiresAt.After(time.Now()) && textValue(transaction.TradeState) == "NOTPAY" {
			if err := client.Close(ctx, row.OrderNo); err != nil {
				return vipModel.Order{}, err
			}
			if err := s.markOrderClosed(ctx, row.ID); err != nil {
				return vipModel.Order{}, err
			}
		}
	case "CLOSED", "REVOKED", "PAYERROR":
		if err := s.markOrderClosed(ctx, row.ID); err != nil {
			return vipModel.Order{}, err
		}
	}
	updated, err := s.loadOrder(ctx, userID, orderID)
	if err != nil {
		return vipModel.Order{}, err
	}
	return updated.order(), nil
}

func (s *sVip) markOrderClosed(ctx context.Context, orderID uint64) error {
	_, err := s.paymentDB().Model("payment_orders").Ctx(ctx).Where("id", orderID).Where("status IN ('pending','paying')").Data(g.Map{"status": "closed", "closed_at": time.Now()}).Update()
	return err
}

func (s *sVip) CancelOrder(ctx context.Context, userID, orderID uint64) (vipModel.Order, error) {
	row, err := s.loadOrder(ctx, userID, orderID)
	if err != nil {
		return vipModel.Order{}, err
	}
	needsSync := false
	err = s.paymentDB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var user struct {
			ID uint64 `orm:"id"`
		}
		if err := tx.Model("users").Fields("id").Where("id", userID).LockUpdate().Scan(&user); err != nil {
			return err
		}
		var current orderRow
		if err := tx.Model("payment_orders").Where("id", orderID).Where("user_id", userID).LockUpdate().Scan(&current); err != nil {
			return err
		}
		row = &current
		if current.Status == "paid" || current.Status == "closed" || current.Status == "refunded" {
			return nil
		}
		if current.PaymentAppID != "" {
			client, err := wechatpay.Load(ctx)
			if err != nil {
				return err
			}
			if current.PaymentAppID != client.Config.AppID || current.PaymentMchID != client.Config.MchID {
				return errors.New("订单商户配置不匹配")
			}
			if err := client.Close(ctx, current.OrderNo); err != nil {
				var apiErr *core.APIError
				if errors.As(err, &apiErr) && apiErr.Code == "ORDERPAID" {
					needsSync = true
					return nil
				}
				if !errors.As(err, &apiErr) || apiErr.Code != "ORDER_NOT_EXIST" {
					return err
				}
			}
		}
		if _, err := tx.Model("payment_orders").Where("id", orderID).Data(g.Map{"status": "closed", "closed_at": time.Now()}).Update(); err != nil {
			return err
		}
		row.Status = "closed"
		return nil
	})
	if err != nil {
		return vipModel.Order{}, err
	}
	if needsSync {
		return s.SyncOrder(ctx, userID, orderID)
	}
	return row.order(), nil
}

func (s *sVip) HandleWechatPaymentNotification(ctx context.Context, request *http.Request) error {
	client, err := wechatpay.Load(ctx)
	if err != nil {
		return err
	}
	notification, err := client.ParseNotification(ctx, request)
	if err != nil {
		return err
	}
	if textValue(notification.Transaction.Appid) != client.Config.AppID || textValue(notification.Transaction.Mchid) != client.Config.MchID {
		return errors.New("notification merchant or app mismatch")
	}
	return s.applyPayment(ctx, notification.ID, "TRANSACTION.SUCCESS", notification.Payload, notification.Transaction)
}

func (s *sVip) applyPayment(ctx context.Context, eventID, eventType string, payload []byte, transaction *payments.Transaction) (resultErr error) {
	if transaction == nil || eventID == "" || textValue(transaction.OutTradeNo) == "" {
		return errors.New("invalid payment event")
	}
	// Persist failed processing attempts too. Repeated delivery never resets a
	// processed event, and the order lock is the final exactly-once boundary.
	if _, err := s.paymentDB().Model("payment_events").Ctx(ctx).Data(g.Map{
		"provider": "wechatpay", "event_id": eventID, "order_no": textValue(transaction.OutTradeNo),
		"event_type": eventType, "payload": string(payload), "process_status": "pending",
	}).OnDuplicate(g.Map{"event_id": eventID}).Insert(); err != nil {
		return err
	}
	defer func() {
		if resultErr != nil {
			_, _ = s.paymentDB().Model("payment_events").Ctx(ctx).Where("provider", "wechatpay").Where("event_id", eventID).Where("process_status <> 'processed'").Data(g.Map{"process_status": "failed"}).Update()
		}
	}()
	var hint *orderRow
	if err := s.paymentDB().Model("payment_orders").Ctx(ctx).Where("order_no", textValue(transaction.OutTradeNo)).Scan(&hint); err != nil {
		return err
	}
	if hint == nil {
		return errors.New("payment order not found")
	}
	err := s.paymentDB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// Lock user first: callbacks for different orders of the same user must
		// serialize entitlement updates, not just callbacks for the same order.
		var user struct {
			ID uint64 `orm:"id"`
		}
		if err := tx.Model("users").Fields("id").Where("id", hint.UserID).LockUpdate().Scan(&user); err != nil {
			return err
		}
		var row orderRow
		if err := tx.Model("payment_orders").Where("id", hint.ID).LockUpdate().Scan(&row); err != nil {
			return err
		}
		if err := validatePaidTransaction(row, transaction); err != nil {
			return err
		}
		if row.Status == "paid" || row.Status == "refunded" {
			if row.ProviderTransactionID != textValue(transaction.TransactionId) {
				return errors.New("payment transaction id changed")
			}
			return finishPaymentEvent(tx, eventID)
		}
		// Settlement must honor a verified successful charge, even when a
		// different membership was activated after checkout. Extend only a
		// compatible ordinary finite subscription; retain other entitlements.
		active, err := loadSettlementSubscription(ctx, tx, row.UserID, row.Currency)
		if err != nil {
			return err
		}
		now := time.Now().Truncate(time.Millisecond)
		endAt, err := membershipEndAt(now, active, row)
		if err != nil {
			return err
		}
		var subscriptionID uint64
		if active != nil {
			subscriptionID = active.ID
			if _, err := tx.Model("vip_subscriptions").Where("id", active.ID).Data(g.Map{"end_at": endAt, "plan_id": row.PlanID, "plan_type_snapshot": row.PlanTypeSnapshot, "currency_snapshot": row.Currency, "auto_renew": 0, "provider": "wechatpay"}).Update(); err != nil {
				return err
			}
		} else {
			id, err := tx.Model("vip_subscriptions").Data(g.Map{"user_id": row.UserID, "plan_id": row.PlanID, "plan_type_snapshot": row.PlanTypeSnapshot, "currency_snapshot": row.Currency, "status": "active", "auto_renew": 0, "provider": "wechatpay", "start_at": now, "end_at": endAt}).InsertAndGetId()
			if err != nil {
				return err
			}
			subscriptionID = uint64(id)
		}
		paidAt, _ := time.Parse(time.RFC3339, textValue(transaction.SuccessTime))
		grantStart := endAt.Add(-time.Duration(row.DurationDaysSnapshot) * 24 * time.Hour)
		if _, err := tx.Model("payment_orders").Where("id", row.ID).Data(g.Map{"status": "paid", "subscription_id": subscriptionID, "provider_transaction_id": textValue(transaction.TransactionId), "paid_at": paidAt, "entitlement_start_at": grantStart, "entitlement_end_at": endAt}).Update(); err != nil {
			return err
		}
		return finishPaymentEvent(tx, eventID)
	})
	return err
}

func finishPaymentEvent(tx gdb.TX, eventID string) error {
	_, err := tx.Model("payment_events").Where("provider", "wechatpay").Where("event_id", eventID).Data(g.Map{"process_status": "processed", "processed_at": time.Now()}).Update()
	return err
}

func loadPaymentSubscription(ctx context.Context, tx gdb.TX, userID uint64) (*subscriptionRow, error) {
	return scanPaymentSubscription(activePaymentSubscriptions(ctx, tx, userID))
}

func loadSettlementSubscription(ctx context.Context, tx gdb.TX, userID uint64, currency string) (*subscriptionRow, error) {
	return scanPaymentSubscription(activePaymentSubscriptions(ctx, tx, userID).
		Where("s.plan_type_snapshot", "normal").Where("s.currency_snapshot", currency).Where("s.end_at IS NOT NULL"))
}

func activePaymentSubscriptions(ctx context.Context, tx gdb.TX, userID uint64) *gdb.Model {
	return tx.Model("vip_subscriptions s").Ctx(ctx).
		Fields("s.*, s.plan_type_snapshot AS plan_plan_type, s.currency_snapshot AS plan_currency").
		Where("s.user_id", userID).Where("s.status", "active").Where("(s.end_at IS NULL OR s.end_at > NOW())")
}

func scanPaymentSubscription(query *gdb.Model) (*subscriptionRow, error) {
	var active *subscriptionRow
	// The caller holds users.id before its first consistent read. Avoid
	// locking empty child/index gaps: different users can share those gaps.
	err := query.Order("s.end_at IS NULL DESC, s.end_at DESC, s.id DESC").Limit(1).Scan(&active)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return active, err
}

func validateTransactionIdentity(row orderRow, t *payments.Transaction) error {
	if t == nil || row.PaymentAppID == "" || row.PaymentMchID == "" || row.PayerOpenID == "" || row.Provider != "wechatpay" || row.PaymentMethod != "wechat_miniapp" {
		return errors.New("payment identity snapshot missing")
	}
	if textValue(t.Appid) != row.PaymentAppID || textValue(t.Mchid) != row.PaymentMchID || textValue(t.OutTradeNo) != row.OrderNo || textValue(t.TradeType) != "JSAPI" {
		return errors.New("payment merchant, app, order or trade type mismatch")
	}
	if t.Payer == nil || textValue(t.Payer.Openid) != row.PayerOpenID {
		return errors.New("payment payer mismatch")
	}
	if t.Amount == nil || t.Amount.Total == nil || *t.Amount.Total != int64(row.AmountFen) || textValue(t.Amount.Currency) != row.Currency {
		return errors.New("payment amount or currency mismatch")
	}
	return nil
}

func validatePaidTransaction(row orderRow, t *payments.Transaction) error {
	if err := validateTransactionIdentity(row, t); err != nil {
		return err
	}
	if textValue(t.TradeState) != "SUCCESS" || textValue(t.TransactionId) == "" || row.AmountFen == 0 || row.PlanTypeSnapshot != "normal" || row.DurationDaysSnapshot == 0 {
		return errors.New("invalid successful membership payment")
	}
	if _, err := time.Parse(time.RFC3339, textValue(t.SuccessTime)); err != nil {
		return fmt.Errorf("invalid payment success time: %w", err)
	}
	return nil
}

func membershipEndAt(now time.Time, active *subscriptionRow, row orderRow) (time.Time, error) {
	if row.DurationDaysSnapshot == 0 || row.DurationDaysSnapshot > maxMembershipDurationDays {
		return time.Time{}, errors.New("invalid membership duration snapshot")
	}
	base := now
	if active != nil && active.EndAt != nil && active.EndAt.After(now) {
		base = *active.EndAt
	}
	return base.Add(time.Duration(row.DurationDaysSnapshot) * 24 * time.Hour), nil
}

func textValue(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
