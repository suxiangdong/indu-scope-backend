package vip

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	bizctxModel "github.com/suxiangdong/indu-scope-backend/internal/model/bizctx"
	vipModel "github.com/suxiangdong/indu-scope-backend/internal/model/vip"
	"github.com/suxiangdong/indu-scope-backend/internal/wechatpay"
)

type refundGateway interface {
	CreateRefund(context.Context, wechatpay.RefundInput) (*wechatpay.RefundResult, error)
	QueryRefund(context.Context, string) (*wechatpay.RefundResult, error)
	RefundCallbackURL() (string, error)
}

func (s *sVip) loadRefundClient(ctx context.Context) (refundGateway, wechatpay.Config, error) {
	if s.refundGateway != nil {
		return s.refundGateway, s.refundConfig, nil
	}
	client, err := wechatpay.Load(ctx)
	if err != nil {
		return nil, wechatpay.Config{}, errors.New("微信退款配置未就绪，请检查支付配置")
	}
	return client, client.Config, nil
}

type refundRow struct {
	ID                 uint64     `orm:"id"`
	RefundNo           string     `orm:"refund_no"`
	OrderID            uint64     `orm:"order_id"`
	OrderNo            string     `orm:"order_no"`
	UserID             uint64     `orm:"user_id"`
	UserCompany        string     `orm:"user_company"`
	PlanName           string     `orm:"plan_name"`
	RequestedByAdminID *uint64    `orm:"requested_by_admin_id"`
	AdminName          string     `orm:"admin_name"`
	Reason             string     `orm:"reason"`
	AmountFen          uint       `orm:"amount_fen"`
	TotalFen           uint       `orm:"total_fen"`
	Currency           string     `orm:"currency"`
	Status             string     `orm:"status"`
	ProviderRefundID   string     `orm:"provider_refund_id"`
	PayerRefundFen     *uint      `orm:"payer_refund_fen"`
	ReceivedAccount    string     `orm:"received_account"`
	EntitlementStatus  string     `orm:"entitlement_status"`
	ReclaimedSeconds   uint64     `orm:"reclaimed_seconds"`
	LastError          string     `orm:"last_error"`
	SuccessAt          *time.Time `orm:"success_at"`
	CreatedAt          time.Time  `orm:"created_at"`
	UpdatedAt          time.Time  `orm:"updated_at"`
}

func (r refundRow) refund() vipModel.Refund {
	return vipModel.Refund{ID: r.ID, RefundNo: r.RefundNo, OrderID: r.OrderID, OrderNo: r.OrderNo, UserID: r.UserID,
		UserCompany: r.UserCompany, PlanName: r.PlanName, RequestedByAdminID: r.RequestedByAdminID, AdminName: r.AdminName,
		Reason: r.Reason, AmountFen: r.AmountFen, TotalFen: r.TotalFen, Currency: r.Currency, Status: r.Status,
		ProviderRefundID: r.ProviderRefundID, PayerRefundFen: r.PayerRefundFen, ReceivedAccount: r.ReceivedAccount,
		EntitlementStatus: r.EntitlementStatus, ReclaimedSeconds: r.ReclaimedSeconds, LastError: r.LastError,
		SuccessAt: r.SuccessAt, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

func (s *sVip) loadRefund(ctx context.Context, id uint64) (*refundRow, error) {
	var row *refundRow
	err := s.paymentDB().Model("payment_refunds").Ctx(ctx).Where("id", id).Scan(&row)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && row == nil) {
		return nil, errors.New("退款记录不存在")
	}
	return row, err
}

func (s *sVip) loadAdminOrder(ctx context.Context, orderID uint64) (*orderRow, error) {
	var row *orderRow
	err := s.paymentDB().Model("payment_orders").Ctx(ctx).Where("id", orderID).Scan(&row)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && row == nil) {
		return nil, errors.New("会员订单不存在")
	}
	return row, err
}

func validateRefundableOrder(row orderRow, cfg wechatpay.Config, now time.Time) error {
	if row.Status != "paid" {
		return errors.New("仅已支付订单可申请退款")
	}
	if row.AmountFen == 0 || row.Provider != "wechatpay" || row.PaymentMethod != "wechat_miniapp" || row.ProviderTransactionID == "" {
		return errors.New("免费订单或非微信支付订单不支持退款")
	}
	if row.PlanTypeSnapshot != "normal" || row.Currency != "CNY" || row.PaymentAppID != cfg.AppID || row.PaymentMchID != cfg.MchID {
		return errors.New("退款订单类型或商户身份不匹配")
	}
	if row.PaidAt == nil || now.Sub(*row.PaidAt) > 365*24*time.Hour {
		return errors.New("订单已超过微信支付可退款期限")
	}
	return nil
}

