package productcatalog

import (
	"context"
	"errors"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"

	"github.com/suxiangdong/indu-scope-backend/api/products/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	productCatalogModel "github.com/suxiangdong/indu-scope-backend/internal/model/productcatalog"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (*v1.ListRes, error) {
	userID, err := currentBuyerID(ctx)
	if err != nil {
		return nil, err
	}
	filter := productCatalogModel.ListFilter{}
	if req != nil {
		filter.Keyword = req.Keyword
		filter.CategoryID = req.CategoryID
		filter.Page = req.Page
		filter.PageSize = req.PageSize
	}
	result, err := service.ProductCatalog().List(ctx, userID, filter)
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
	userID, err := currentBuyerID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "id is required")
	}
	result, err := service.ProductCatalog().Detail(ctx, userID, req.ID)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	contacts := make([]v1.ContactInfo, 0, len(result.SupplierContacts))
	subscription, err := service.Vip().GetSubscription(ctx, userID)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	contactLocked := subscription == nil || (subscription.EndAt != nil && !subscription.EndAt.After(time.Now()))
	for _, contact := range result.SupplierContacts {
		phone := contact.Phone
		if contactLocked {
			phone = ""
		}
		contacts = append(contacts, v1.ContactInfo{Name: contact.Name, Position: contact.Position, Phone: phone})
	}
	return &v1.DetailRes{ProductInfo: toAPI(result.ProductItem), Images: toDetailImages(result.Images), Attributes: toDetailAttributes(result.Attributes), SupplierContacts: contacts, ContactLocked: contactLocked}, nil
}

func currentBuyerID(ctx context.Context) (uint64, error) {
	value := service.Bizctx().Get(ctx)
	if value == nil || value.User == nil {
		return 0, gerror.NewCode(gcode.CodeNotAuthorized, "authorization is required")
	}
	if value.User.Status != 1 || value.User.Role != 1 {
		return 0, gerror.NewCode(gcode.CodeNotAuthorized, "buyer authorization is required")
	}
	return uint64(value.User.ID), nil
}

func toAPI(item productCatalogModel.ProductItem) v1.ProductInfo {
	return v1.ProductInfo{
		ID: item.ID, SupplierUserID: item.SupplierUserID, SupplierCompany: item.SupplierCompany,
		CategoryID: item.CategoryID, CategoryName: item.CategoryName, Name: item.Name,
		StockQuantity: item.StockQuantity, StockUnitID: item.StockUnitID, StockUnitCode: item.StockUnitCode,
		StockUnitName: item.StockUnitName, CoverURL: item.CoverURL, ImageCount: item.ImageCount,
		Status: item.Status, Favorited: item.Favorited, CreatedAt: item.CreatedAt,
	}
}

func toDetailImages(images []productCatalogModel.ProductImage) []v1.ProductImageInfo {
	result := make([]v1.ProductImageInfo, 0, len(images))
	for _, image := range images {
		result = append(result, v1.ProductImageInfo{ID: image.ID, URL: image.URL, Path: image.Path, OriginalName: image.OriginalName, ContentType: image.ContentType, Size: image.Size, SortOrder: image.SortOrder})
	}
	return result
}

func toDetailAttributes(attributes []productCatalogModel.AttributeValue) []v1.AttributeValueInfo {
	result := make([]v1.AttributeValueInfo, 0, len(attributes))
	for _, attribute := range attributes {
		result = append(result, v1.AttributeValueInfo{AttributeID: attribute.AttributeID, Code: attribute.Code, Name: attribute.Name, Unit: attribute.Unit, Value: attribute.Value})
	}
	return result
}

func toAPIError(ctx context.Context, err error) error {
	if errors.Is(err, consts.ErrTokenInvalid) {
		return gerror.NewCode(gcode.CodeNotAuthorized, err.Error())
	}
	if errors.Is(err, consts.ErrProductNotFound) {
		return gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
	}
	if errors.Is(err, consts.ErrProductUnavailable) {
		return gerror.NewCode(gcode.CodeBusinessValidationFailed, "产品已下架")
	}
	var parameterError *consts.InvalidParameterError
	if errors.As(err, &parameterError) {
		return gerror.NewCode(gcode.CodeInvalidParameter, parameterError.Error())
	}
	glog.Errorf(ctx, "product catalog internal error: %+v", err)
	return gerror.NewCode(gcode.CodeInternalError, "internal server error")
}
