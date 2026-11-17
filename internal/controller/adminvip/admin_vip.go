package adminvip

import "github.com/suxiangdong/indu-scope-backend/api/adminvip"

type ControllerV1 struct{}

func NewV1() adminvip.IAdminVipV1 { return &ControllerV1{} }

// Keep money-moving operations off the ordinary administrator controller.
type RefundControllerV1 struct{}

func NewRefundV1() adminvip.IAdminVipRefundV1 { return &RefundControllerV1{} }
