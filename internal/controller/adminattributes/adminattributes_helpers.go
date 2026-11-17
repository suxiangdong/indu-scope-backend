package adminattributes

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"

	"github.com/suxiangdong/indu-scope-backend/api/adminattributes/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	attributeModel "github.com/suxiangdong/indu-scope-backend/internal/model/attribute"
)

func toAPIAttribute(item attributeModel.AttributeItem) v1.AttributeInfo {
	groupIDs := make([]int64, 0, len(item.GroupIDs))
	for _, groupID := range item.GroupIDs {
		groupIDs = append(groupIDs, int64(groupID))
	}
	return v1.AttributeInfo{
		ID: int64(item.ID), Code: item.Code, Name: item.Name,
		ValueType: item.ValueType, InputType: item.InputType, Unit: item.Unit,
		Placeholder: item.Placeholder, ValidationRules: item.ValidationRules,
		HelpText: item.HelpText, Status: item.Status, OptionCount: item.OptionCount,
		GroupCount: item.GroupCount,
		GroupIDs:   groupIDs,
		CreatedAt:  item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}

func toAPIOption(item attributeModel.AttributeOptionItem) v1.AttributeOptionInfo {
	return v1.AttributeOptionInfo{
		ID: int64(item.ID), AttributeID: int64(item.AttributeID), Code: item.Code,
		Name: item.Name, SortOrder: item.SortOrder, Status: item.Status,
		CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}

func toAPIBinding(item attributeModel.CategoryAttributeItem) v1.CategoryAttributeInfo {
	var groupID *int64
	if item.GroupID != nil {
		value := int64(*item.GroupID)
		groupID = &value
	}
	return v1.CategoryAttributeInfo{
		CategoryID: int64(item.CategoryID), CategoryName: item.CategoryName,
		AttributeID: int64(item.AttributeID), AttributeCode: item.AttributeCode,
		AttributeName: item.AttributeName, IsRequired: item.IsRequired,
		GroupID: groupID, GroupCode: item.GroupCode, GroupName: item.GroupName,
		GroupDisplayName: item.GroupDisplayName,
		IsFilterable:     item.IsFilterable, IsSearchable: item.IsSearchable,
		SortOrder: item.SortOrder,
		CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}

func toAPIGroup(item attributeModel.AttributeGroupItem) v1.AttributeGroupInfo {
	return v1.AttributeGroupInfo{
		ID: int64(item.ID), Code: item.Code,
		Name: item.Name, DisplayName: item.DisplayName, SortOrder: item.SortOrder, Status: item.Status,
		AttributeCount: item.AttributeCount, CategoryCount: item.CategoryCount,
		CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}

func toAPIGroupAttribute(item attributeModel.AttributeGroupAttributeItem) v1.GroupAttributeInfo {
	return v1.GroupAttributeInfo{
		GroupID: int64(item.GroupID), AttributeID: int64(item.AttributeID), Code: item.Code,
		Name: item.Name, ValueType: item.ValueType, InputType: item.InputType,
		SortOrder: item.SortOrder, Status: item.Status, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}

func toAPICategoryGroup(item attributeModel.CategoryGroupBindingItem) v1.CategoryGroupInfo {
	return v1.CategoryGroupInfo{
		CategoryID: int64(item.CategoryID), CategoryName: item.CategoryName,
		GroupID: int64(item.GroupID), GroupCode: item.GroupCode, GroupName: item.GroupName,
		GroupDisplayName: item.GroupDisplayName,
		SortOrder:        item.SortOrder, Status: item.Status, AttributeCount: item.AttributeCount,
		CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}

func toAPIError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, consts.ErrAttributeAlreadyExists),
		errors.Is(err, consts.ErrAttributeNotFound),
		errors.Is(err, consts.ErrAttributeHasCategories),
		errors.Is(err, consts.ErrAttributeHasProducts),
		errors.Is(err, consts.ErrAttributeOptionExists),
		errors.Is(err, consts.ErrAttributeOptionNotFound),
		errors.Is(err, consts.ErrAttributeOptionInUse),
		errors.Is(err, consts.ErrCategoryAttributeExists),
		errors.Is(err, consts.ErrCategoryAttributeNotFound),
		errors.Is(err, consts.ErrAttributeGroupAttributeNotFound),
		errors.Is(err, consts.ErrCategoryGroupBindingExists),
		errors.Is(err, consts.ErrCategoryGroupBindingNotFound),
		errors.Is(err, consts.ErrCategoryNotFound),
		errors.Is(err, consts.ErrCategoryHasChildren):
		return gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
	case errors.Is(err, consts.ErrAttributeGroupExists),
		errors.Is(err, consts.ErrAttributeGroupNotFound),
		errors.Is(err, consts.ErrAttributeGroupInUse):
		return gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
	}
	var parameterError *consts.InvalidParameterError
	if errors.As(err, &parameterError) {
		return gerror.NewCode(gcode.CodeInvalidParameter, parameterError.Error())
	}
	glog.Errorf(ctx, "admin attributes internal error: %+v", err)
	return gerror.NewCode(gcode.CodeInternalError, "internal server error")
}

func statusValue(status *uint) uint {
	if status == nil {
		return 1
	}
	return *status
}

func invalidRequest(message string) error {
	return gerror.NewCode(gcode.CodeInvalidParameter, message)
}
