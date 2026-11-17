package products

import (
	"context"

	v1 "github.com/suxiangdong/indu-scope-backend/api/products/v1"
)

// IProductsV1 defines the buyer-facing product catalog API.
type IProductsV1 interface {
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error)
}
