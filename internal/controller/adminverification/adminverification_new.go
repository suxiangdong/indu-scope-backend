package adminverification

import "github.com/suxiangdong/indu-scope-backend/api/adminverification"

type ControllerV1 struct{}

func NewV1() adminverification.IAdminverificationV1 {
	return &ControllerV1{}
}
