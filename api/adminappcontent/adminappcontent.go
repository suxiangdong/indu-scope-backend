package adminappcontent

import (
	"context"
	"github.com/suxiangdong/indu-scope-backend/api/adminappcontent/v1"
)

type IAdminAppContentV1 interface {
	List(ctx context.Context, req *v1.ListReq) (*v1.ListRes, error)
	Update(ctx context.Context, req *v1.UpdateReq) (*v1.UpdateRes, error)
}
