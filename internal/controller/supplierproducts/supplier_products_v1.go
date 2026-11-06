package supplierproducts

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"

	"github.com/suxiangdong/indu-scope-backend/api/supplierproducts/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	supplierProductModel "github.com/suxiangdong/indu-scope-backend/internal/model/supplierproduct"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) FormOptions(ctx context.Context, req *v1.FormOptionsReq) (*v1.FormOptionsRes, error) {
	var categoryID *uint64
	if req != nil {
		categoryID = req.CategoryID
	}
	if err := requireSupplier(ctx); err != nil {
		return nil, err
	}
	result, err := service.SupplierProducts().FormOptions(ctx, categoryID)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	res := &v1.FormOptionsRes{Categories: make([]v1.CategoryInfo, 0, len(result.Categories)), Attributes: make([]v1.AttributeInfo, 0, len(result.Attributes))}
	for _, item := range result.Categories {
		res.Categories = append(res.Categories, v1.CategoryInfo{ID: item.ID, ParentID: item.ParentID, Name: item.Name, SortOrder: item.SortOrder})
	}
	for _, item := range result.Attributes {
		attribute := v1.AttributeInfo{
			ID: item.ID, Code: item.Code, Name: item.Name, ValueType: item.ValueType,
			InputType: item.InputType, Unit: item.Unit, Placeholder: item.Placeholder,
			HelpText: item.HelpText, IsRequired: item.IsRequired, SortOrder: item.SortOrder,
			GroupDisplayName: item.GroupDisplayName, Options: make([]v1.AttributeOptionInfo, 0, len(item.Options)),
		}
		for _, option := range item.Options {
			attribute.Options = append(attribute.Options, v1.AttributeOptionInfo{ID: option.ID, Code: option.Code, Name: option.Name})
		}
		res.Attributes = append(res.Attributes, attribute)
	}
	return res, nil
}

func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (*v1.CreateRes, error) {
	if err := requireSupplier(ctx); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "product is required")
	}
	valueInputs := make([]supplierProductModel.AttributeValueInput, 0, len(req.Attributes))
	for _, value := range req.Attributes {
		valueInputs = append(valueInputs, supplierProductModel.AttributeValueInput{
			AttributeID: value.AttributeID, Value: value.Value, OptionID: value.OptionID, OptionIDs: value.OptionIDs,
		})
	}
	user := service.Bizctx().Get(ctx).User
	result, err := service.SupplierProducts().Create(ctx, supplierProductModel.CreateInput{
		SupplierUserID: uint64(user.ID), CategoryID: req.CategoryID, Name: req.Name,
		DescriptionHTML: req.DescriptionHTML, ConditionCode: req.ConditionCode,
		StockQuantity: req.StockQuantity, StockUnitCode: req.StockUnitCode, Attributes: valueInputs,
	})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.CreateRes{ProductInfo: toAPI(result)}, nil
}

func requireSupplier(ctx context.Context) error {
	value := service.Bizctx().Get(ctx)
	if value == nil || value.User == nil {
		return gerror.NewCode(gcode.CodeNotAuthorized, "authorization is required")
	}
	if value.User.Role != 2 {
		return gerror.NewCode(gcode.CodeNotAuthorized, "supplier authorization is required")
	}
	return nil
}

func toAPI(item supplierProductModel.ProductItem) v1.ProductInfo {
	return v1.ProductInfo{ID: item.ID, SupplierUserID: item.SupplierUserID, CategoryID: item.CategoryID, Name: item.Name, DescriptionHTML: item.DescriptionHTML, ConditionCode: item.ConditionCode, StockQuantity: item.StockQuantity, StockUnitCode: item.StockUnitCode, Status: item.Status}
}

func toAPIError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, consts.ErrCategoryNotFound), errors.Is(err, consts.ErrSupplierRequired):
		return gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
	}
	var parameterError *consts.InvalidParameterError
	if errors.As(err, &parameterError) {
		return gerror.NewCode(gcode.CodeInvalidParameter, parameterError.Error())
	}
	glog.Errorf(ctx, "supplier products internal error: %+v", err)
	return gerror.NewCode(gcode.CodeInternalError, "internal server error")
}
