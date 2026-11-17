package adminhotcategories

import (
	"context"

	"github.com/suxiangdong/indu-scope-backend/api/adminhotcategories/v1"
)

type IAdminHotCategoriesV1 interface {
	List(ctx context.Context, req *v1.ListReq) (*v1.ListRes, error)
	Create(ctx context.Context, req *v1.CreateReq) (*v1.CreateRes, error)
	Update(ctx context.Context, req *v1.UpdateReq) (*v1.UpdateRes, error)
	Delete(ctx context.Context, req *v1.DeleteReq) (*v1.DeleteRes, error)
	Status(ctx context.Context, req *v1.StatusReq) (*v1.StatusRes, error)
	Sort(ctx context.Context, req *v1.SortReq) (*v1.SortRes, error)
}
