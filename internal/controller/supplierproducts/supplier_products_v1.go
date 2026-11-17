package supplierproducts

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
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
	res := &v1.FormOptionsRes{Categories: make([]v1.CategoryInfo, 0, len(result.Categories)), Attributes: make([]v1.AttributeInfo, 0, len(result.Attributes)), StockUnits: make([]v1.StockUnitInfo, 0, len(result.StockUnits))}
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
	for _, item := range result.StockUnits {
		res.StockUnits = append(res.StockUnits, v1.StockUnitInfo{ID: item.ID, Code: item.Code, Name: item.Name, SortOrder: item.SortOrder})
	}
	return res, nil
}

func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (*v1.ListRes, error) {
	if err := requireSupplier(ctx); err != nil {
		return nil, err
	}
	var page, pageSize uint
	if req != nil {
		page, pageSize = req.Page, req.PageSize
	}
	user := service.Bizctx().Get(ctx).User
	result, err := service.SupplierProducts().List(ctx, uint64(user.ID), page, pageSize)
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
	if err := requireSupplier(ctx); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "id is required")
	}
	user := service.Bizctx().Get(ctx).User
	result, err := service.SupplierProducts().Detail(ctx, uint64(user.ID), req.ID)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	contacts := make([]v1.ContactInfo, 0, len(result.Contacts))
	for _, contact := range result.Contacts {
		contacts = append(contacts, v1.ContactInfo{Name: contact.Name, Position: contact.Position, Phone: contact.Phone})
	}
	return &v1.DetailRes{ProductInfo: toAPI(result), SupplierContacts: contacts}, nil
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
		StockQuantity: req.StockQuantity, StockUnitID: req.StockUnitID, ImagePaths: req.ImagePaths, Attributes: valueInputs,
	})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.CreateRes{ProductInfo: toAPI(result)}, nil
}

func (c *ControllerV1) UpdateStatus(ctx context.Context, req *v1.UpdateStatusReq) (*v1.UpdateStatusRes, error) {
	if err := requireSupplier(ctx); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "product status is required")
	}
	user := service.Bizctx().Get(ctx).User
	result, err := service.SupplierProducts().UpdateStatus(ctx, uint64(user.ID), req.ID, req.Status)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.UpdateStatusRes{ProductInfo: toAPI(result)}, nil
}

func (c *ControllerV1) UpdateStock(ctx context.Context, req *v1.UpdateStockReq) (*v1.UpdateStockRes, error) {
	if err := requireSupplier(ctx); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "product stock is required")
	}
	user := service.Bizctx().Get(ctx).User
	result, err := service.SupplierProducts().UpdateStock(ctx, supplierProductModel.UpdateStockInput{
		SupplierUserID: uint64(user.ID), ProductID: req.ID, StockQuantity: req.StockQuantity,
	})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.UpdateStockRes{ProductInfo: toAPI(result)}, nil
}

func (c *ControllerV1) UploadImage(ctx context.Context, req *v1.UploadImageReq) (*v1.UploadImageRes, error) {
	if err := requireSupplier(ctx); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "product image is required")
	}
	upload := req.File
	if upload == nil {
		upload = g.RequestFromCtx(ctx).GetUploadFile("file")
	}
	if upload == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "product image is required")
	}
	file, err := upload.Open()
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "product image could not be opened")
	}
	defer file.Close()
	head := make([]byte, 512)
	read, readErr := io.ReadFull(file, head)
	if readErr != nil && readErr != io.ErrUnexpectedEOF && readErr != io.EOF {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "product image could not be read")
	}
	if read == 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "product image is empty")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "product image could not be read")
	}
	contentType := http.DetectContentType(head[:read])
	result, err := service.SupplierProducts().UploadImage(ctx, supplierProductModel.ProductImageUploadInput{
		Reader: file, Filename: upload.Filename, ContentType: contentType, Size: upload.Size,
	})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.UploadImageRes{Endpoint: result.Endpoint, Bucket: result.Bucket, Path: result.Path, OriginalName: result.OriginalName, ContentType: result.ContentType, Size: result.Size}, nil
}

func requireSupplier(ctx context.Context) error {
	value := service.Bizctx().Get(ctx)
	if value == nil || value.User == nil {
		return gerror.NewCode(gcode.CodeNotAuthorized, "authorization is required")
	}
	if value.User.Status != 1 {
		return gerror.NewCode(gcode.CodeNotAuthorized, "verified profile is required")
	}
	if value.User.Role != 2 {
		return gerror.NewCode(gcode.CodeNotAuthorized, "supplier authorization is required")
	}
	return nil
}

func toAPI(item supplierProductModel.ProductItem) v1.ProductInfo {
	images := make([]v1.ProductImageInfo, 0, len(item.Images))
	for _, image := range item.Images {
		images = append(images, v1.ProductImageInfo{ID: image.ID, Endpoint: image.Endpoint, Bucket: image.Bucket, Path: image.Path, URL: image.URL, OriginalName: image.OriginalName, ContentType: image.ContentType, Size: image.Size, SortOrder: image.SortOrder, CreatedAt: image.CreatedAt})
	}
	attributes := make([]v1.AttributeValueInfo, 0, len(item.Attributes))
	for _, attribute := range item.Attributes {
		attributes = append(attributes, v1.AttributeValueInfo{AttributeID: attribute.AttributeID, Code: attribute.Code, Name: attribute.Name, Unit: attribute.Unit, Value: attribute.Value})
	}
	return v1.ProductInfo{ID: item.ID, SupplierUserID: item.SupplierUserID, CategoryID: item.CategoryID, Name: item.Name, StockQuantity: item.StockQuantity, StockUnitID: item.StockUnitID, StockUnitCode: item.StockUnitCode, StockUnitName: item.StockUnitName, CategoryName: item.CategoryName, CoverURL: item.CoverURL, ImageCount: item.ImageCount, Images: images, Attributes: attributes, Status: item.Status, CreatedAt: item.CreatedAt}
}

func toAPIError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, consts.ErrCategoryNotFound), errors.Is(err, consts.ErrSupplierRequired), errors.Is(err, consts.ErrProductNotFound), errors.Is(err, consts.ErrProductPendingReview), errors.Is(err, consts.ErrProductUnavailable):
		return gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
	}
	var parameterError *consts.InvalidParameterError
	if errors.As(err, &parameterError) {
		return gerror.NewCode(gcode.CodeInvalidParameter, parameterError.Error())
	}
	glog.Errorf(ctx, "supplier products internal error: %+v", err)
	return gerror.NewCode(gcode.CodeInternalError, "internal server error")
}
