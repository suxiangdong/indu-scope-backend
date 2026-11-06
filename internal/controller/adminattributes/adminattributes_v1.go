package adminattributes

import (
	"context"

	"github.com/suxiangdong/indu-scope-backend/api/adminattributes/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	"github.com/suxiangdong/indu-scope-backend/internal/model/attribute"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (*v1.ListRes, error) {
	result, err := service.AdminAttributes().List(ctx)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	items := make([]v1.AttributeInfo, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toAPIAttribute(item))
	}
	return &v1.ListRes{Items: items}, nil
}

func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (*v1.CreateRes, error) {
	item, err := service.AdminAttributes().Create(ctx, attribute.AttributeInput{
		Name: req.Name, ValueType: req.ValueType, InputType: req.InputType,
		Unit: req.Unit, Placeholder: req.Placeholder, ValidationRules: req.ValidationRules,
		HelpText: req.HelpText, Status: statusValue(req.Status), GroupIDs: req.GroupIDs,
	})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.CreateRes{AttributeInfo: toAPIAttribute(item)}, nil
}

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (*v1.UpdateRes, error) {
	item, err := service.AdminAttributes().Update(ctx, req.ID, attribute.AttributeInput{
		Name: req.Name, ValueType: req.ValueType, InputType: req.InputType,
		Unit: req.Unit, Placeholder: req.Placeholder, ValidationRules: req.ValidationRules,
		HelpText: req.HelpText, Status: statusValue(req.Status), GroupIDs: req.GroupIDs,
	})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.UpdateRes{AttributeInfo: toAPIAttribute(item)}, nil
}

func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (*v1.DeleteRes, error) {
	if err := service.AdminAttributes().Delete(ctx, req.ID); err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.DeleteRes{Success: true}, nil
}

func (c *ControllerV1) ListOptions(ctx context.Context, req *v1.ListOptionsReq) (*v1.ListOptionsRes, error) {
	result, err := service.AdminAttributes().ListOptions(ctx, req.AttributeID)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	items := make([]v1.AttributeOptionInfo, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toAPIOption(item))
	}
	return &v1.ListOptionsRes{Items: items}, nil
}

func (c *ControllerV1) CreateOption(ctx context.Context, req *v1.CreateOptionReq) (*v1.CreateOptionRes, error) {
	item, err := service.AdminAttributes().CreateOption(ctx, req.AttributeID, attribute.AttributeOptionInput{
		Name: req.Name, SortOrder: req.SortOrder, Status: statusValue(req.Status),
	})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.CreateOptionRes{AttributeOptionInfo: toAPIOption(item)}, nil
}

func (c *ControllerV1) UpdateOption(ctx context.Context, req *v1.UpdateOptionReq) (*v1.UpdateOptionRes, error) {
	item, err := service.AdminAttributes().UpdateOption(ctx, req.ID, attribute.AttributeOptionInput{
		Name: req.Name, SortOrder: req.SortOrder, Status: statusValue(req.Status),
	})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.UpdateOptionRes{AttributeOptionInfo: toAPIOption(item)}, nil
}

func (c *ControllerV1) DeleteOption(ctx context.Context, req *v1.DeleteOptionReq) (*v1.DeleteOptionRes, error) {
	if err := service.AdminAttributes().DeleteOption(ctx, req.ID); err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.DeleteOptionRes{Success: true}, nil
}

func (c *ControllerV1) ListCategoryBindings(ctx context.Context, req *v1.ListCategoryBindingsReq) (*v1.ListCategoryBindingsRes, error) {
	result, err := service.AdminAttributes().ListCategoryBindings(ctx, req.CategoryID, req.AttributeID)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	items := make([]v1.CategoryAttributeInfo, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toAPIBinding(item))
	}
	return &v1.ListCategoryBindingsRes{Items: items}, nil
}

func (c *ControllerV1) SaveCategoryBinding(ctx context.Context, req *v1.SaveCategoryBindingReq) (*v1.SaveCategoryBindingRes, error) {
	item, err := service.AdminAttributes().SaveCategoryBinding(ctx, attribute.CategoryAttributeInput{
		CategoryID: req.CategoryID, AttributeID: req.AttributeID,
		GroupID: req.GroupID, GroupName: req.GroupName,
		IsRequired: req.IsRequired, IsFilterable: req.IsFilterable,
		IsSearchable: req.IsSearchable, SortOrder: req.SortOrder,
	})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.SaveCategoryBindingRes{CategoryAttributeInfo: toAPIBinding(item)}, nil
}

func (c *ControllerV1) DeleteCategoryBinding(ctx context.Context, req *v1.DeleteCategoryBindingReq) (*v1.DeleteCategoryBindingRes, error) {
	if err := service.AdminAttributes().DeleteCategoryBinding(ctx, req.CategoryID, req.AttributeID); err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.DeleteCategoryBindingRes{Success: true}, nil
}

