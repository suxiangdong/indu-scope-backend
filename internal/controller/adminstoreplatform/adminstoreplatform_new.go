package adminstoreplatform

import "github.com/suxiangdong/indu-scope-backend/api/adminstoreplatform"

type ControllerV1 struct{}

func NewV1() adminstoreplatform.IAdminStorePlatformV1 { return &ControllerV1{} }
