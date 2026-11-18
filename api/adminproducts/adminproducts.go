package adminproducts

import (
	"context"

	v1 "github.com/suxiangdong/indu-scope-backend/api/adminproducts/v1"
)

type IAdminProductsV1 interface {
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error)
	UpdateStatus(ctx context.Context, req *v1.UpdateStatusReq) (res *v1.UpdateStatusRes, err error)
	Review(ctx context.Context, req *v1.ReviewReq) (res *v1.ReviewRes, err error)
}
