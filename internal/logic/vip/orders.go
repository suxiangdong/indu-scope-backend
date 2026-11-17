package vip

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	vipModel "github.com/suxiangdong/indu-scope-backend/internal/model/vip"
)

func (s *sVip) CreateOrder(ctx context.Context, userID, planID uint64, autoRenew bool, idempotencyKey string) (vipModel.Order, bool, error) {
	if userID == 0 {
		return vipModel.Order{}, false, consts.ErrTokenInvalid
	}
	if planID == 0 {
		return vipModel.Order{}, false, &consts.InvalidParameterError{Message: "plan_id is required"}
	}
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if len(idempotencyKey) > 128 {
		return vipModel.Order{}, false, &consts.InvalidParameterError{Message: "idempotency_key is too long"}
	}
	var result vipModel.Order
	var reused bool
	err := s.paymentDB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// Serialize all checkout creation/settlement for this user. The key is
		// read only after this lock, so concurrent retries see the committed
		// order before mutable plans, agreements or free eligibility are read.
		var user struct {
			ID uint64 `orm:"id"`
		}
		if err := tx.Model("users").Fields("id").Where("id", userID).LockUpdate().Scan(&user); err != nil {
			return err
		}
		if user.ID == 0 {
			return consts.ErrTokenInvalid
		}
		if idempotencyKey != "" {
			var existing *orderRow
			err := tx.Model("payment_orders").Where("user_id", userID).Where("idempotency_key", idempotencyKey).Limit(1).Scan(&existing)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return err
			}
			if existing != nil {
				if existing.PlanID != planID || (existing.PlanTypeSnapshot == "auto_renew") != autoRenew {
					return &consts.InvalidParameterError{Message: "idempotency_key belongs to a different checkout"}
				}
				result, reused = existing.order(), true
				return nil
			}
		}
		var plan *planRow
		if err := tx.Model("vip_plans").Where("id", planID).Where("status", 1).Scan(&plan); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if plan == nil {
			return errors.New("vip plan is unavailable")
		}
		if plan.DurationDays == 0 || plan.DurationDays > maxMembershipDurationDays {
			return errors.New("套餐有效天数必须在1至36500之间")
		}
		planType := plan.PlanType
		if planType == "" {
			planType = "normal"
		}
		if planType != "normal" && planType != "auto_renew" {
			return errors.New("vip plan type is invalid")
		}
		if (planType == "auto_renew") != autoRenew {
			return errors.New("套餐类型与自动续费选项不匹配")
		}
		if planType == "auto_renew" {
			return errors.New("连续包月签约尚未开通，请选择普通套餐")
		}
		serviceAgreement, err := latestAgreement(tx.Model("vip_agreements"), vipModel.AgreementKeyService, true)
		if err != nil {
			return err
		}
		if serviceAgreement == nil {
			return errors.New("active service agreement is unavailable")
		}
		var autoRenewAgreement *vipModel.Agreement
		if planType == "auto_renew" {
			autoRenewAgreement, err = latestAgreement(tx.Model("vip_agreements"), vipModel.AgreementKeyAutoRenew, true)
			if err != nil {
				return err
			}
			if autoRenewAgreement == nil {
				return errors.New("active auto renewal agreement is unavailable")
			}
		}
		active, err := loadPaymentSubscription(ctx, tx, userID)
		if err != nil {
			return err
		}
		if active != nil {
			if err := validateNormalRenewal(active.subscription().Plan, active.EndAt, plan.plan()); err != nil {
				return err
			}
		}
		now := time.Now().Truncate(time.Millisecond)
		random := make([]byte, 12)
		if _, err := rand.Read(random); err != nil {
			return err
		}
		row := orderRow{OrderNo: "VIP" + now.Format("20060102150405") + hex.EncodeToString(random), UserID: userID, PlanID: planID,
			AmountFen: plan.PriceFen, Currency: plan.Currency, Status: "pending", PaymentMethod: "wechat_miniapp", Provider: "wechatpay",
			PlanCodeSnapshot: plan.Code, PlanNameSnapshot: plan.Name, PlanDescriptionSnapshot: plan.Description,
			OriginalPriceFenSnapshot: plan.OriginalPriceFen, PlanTypeSnapshot: planType, BillingCycleSnapshot: plan.BillingCycle,
			DurationDaysSnapshot: plan.DurationDays, AutoRenewEnabledSnapshot: plan.AutoRenewEnabled,
			ServiceAgreementID: uint64Ptr(serviceAgreement.ID), ServiceAgreementVersion: uintPtr(serviceAgreement.Version),
			AutoRenewAgreementID: agreementIDPtr(autoRenewAgreement), AutoRenewAgreementVersion: agreementVersionPtr(autoRenewAgreement), CreatedAt: now}
		data := g.Map{"order_no": row.OrderNo, "user_id": userID, "plan_id": planID, "amount_fen": row.AmountFen, "currency": row.Currency,
			"plan_code_snapshot": row.PlanCodeSnapshot, "plan_name_snapshot": row.PlanNameSnapshot, "plan_description_snapshot": nullableString(row.PlanDescriptionSnapshot),
			"original_price_fen_snapshot": row.OriginalPriceFenSnapshot, "plan_type_snapshot": row.PlanTypeSnapshot,
			"billing_cycle_snapshot": row.BillingCycleSnapshot, "duration_days_snapshot": row.DurationDaysSnapshot,
			"auto_renew_enabled_snapshot": row.AutoRenewEnabledSnapshot, "service_agreement_id": row.ServiceAgreementID,
			"service_agreement_version": row.ServiceAgreementVersion, "auto_renew_agreement_id": agreementID(autoRenewAgreement),
			"auto_renew_agreement_version": agreementVersion(autoRenewAgreement), "idempotency_key": nullableString(idempotencyKey)}
		if plan.PriceFen == 0 {
			count, err := tx.Model("payment_orders").Where("user_id", userID).Where("status", "paying").Where("expires_at > NOW()").Count()
			if err != nil {
				return err
			}
			if count > 0 {
				return errors.New("请先处理待完成的支付订单")
			}
			endAt := now.Add(time.Duration(plan.DurationDays) * 24 * time.Hour)
			id, err := tx.Model("vip_subscriptions").Data(g.Map{"user_id": userID, "plan_id": planID, "plan_type_snapshot": planType, "currency_snapshot": row.Currency,
				"status": "active", "auto_renew": 0, "provider": "manual", "start_at": now, "end_at": endAt}).InsertAndGetId()
			if err != nil {
				return err
			}
			row.SubscriptionID, row.PaidAt, row.EntitlementStartAt, row.EntitlementEndAt = uint64Ptr(uint64(id)), &now, &now, &endAt
			row.Status, row.PaymentMethod, row.Provider = "paid", "free", "manual"
			data["subscription_id"], data["paid_at"], data["entitlement_start_at"], data["entitlement_end_at"] = row.SubscriptionID, now, now, endAt
		} else {
			expiresAt := now.Add(30 * time.Minute)
			row.ExpiresAt, data["expires_at"] = &expiresAt, expiresAt
		}
		data["status"], data["payment_method"], data["provider"] = row.Status, row.PaymentMethod, row.Provider
		id, err := tx.Model("payment_orders").Data(data).InsertAndGetId()
		if err != nil {
			return err
		}
		row.ID = uint64(id)
		result = row.order()
		return nil
	})
	return result, reused, err
}
