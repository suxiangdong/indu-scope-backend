package service

import (
	"context"

	attributeModel "github.com/suxiangdong/indu-scope-backend/internal/model/attribute"
)

type IAdminAttributes interface {
	List(ctx context.Context) (attributeModel.AttributeListResult, error)
	Create(ctx context.Context, input attributeModel.AttributeInput) (attributeModel.AttributeItem, error)
	Update(ctx context.Context, id uint64, input attributeModel.AttributeInput) (attributeModel.AttributeItem, error)
	Delete(ctx context.Context, id uint64) error
	ListOptions(ctx context.Context, attributeID uint64) (attributeModel.AttributeOptionListResult, error)
	CreateOption(ctx context.Context, attributeID uint64, input attributeModel.AttributeOptionInput) (attributeModel.AttributeOptionItem, error)
	UpdateOption(ctx context.Context, id uint64, input attributeModel.AttributeOptionInput) (attributeModel.AttributeOptionItem, error)
	DeleteOption(ctx context.Context, id uint64) error
	ListCategoryBindings(ctx context.Context, categoryID, attributeID *uint64) (attributeModel.CategoryAttributeListResult, error)
	SaveCategoryBinding(ctx context.Context, input attributeModel.CategoryAttributeInput) (attributeModel.CategoryAttributeItem, error)
	DeleteCategoryBinding(ctx context.Context, categoryID, attributeID uint64) error
	ListGroups(ctx context.Context) (attributeModel.AttributeGroupListResult, error)
	CreateGroup(ctx context.Context, input attributeModel.AttributeGroupInput) (attributeModel.AttributeGroupItem, error)
	UpdateGroup(ctx context.Context, id uint64, input attributeModel.AttributeGroupInput) (attributeModel.AttributeGroupItem, error)
	DeleteGroup(ctx context.Context, id uint64) error
	ListGroupAttributes(ctx context.Context, groupID uint64) (attributeModel.AttributeGroupAttributeListResult, error)
	SaveGroupAttribute(ctx context.Context, input attributeModel.AttributeGroupAttributeInput) (attributeModel.AttributeGroupAttributeItem, error)
	DeleteGroupAttribute(ctx context.Context, groupID, attributeID uint64) error
	ListCategoryGroups(ctx context.Context, categoryID uint64) (attributeModel.CategoryGroupBindingListResult, error)
	BindCategoryGroup(ctx context.Context, input attributeModel.CategoryGroupBindingInput) (attributeModel.CategoryGroupBindingItem, error)
	UnbindCategoryGroup(ctx context.Context, categoryID, groupID uint64) error
}

var localAdminAttributes IAdminAttributes

func AdminAttributes() IAdminAttributes {
	if localAdminAttributes == nil {
		panic("implement not found for interface IAdminAttributes, forgot register?")
	}
	return localAdminAttributes
}

func RegisterAdminAttributes(i IAdminAttributes) {
	localAdminAttributes = i
}
