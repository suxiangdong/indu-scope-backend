package admincustomerusers

import (
	"context"

	"github.com/suxiangdong/indu-scope-backend/api/admincustomerusers/v1"
)

type IAdminCustomerUsersV1 interface {
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	UpdateStatus(ctx context.Context, req *v1.UpdateStatusReq) (res *v1.UpdateStatusRes, err error)
}