func (s *sVip) CreateRefund(ctx context.Context, orderID, adminID uint64, reason string) (*vipModel.Refund, error) {
	operatorID, err := refundOperator(ctx)
	if err != nil {
		return nil, err
	}
	if operatorID != adminID {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "退款操作人与登录身份不匹配")
	}
	reason = strings.TrimSpace(reason)
	if orderID == 0 || adminID == 0 || reason == "" || len(reason) > 80 {
		return nil, errors.New("请填写退款订单及原因，原因不能超过80字节")
	}
	client, cfg, err := s.loadRefundClient(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := client.RefundCallbackURL(); err != nil {
		return nil, err
	}
	hint, err := s.loadAdminOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	var refundID uint64
	err = s.paymentDB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var user struct {
			ID uint64 `orm:"id"`
		}
		if err := tx.Model("users").Fields("id").Where("id", hint.UserID).LockUpdate().Scan(&user); err != nil {
			return err
		}
		var order orderRow
		if err := tx.Model("payment_orders").Where("id", orderID).LockUpdate().Scan(&order); err != nil {
			return err
		}
		// All create/retry/notification paths use user -> order -> refund locks.
		// A pending/accepted/abnormal/success refund reserves the entire order.
		var existing *refundRow
		if err := tx.Model("payment_refunds").Where("order_id", orderID).Where("status <> 'closed'").OrderDesc("id").Limit(1).Scan(&existing); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if existing != nil {
			refundID = existing.ID
			return nil
		}
		if err := validateRefundableOrder(order, cfg, time.Now()); err != nil {
			return err
		}
		random := make([]byte, 12)
		if _, err := rand.Read(random); err != nil {
			return err
		}
		refundNo := "RF" + time.Now().Format("20060102150405") + hex.EncodeToString(random)
		id, err := tx.Model("payment_refunds").Data(g.Map{"refund_no": refundNo, "order_id": order.ID, "order_no": order.OrderNo, "user_id": order.UserID,
			"requested_by_admin_id": adminID, "reason": reason, "amount_fen": order.AmountFen, "total_fen": order.AmountFen, "currency": order.Currency, "status": "created", "provider": "wechatpay", "entitlement_status": "pending"}).InsertAndGetId()
		refundID = uint64(id)
		return err
	})
	if err != nil {
		return nil, err
	}
	return s.RetryRefund(ctx, refundID)
}

func (s *sVip) RetryRefund(ctx context.Context, id uint64) (*vipModel.Refund, error) {
	if _, err := refundOperator(ctx); err != nil {
		return nil, err
	}
	row, err := s.loadRefund(ctx, id)
	if err != nil {
		return nil, err
	}
	if row.Status != "created" {
		item := row.refund()
		return &item, nil
	}
	client, cfg, err := s.loadRefundClient(ctx)
	if err != nil {
		return nil, err
	}
	order, err := s.loadAdminOrder(ctx, row.OrderID)
	if err != nil {
		return nil, err
	}
	if err := validateRefundableOrder(*order, cfg, time.Now()); err != nil {
		return nil, err
	}
	result, err := client.CreateRefund(ctx, wechatpay.RefundInput{OrderNo: row.OrderNo, TransactionID: order.ProviderTransactionID, RefundNo: row.RefundNo,
		Reason: row.Reason, Currency: row.Currency, TotalFen: int64(row.TotalFen), RefundFen: int64(row.AmountFen)})
	if err != nil {
		// The request might have reached WeChat. Never generate another refund
		// number on transport errors or interpret acceptance as success.
		_, dbErr := s.paymentDB().Model("payment_refunds").Ctx(ctx).Where("id", id).Where("status", "created").Data(g.Map{"last_error": "微信退款结果尚未确认，请刷新状态或重试原退款单"}).Update()
		if dbErr != nil {
			return nil, dbErr
		}
		latest, loadErr := s.loadRefund(ctx, id)
		if loadErr != nil {
			return nil, loadErr
		}
		item := latest.refund()
		return &item, nil
	}
	if err := s.applyRefundResult(ctx, "refund-create:"+row.RefundNo+":"+result.Status, "REFUND.CREATE."+result.Status, nil, result); err != nil {
		return nil, err
	}
	latest, err := s.loadRefund(ctx, id)
	if err != nil {
		return nil, err
	}
	item := latest.refund()
	return &item, nil
}

