package adminprocurementrequirements

import (
	"context"

	v1 "github.com/suxiangdong/indu-scope-backend/api/adminprocurementrequirements/v1"
)

type IAdminProcurementRequirementsV1 interface {
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error)
	Review(ctx context.Context, req *v1.ReviewReq) (res *v1.ReviewRes, err error)
}
