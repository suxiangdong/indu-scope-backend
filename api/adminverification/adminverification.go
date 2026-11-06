package adminverification

import (
	"context"

	v1 "github.com/suxiangdong/indu-scope-backend/api/adminverification/v1"
)

type IAdminverificationV1 interface {
	List(ctx context.Context, req *v1.ListReq) (*v1.ListRes, error)
	Detail(ctx context.Context, req *v1.DetailReq) (*v1.DetailRes, error)
	Review(ctx context.Context, req *v1.ReviewReq) (*v1.ReviewRes, error)
}
