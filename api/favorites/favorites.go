package favorites

import (
	"context"

	v1 "github.com/suxiangdong/indu-scope-backend/api/favorites/v1"
)

type IFavoritesV1 interface {
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	Toggle(ctx context.Context, req *v1.ToggleReq) (res *v1.ToggleRes, err error)
	Status(ctx context.Context, req *v1.StatusReq) (res *v1.StatusRes, err error)
}