func (c *ControllerV1) ListGroups(ctx context.Context, req *v1.ListGroupsReq) (*v1.ListGroupsRes, error) {
	result, err := service.AdminAttributes().ListGroups(ctx)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	items := make([]v1.AttributeGroupInfo, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toAPIGroup(item))
	}
	return &v1.ListGroupsRes{Items: items}, nil
}

func (c *ControllerV1) CreateGroup(ctx context.Context, req *v1.CreateGroupReq) (*v1.CreateGroupRes, error) {
	item, err := service.AdminAttributes().CreateGroup(ctx, attribute.AttributeGroupInput{
		Name: req.Name, DisplayName: req.DisplayName,
		SortOrder: req.SortOrder, Status: statusValue(req.Status),
	})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.CreateGroupRes{AttributeGroupInfo: toAPIGroup(item)}, nil
}

func (c *ControllerV1) UpdateGroup(ctx context.Context, req *v1.UpdateGroupReq) (*v1.UpdateGroupRes, error) {
	item, err := service.AdminAttributes().UpdateGroup(ctx, req.ID, attribute.AttributeGroupInput{
		Name: req.Name, DisplayName: req.DisplayName,
		SortOrder: req.SortOrder, Status: statusValue(req.Status),
	})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.UpdateGroupRes{AttributeGroupInfo: toAPIGroup(item)}, nil
}

func (c *ControllerV1) DeleteGroup(ctx context.Context, req *v1.DeleteGroupReq) (*v1.DeleteGroupRes, error) {
	if err := service.AdminAttributes().DeleteGroup(ctx, req.ID); err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.DeleteGroupRes{Success: true}, nil
}

func (c *ControllerV1) ListGroupAttributes(ctx context.Context, req *v1.ListGroupAttributesReq) (*v1.ListGroupAttributesRes, error) {
	result, err := service.AdminAttributes().ListGroupAttributes(ctx, req.GroupID)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	items := make([]v1.GroupAttributeInfo, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toAPIGroupAttribute(item))
	}
	return &v1.ListGroupAttributesRes{Items: items}, nil
}

func (c *ControllerV1) SaveGroupAttribute(ctx context.Context, req *v1.SaveGroupAttributeReq) (*v1.SaveGroupAttributeRes, error) {
	item, err := service.AdminAttributes().SaveGroupAttribute(ctx, attribute.AttributeGroupAttributeInput{
		GroupID: req.GroupID, AttributeID: req.AttributeID, SortOrder: req.SortOrder,
	})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.SaveGroupAttributeRes{GroupAttributeInfo: toAPIGroupAttribute(item)}, nil
}

func (c *ControllerV1) DeleteGroupAttribute(ctx context.Context, req *v1.DeleteGroupAttributeReq) (*v1.DeleteGroupAttributeRes, error) {
	if err := service.AdminAttributes().DeleteGroupAttribute(ctx, req.GroupID, req.AttributeID); err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.DeleteGroupAttributeRes{Success: true}, nil
}

func (c *ControllerV1) ListCategoryGroups(ctx context.Context, req *v1.ListCategoryGroupsReq) (*v1.ListCategoryGroupsRes, error) {
	if req.CategoryID == 0 && req.GroupID == 0 {
		return nil, toAPIError(ctx, &consts.InvalidParameterError{Message: "category_id or group_id is required"})
	}
	var (
		result attribute.CategoryGroupBindingListResult
		err    error
	)
	if req.GroupID != 0 {
		result, err = service.AdminAttributes().ListGroupCategories(ctx, req.GroupID)
	} else {
		result, err = service.AdminAttributes().ListCategoryGroups(ctx, req.CategoryID)
	}
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	items := make([]v1.CategoryGroupInfo, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toAPICategoryGroup(item))
	}
	return &v1.ListCategoryGroupsRes{Items: items}, nil
}

func (c *ControllerV1) BindCategoryGroup(ctx context.Context, req *v1.BindCategoryGroupReq) (*v1.BindCategoryGroupRes, error) {
	item, err := service.AdminAttributes().BindCategoryGroup(ctx, attribute.CategoryGroupBindingInput{
		CategoryID: req.CategoryID, GroupID: req.GroupID, SortOrder: req.SortOrder, Status: statusValue(req.Status),
	})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.BindCategoryGroupRes{CategoryGroupInfo: toAPICategoryGroup(item)}, nil
}

func (c *ControllerV1) UnbindCategoryGroup(ctx context.Context, req *v1.UnbindCategoryGroupReq) (*v1.UnbindCategoryGroupRes, error) {
	if err := service.AdminAttributes().UnbindCategoryGroup(ctx, req.CategoryID, req.GroupID); err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.UnbindCategoryGroupRes{Success: true}, nil
}
