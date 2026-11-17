package admincustomerusers

import "github.com/suxiangdong/indu-scope-backend/api/admincustomerusers"

type ControllerV1 struct{}

func NewV1() admincustomerusers.IAdminCustomerUsersV1 { return &ControllerV1{} }

// RoleControllerV1 is bound separately so the role-changing endpoint can be
// protected by the super-admin middleware without widening permissions for the
// regular customer-user endpoints.
type RoleControllerV1 struct{}

func NewRoleV1() *RoleControllerV1 { return &RoleControllerV1{} }