// Check the server-populated identity again at the money-moving service
// boundary. Client-provided roles or administrator IDs cannot grant refunds.
func refundOperator(ctx context.Context) (uint64, error) {
	var value *bizctxModel.Context
	if ctx != nil {
		value, _ = ctx.Value(consts.BizctxContextKey).(*bizctxModel.Context)
	}
	if value == nil || value.AdminUser == nil || value.AdminUser.ID <= 0 || value.AdminUser.Status == 0 || !value.AdminUser.IsSuperAdmin {
		return 0, gerror.NewCode(gcode.CodeNotAuthorized, "仅超级管理员允许操作退款")
	}
	return uint64(value.AdminUser.ID), nil
}

func (s *sVip) SyncRefund(ctx context.Context, id uint64) (*vipModel.Refund, error) {
	row, err := s.loadRefund(ctx, id)
	if err != nil {
		return nil, err
	}
	if row.Status == "success" || row.Status == "closed" {
		item := row.refund()
		return &item, nil
	}
	client, cfg, err := s.loadRefundClient(ctx)
	if err != nil {
		return nil, err
	}
	order, err := s.loadAdminOrder(ctx, row.OrderID)
	if err != nil {
		return nil, err
	}
	if order.PaymentMchID != cfg.MchID || order.PaymentAppID != cfg.AppID {
		return nil, errors.New("订单商户配置不匹配")
	}
	result, err := client.QueryRefund(ctx, row.RefundNo)
	if err != nil {
		return nil, errors.New("微信退款状态暂未确认；可稍后刷新，未受理的退款请重试原退款单")
	}
	if err := s.applyRefundResult(ctx, "refund-query:"+row.RefundNo+":"+result.Status, "REFUND.QUERY."+result.Status, nil, result); err != nil {
		return nil, err
	}
	latest, err := s.loadRefund(ctx, id)
	if err != nil {
		return nil, err
	}
	item := latest.refund()
	return &item, nil
}

func (s *sVip) ListRefunds(ctx context.Context, f vipModel.RefundListFilter) (vipModel.RefundListResult, error) {
	if f.Page == 0 {
		f.Page = 1
	}
	if f.PageSize == 0 || f.PageSize > 100 {
		f.PageSize = 20
	}
	q := s.paymentDB().Model("payment_refunds r").Ctx(ctx).LeftJoin("payment_orders o", "o.id=r.order_id").LeftJoin("users u", "u.id=r.user_id").LeftJoin("admin_users a", "a.id=r.requested_by_admin_id")
	if f.OrderID > 0 {
		q = q.Where("r.order_id", f.OrderID)
	}
	if f.OrderNo != "" {
		q = q.Where("r.order_no", strings.TrimSpace(f.OrderNo))
	}
	if f.RefundNo != "" {
		q = q.Where("r.refund_no", strings.TrimSpace(f.RefundNo))
	}
	if f.Status != "" {
		q = q.Where("r.status", f.Status)
	}
	total, err := q.Count()
	if err != nil {
		return vipModel.RefundListResult{}, err
	}
	var rows []refundRow
	err = q.Fields("r.*, o.plan_name_snapshot AS plan_name, u.company_name AS user_company, a.name AS admin_name").OrderDesc("r.id").Limit(int((f.Page-1)*f.PageSize), int(f.PageSize)).Scan(&rows)
	if err != nil {
		return vipModel.RefundListResult{}, err
	}
	items := make([]vipModel.Refund, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.refund())
	}
	return vipModel.RefundListResult{Items: items, Total: uint64(total), Page: f.Page, PageSize: f.PageSize}, nil
}

func (s *sVip) HandleWechatRefundNotification(ctx context.Context, req *http.Request) error {
	client, err := wechatpay.Load(ctx)
	if err != nil {
		return err
	}
	n, err := client.ParseRefundNotification(ctx, req)
	if err != nil {
		return err
	}
	return s.applyRefundResult(ctx, n.ID, n.EventType, n.Payload, n.Result)
}

