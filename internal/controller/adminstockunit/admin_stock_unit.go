package adminstockunit

import "github.com/suxiangdong/indu-scope-backend/api/adminstockunit"

type ControllerV1 struct{}

func NewV1() adminstockunit.IAdminStockUnitV1 { return &ControllerV1{} }
