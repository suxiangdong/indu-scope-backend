package vip

import "github.com/suxiangdong/indu-scope-backend/api/vip"

type ControllerV1 struct{}

func NewV1() vip.IVipV1 { return &ControllerV1{} }
