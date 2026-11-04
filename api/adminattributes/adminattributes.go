package adminattributes

import (
	"context"

	v1 "github.com/suxiangdong/indu-scope-backend/api/adminattributes/v1"
)

type IAdminattributesV1 interface {
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error)
	Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error)
	Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error)
	ListOptions(ctx context.Context, req *v1.ListOptionsReq) (res *v1.ListOptionsRes, err error)
	CreateOption(ctx context.Context, req *v1.CreateOptionReq) (res *v1.CreateOptionRes, err error)
	UpdateOption(ctx context.Context, req *v1.UpdateOptionReq) (res *v1.UpdateOptionRes, err error)
	DeleteOption(ctx context.Context, req *v1.DeleteOptionReq) (res *v1.DeleteOptionRes, err error)
	ListCategoryBindings(ctx context.Context, req *v1.ListCategoryBindingsReq) (res *v1.ListCategoryBindingsRes, err error)
	SaveCategoryBinding(ctx context.Context, req *v1.SaveCategoryBindingReq) (res *v1.SaveCategoryBindingRes, err error)
	DeleteCategoryBinding(ctx context.Context, req *v1.DeleteCategoryBindingReq) (res *v1.DeleteCategoryBindingRes, err error)
	ListGroups(ctx context.Context, req *v1.ListGroupsReq) (res *v1.ListGroupsRes, err error)
	CreateGroup(ctx context.Context, req *v1.CreateGroupReq) (res *v1.CreateGroupRes, err error)
	UpdateGroup(ctx context.Context, req *v1.UpdateGroupReq) (res *v1.UpdateGroupRes, err error)
	DeleteGroup(ctx context.Context, req *v1.DeleteGroupReq) (res *v1.DeleteGroupRes, err error)
	ListGroupAttributes(ctx context.Context, req *v1.ListGroupAttributesReq) (res *v1.ListGroupAttributesRes, err error)
	SaveGroupAttribute(ctx context.Context, req *v1.SaveGroupAttributeReq) (res *v1.SaveGroupAttributeRes, err error)
	DeleteGroupAttribute(ctx context.Context, req *v1.DeleteGroupAttributeReq) (res *v1.DeleteGroupAttributeRes, err error)
	ListCategoryGroups(ctx context.Context, req *v1.ListCategoryGroupsReq) (res *v1.ListCategoryGroupsRes, err error)
	BindCategoryGroup(ctx context.Context, req *v1.BindCategoryGroupReq) (res *v1.BindCategoryGroupRes, err error)
	UnbindCategoryGroup(ctx context.Context, req *v1.UnbindCategoryGroupReq) (res *v1.UnbindCategoryGroupRes, err error)
}
