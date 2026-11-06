package admincustomerusers

import "github.com/suxiangdong/indu-scope-backend/api/admincustomerusers"

type ControllerV1 struct{}

func NewV1() admincustomerusers.IAdminCustomerUsersV1 { return &ControllerV1{} }