func validateRefundResult(order orderRow, row refundRow, r *wechatpay.RefundResult) error {
	if r == nil || r.RefundNo != row.RefundNo || r.RefundID == "" || r.OrderNo != order.OrderNo || row.OrderNo != order.OrderNo || row.UserID != order.UserID || row.OrderID != order.ID || r.TransactionID != order.ProviderTransactionID || r.MchID != order.PaymentMchID || order.Provider != "wechatpay" || order.PaymentMethod != "wechat_miniapp" {
		return errors.New("退款身份与订单快照不匹配")
	}
	if r.TotalFen != int64(order.AmountFen) || r.RefundFen != int64(row.AmountFen) || row.TotalFen != order.AmountFen || row.AmountFen != order.AmountFen || row.AmountFen == 0 || r.Currency != order.Currency || row.Currency != order.Currency {
		return errors.New("退款金额或币种与订单快照不匹配")
	}
	if row.ProviderRefundID != "" && row.ProviderRefundID != r.RefundID {
		return errors.New("微信退款单号已变化")
	}
	if r.PayerRefundFen != nil && (*r.PayerRefundFen < 0 || *r.PayerRefundFen > r.RefundFen) {
		return errors.New("用户实际退款金额无效")
	}
	if r.Status != "SUCCESS" && r.Status != "PROCESSING" && r.Status != "CLOSED" && r.Status != "ABNORMAL" {
		return errors.New("未知微信退款状态")
	}
	if r.Status == "SUCCESS" && (r.SuccessAt == nil || r.SuccessAt.IsZero()) {
		return errors.New("缺少退款成功时间")
	}
	if order.Status != "paid" && order.Status != "refunded" {
		return errors.New("退款订单状态不匹配")
	}
	return nil
}

func (s *sVip) applyRefundResult(ctx context.Context, eventID, eventType string, payload []byte, result *wechatpay.RefundResult) (resultErr error) {
	if result == nil || eventID == "" || result.RefundNo == "" {
		return errors.New("无效退款事件")
	}
	if len(payload) == 0 {
		var err error
		payload, err = json.Marshal(result)
		if err != nil {
			return err
		}
	}
	if _, err := s.paymentDB().Model("payment_events").Ctx(ctx).Data(g.Map{"provider": "wechatpay", "event_id": eventID, "order_no": result.OrderNo, "event_type": eventType, "payload": string(payload), "process_status": "pending"}).OnDuplicate(g.Map{"event_id": eventID}).Insert(); err != nil {
		return err
	}
	defer func() {
		if resultErr != nil {
			_, _ = s.paymentDB().Model("payment_events").Ctx(ctx).Where("provider", "wechatpay").Where("event_id", eventID).Where("process_status <> 'processed'").Data(g.Map{"process_status": "failed"}).Update()
		}
	}()
	var hint *refundRow
	if err := s.paymentDB().Model("payment_refunds").Ctx(ctx).Where("refund_no", result.RefundNo).Scan(&hint); err != nil {
		return err
	}
	if hint == nil {
		return errors.New("退款记录不存在")
	}
	return s.paymentDB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var user struct {
			ID uint64 `orm:"id"`
		}
		if err := tx.Model("users").Fields("id").Where("id", hint.UserID).LockUpdate().Scan(&user); err != nil {
			return err
		}
		var order orderRow
		if err := tx.Model("payment_orders").Where("id", hint.OrderID).LockUpdate().Scan(&order); err != nil {
			return err
		}
		var refund refundRow
		if err := tx.Model("payment_refunds").Where("id", hint.ID).LockUpdate().Scan(&refund); err != nil {
			return err
		}
		if err := validateRefundResult(order, refund, result); err != nil {
			return err
		}
		if refund.Status == "success" {
			return finishPaymentEvent(tx, eventID)
		}
		if result.Status == "PROCESSING" && refund.Status != "created" && refund.Status != "processing" {
			return finishPaymentEvent(tx, eventID)
		}
		if refund.Status == "closed" && result.Status != "SUCCESS" {
			return finishPaymentEvent(tx, eventID)
		}
		data := g.Map{"status": strings.ToLower(result.Status), "provider_refund_id": result.RefundID, "last_error": nil, "received_account": nullableString(result.ReceivedAccount)}
		if result.PayerRefundFen != nil && *result.PayerRefundFen >= 0 {
			data["payer_refund_fen"] = *result.PayerRefundFen
		}
		if result.Status == "SUCCESS" {
			data["success_at"] = *result.SuccessAt
			if order.Status != "refunded" {
				status, seconds, err := reclaimOrderEntitlement(ctx, tx, order, time.Now().Truncate(time.Millisecond))
				if err != nil {
					return err
				}
				data["entitlement_status"], data["reclaimed_seconds"] = status, seconds
				if _, err := tx.Model("payment_orders").Where("id", order.ID).Data(g.Map{"status": "refunded"}).Update(); err != nil {
					return err
				}
			} else {
				data["entitlement_status"] = "none"
			}
		}
		if _, err := tx.Model("payment_refunds").Where("id", refund.ID).Data(data).Update(); err != nil {
			return err
		}
		return finishPaymentEvent(tx, eventID)
	})
}
