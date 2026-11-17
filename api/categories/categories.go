package categories

import (
	"context"

	"github.com/suxiangdong/indu-scope-backend/api/categories/v1"
)

type ICategoriesV1 interface {
	Hot(ctx context.Context, req *v1.HotReq) (*v1.HotRes, error)
}
