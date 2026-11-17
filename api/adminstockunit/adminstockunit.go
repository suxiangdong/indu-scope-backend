package adminstockunit

import (
	"context"

	v1 "github.com/suxiangdong/indu-scope-backend/api/adminstockunit/v1"
)

type IAdminStockUnitV1 interface {
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error)
	Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error)
	Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error)
	ListCategories(ctx context.Context, req *v1.ListCategoriesReq) (res *v1.ListCategoriesRes, err error)
	BindCategory(ctx context.Context, req *v1.BindCategoryReq) (res *v1.BindCategoryRes, err error)
	UnbindCategory(ctx context.Context, req *v1.UnbindCategoryReq) (res *v1.UnbindCategoryRes, err error)
}
