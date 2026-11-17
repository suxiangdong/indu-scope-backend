package adminappcontent

import "github.com/suxiangdong/indu-scope-backend/api/adminappcontent"

type ControllerV1 struct{}

func NewV1() adminappcontent.IAdminAppContentV1 { return &ControllerV1{} }
