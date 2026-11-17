package adminprocurementrequirements

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"

	"github.com/suxiangdong/indu-scope-backend/api/adminprocurementrequirements/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	procurementRequirementModel "github.com/suxiangdong/indu-scope-backend/internal/model/procurementrequirement"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (*v1.ListRes, error) {
	filter := procurementRequirementModel.AdminListFilter{Page: 1, PageSize: 20}
	if req != nil {
		filter.Keyword, filter.CategoryID, filter.BuyerUserID = req.Keyword, req.CategoryID, req.BuyerUserID
		filter.ReviewStatus, filter.Status = req.ReviewStatus, req.Status
		if req.Page != 0 {
			filter.Page = req.Page
		}
		if req.PageSize != 0 {
			filter.PageSize = req.PageSize
		}
	}
	result, err := service.ProcurementRequirements().AdminList(ctx, filter)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	items := make([]v1.RequirementInfo, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toAPIRequirement(item))
	}
	return &v1.ListRes{Items: items, Total: result.Total, Page: result.Page, PageSize: result.PageSize}, nil
}

func (c *ControllerV1) Detail(ctx context.Context, req *v1.DetailReq) (*v1.DetailRes, error) {
	if req == nil || req.ID == 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "id is required")
	}
	result, err := service.ProcurementRequirements().AdminDetail(ctx, req.ID)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	res := &v1.DetailRes{RequirementInfo: toAPIRequirement(result.Requirement), Responses: make([]v1.RequirementResponseInfo, 0, len(result.Responses))}
	for _, item := range result.Responses {
		res.Responses = append(res.Responses, toAPIResponse(item))
	}
	return res, nil
}

func (c *ControllerV1) Review(ctx context.Context, req *v1.ReviewReq) (*v1.ReviewRes, error) {
	value := service.Bizctx().Get(ctx)
	if value == nil || value.AdminUser == nil {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "admin authorization is required")
	}
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "review request is required")
	}
	result, err := service.ProcurementRequirements().Review(ctx, uint64(value.AdminUser.ID), req.ID, req.ReviewStatus, req.Reason)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.ReviewRes{RequirementInfo: toAPIRequirement(result.Requirement)}, nil
}

func toAPIRequirement(item procurementRequirementModel.RequirementItem) v1.RequirementInfo {
	attributes := make([]v1.RequirementAttributeValueInfo, 0, len(item.Attributes))
	for _, value := range item.Attributes {
		attributes = append(attributes, v1.RequirementAttributeValueInfo{AttributeID: value.AttributeID, AttributeName: value.AttributeName, Value: value.Value, OptionID: value.OptionID, OptionIDs: value.OptionIDs})
	}
	files := make([]v1.RequirementFileInfo, 0, len(item.Files))
	for _, file := range item.Files {
		files = append(files, v1.RequirementFileInfo{ID: file.ID, Endpoint: file.Endpoint, Bucket: file.Bucket, Path: file.Path, URL: file.URL, OriginalName: file.OriginalName, ContentType: file.ContentType, Size: file.Size, SortOrder: file.SortOrder, CreatedAt: file.CreatedAt})
	}
	return v1.RequirementInfo{ID: item.ID, BuyerUserID: item.BuyerUserID, BuyerCompany: item.BuyerCompany, CategoryID: item.CategoryID, CategoryName: item.CategoryName, CategoryNameSnapshot: item.CategoryNameSnapshot, Name: item.Name, Quantity: item.Quantity, StockUnitID: item.StockUnitID, StockUnitName: item.StockUnitName, StockUnitNameSnapshot: item.StockUnitNameSnapshot, Description: item.Description, DeliveryAddress: item.DeliveryAddress, ExpectedDeliveryDate: item.ExpectedDeliveryDate, ContactName: item.ContactName, ContactPhone: item.ContactPhone, ReviewStatus: item.ReviewStatus, ReviewReason: item.ReviewReason, Status: item.Status, PublishedAt: item.PublishedAt, ExpiresAt: item.ExpiresAt, ResponseCount: item.ResponseCount, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt, Attributes: attributes, Files: files}
}

func toAPIResponse(item procurementRequirementModel.ResponseItem) v1.RequirementResponseInfo {
	return v1.RequirementResponseInfo{ID: item.ID, RequirementID: item.RequirementID, SupplierUserID: item.SupplierUserID, SupplierCompany: item.SupplierCompany, Message: item.Message, AvailableQuantity: item.AvailableQuantity, LeadTimeDays: item.LeadTimeDays, Status: item.Status, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}

func toAPIError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, consts.ErrRequirementNotFound), errors.Is(err, consts.ErrRequirementClosed), errors.Is(err, consts.ErrRequirementResponseExists), errors.Is(err, consts.ErrRequirementResponseNotFound):
		return gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
	}
	var parameterError *consts.InvalidParameterError
	if errors.As(err, &parameterError) {
		return gerror.NewCode(gcode.CodeInvalidParameter, parameterError.Error())
	}
	glog.Errorf(ctx, "admin procurement requirements internal error: %+v", err)
	return gerror.NewCode(gcode.CodeInternalError, "internal server error")
}
