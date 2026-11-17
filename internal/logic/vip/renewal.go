package vip

import (
	"errors"
	"time"

	vipModel "github.com/suxiangdong/indu-scope-backend/internal/model/vip"
)

func normalPlan(planType string) bool { return planType == "" || planType == "normal" }

// Any ordinary paid billing period may renew an ordinary membership. This
// check is only for new purchases/prepayment, never for settling paid orders.
func validateNormalRenewal(current vipModel.Plan, endAt *time.Time, target vipModel.Plan) error {
	if endAt == nil {
		return errors.New("长期有效会员无需续费")
	}
	if !normalPlan(current.PlanType) || !normalPlan(target.PlanType) {
		return errors.New("有效会员仅支持普通套餐续费")
	}
	if target.PriceFen == 0 {
		return errors.New("已有有效会员，不能使用免费套餐续费")
	}
	if target.Currency != current.Currency {
		return errors.New("请选择同币种的普通付费套餐续费")
	}
	return nil
}

func validatePaidRenewal(active *subscriptionRow, row orderRow) error {
	if active == nil {
		return nil
	}
	return validateNormalRenewal(active.subscription().Plan, active.EndAt, vipModel.Plan{
		ID: row.PlanID, PlanType: row.PlanTypeSnapshot,
		PriceFen: row.AmountFen, Currency: row.Currency,
	})
}
