package adminproducts

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"

	"github.com/suxiangdong/indu-scope-backend/api/adminproducts/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	adminProductModel "github.com/suxiangdong/indu-scope-backend/internal/model/adminproduct"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (*v1.ListRes, error) {
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "product list request is required")
	}
	result, err := service.AdminProducts().List(ctx, adminProductModel.ListFilter{
		Keyword: req.Keyword, CategoryID: req.CategoryID, SupplierUserID: req.SupplierUserID,
		Status: req.Status, Page: req.Page, PageSize: req.PageSize,
	})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	items := make([]v1.ProductInfo, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toAPI(item))
	}
	return &v1.ListRes{Items: items, Total: result.Total, Page: result.Page, PageSize: result.PageSize}, nil
}

func (c *ControllerV1) Detail(ctx context.Context, req *v1.DetailReq) (*v1.DetailRes, error) {
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "product detail request is required")
	}
	item, err := service.AdminProducts().Detail(ctx, req.ID)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.DetailRes{ProductInfo: toAPI(item)}, nil
}

func (c *ControllerV1) UpdateStatus(ctx context.Context, req *v1.UpdateStatusReq) (*v1.UpdateStatusRes, error) {
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "product status request is required")
	}
	item, err := service.AdminProducts().UpdateStatus(ctx, req.ID, req.Status)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.UpdateStatusRes{ProductInfo: toAPI(item)}, nil
}

func (c *ControllerV1) Review(ctx context.Context, req *v1.ReviewReq) (*v1.ReviewRes, error) {
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "product review request is required")
	}
	item, err := service.AdminProducts().Review(ctx, req.ID)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.ReviewRes{ProductInfo: toAPI(item)}, nil
}

func toAPI(item adminProductModel.ProductItem) v1.ProductInfo {
	images := make([]v1.ProductImageInfo, 0, len(item.Images))
	for _, image := range item.Images {
		images = append(images, v1.ProductImageInfo{ID: image.ID, URL: image.URL, Path: image.Path, OriginalName: image.OriginalName, ContentType: image.ContentType, Size: image.Size, SortOrder: image.SortOrder})
	}
	attributes := make([]v1.AttributeValueInfo, 0, len(item.Attributes))
	for _, attribute := range item.Attributes {
		attributes = append(attributes, v1.AttributeValueInfo{AttributeID: attribute.AttributeID, Code: attribute.Code, Name: attribute.Name, Unit: attribute.Unit, Value: attribute.Value})
	}
	return v1.ProductInfo{
		ID: item.ID, SupplierUserID: item.SupplierUserID, SupplierCompany: item.SupplierCompany,
		CategoryID: item.CategoryID, CategoryName: item.CategoryName, Name: item.Name,
		StockQuantity: item.StockQuantity, StockUnitID: item.StockUnitID, StockUnitName: item.StockUnitName,
		CoverURL: item.CoverURL, ImageCount: item.ImageCount, Status: item.Status,
		CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt, Images: images, Attributes: attributes,
	}
}

func toAPIError(ctx context.Context, err error) error {
	if errors.Is(err, consts.ErrProductNotFound) || errors.Is(err, consts.ErrProductPendingReview) ||
		errors.Is(err, consts.ErrProductAlreadyReviewed) || errors.Is(err, consts.ErrProductStatusChanged) {
		return gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
	}
	var parameterError *consts.InvalidParameterError
	if errors.As(err, &parameterError) {
		return gerror.NewCode(gcode.CodeInvalidParameter, parameterError.Error())
	}
	glog.Errorf(ctx, "admin products internal error: %+v", err)
	return gerror.NewCode(gcode.CodeInternalError, "internal server error")
}
