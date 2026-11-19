package procurementrequirements

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"

	"github.com/suxiangdong/indu-scope-backend/api/procurementrequirements/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/companyaccess"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	procurementRequirementModel "github.com/suxiangdong/indu-scope-backend/internal/model/procurementrequirement"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) FormOptions(ctx context.Context, req *v1.FormOptionsReq) (*v1.FormOptionsRes, error) {
	user, err := currentVerifiedUser(ctx)
	if err != nil {
		return nil, err
	}
	if user.Role != 1 && user.Role != 2 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "buyer or supplier authorization is required")
	}
	var categoryID *uint64
	if req != nil {
		categoryID = req.CategoryID
	}
	result, err := service.ProcurementRequirements().FormOptions(ctx, categoryID)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	res := &v1.FormOptionsRes{
		Categories: make([]v1.CategoryInfo, 0, len(result.Categories)),
		Attributes: make([]v1.AttributeInfo, 0, len(result.Attributes)),
		StockUnits: make([]v1.StockUnitInfo, 0, len(result.StockUnits)),
	}
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
	user, err := currentVerifiedUser(ctx)
	if err != nil {
		return nil, err
	}
	if user.Role != 1 && user.Role != 2 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "buyer or supplier authorization is required")
	}
	filter := procurementRequirementModel.ListFilter{UserID: uint64(user.ID), Page: 1, PageSize: 20}
	if user.Role == 1 {
		filter.Scope = "mine"
	} else if user.Role == 2 {
		filter.Scope = "open"
	} else {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "buyer or supplier authorization is required")
	}
	if req != nil {
		if req.Scope != "" {
			filter.Scope = req.Scope
		}
		filter.Keyword, filter.CategoryID = req.Keyword, req.CategoryID
		if req.Page != 0 {
			filter.Page = req.Page
		}
		if req.PageSize != 0 {
			filter.PageSize = req.PageSize
		}
	}
	if (user.Role == 1 && filter.Scope != "mine") || (user.Role == 2 && filter.Scope != "open") {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "the requested requirement scope is not available for this role")
	}
	result, err := service.ProcurementRequirements().List(ctx, filter)
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
	user, err := currentVerifiedUser(ctx)
	if err != nil {
		return nil, err
	}
	if user.Role != 1 && user.Role != 2 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "buyer or supplier authorization is required")
	}
	if req == nil || req.ID == 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "id is required")
	}
	result, err := service.ProcurementRequirements().Detail(ctx, uint64(user.ID), user.Role, req.ID)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	companyAllowed, err := companyaccess.Allowed(ctx, uint64(user.ID))
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	contacts := make([]v1.ContactInfo, 0, len(result.BuyerContacts))
	for _, contact := range result.BuyerContacts {
		contacts = append(contacts, v1.ContactInfo{Name: contact.Name, Position: contact.Position, Phone: contact.Phone})
	}
	res := &v1.DetailRes{
		Requirement:   toAPIRequirement(result.Requirement),
		CompanyInfo:   companyaccess.Summary(result.Requirement.BuyerUserID, result.Requirement.BuyerCompany, companyAllowed),
		CompanyLocked: !companyAllowed,
		Responses:     make([]v1.RequirementResponseInfo, 0, len(result.Responses)), BuyerContacts: contacts,
	}
	for _, item := range result.Responses {
		res.Responses = append(res.Responses, toAPIResponse(item, companyAllowed))
	}
	return res, nil
}

func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (*v1.CreateRes, error) {
	user, err := currentVerifiedUser(ctx)
	if err != nil {
		return nil, err
	}
	if user.Role != 1 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "buyer authorization is required")
	}
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "requirement is required")
	}
	attributes := make([]procurementRequirementModel.AttributeValueInput, 0, len(req.Attributes))
	for _, value := range req.Attributes {
		attributes = append(attributes, procurementRequirementModel.AttributeValueInput{AttributeID: value.AttributeID, Value: value.Value, OptionID: value.OptionID, OptionIDs: value.OptionIDs})
	}
	result, err := service.ProcurementRequirements().Create(ctx, procurementRequirementModel.CreateInput{
		BuyerUserID: uint64(user.ID), CategoryID: req.CategoryID, Name: req.Name, Quantity: req.Quantity,
		StockUnitID: req.StockUnitID, Description: req.Description,
		DeliveryAddress: req.DeliveryAddress, ExpectedDeliveryDate: req.ExpectedDeliveryDate,
		ContactName: req.ContactName, ContactPhone: req.ContactPhone, FilePaths: req.FilePaths, Attributes: attributes,
	})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.CreateRes{RequirementInfo: toAPIRequirement(result)}, nil
}

func (c *ControllerV1) UpdateStatus(ctx context.Context, req *v1.UpdateStatusReq) (*v1.UpdateStatusRes, error) {
	user, err := currentVerifiedUser(ctx)
	if err != nil {
		return nil, err
	}
	if user.Role != 1 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "buyer authorization is required")
	}
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "requirement status is required")
	}
	result, err := service.ProcurementRequirements().UpdateStatus(ctx, uint64(user.ID), req.ID, req.Status)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.UpdateStatusRes{RequirementInfo: toAPIRequirement(result)}, nil
}

func (c *ControllerV1) UploadFile(ctx context.Context, req *v1.UploadFileReq) (*v1.UploadFileRes, error) {
	user, err := currentVerifiedUser(ctx)
	if err != nil {
		return nil, err
	}
	if user.Role != 1 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "buyer authorization is required")
	}
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "requirement attachment is required")
	}
	upload := req.File
	if upload == nil {
		upload = g.RequestFromCtx(ctx).GetUploadFile("file")
	}
	if upload == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "requirement attachment is required")
	}
	file, err := upload.Open()
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "requirement attachment could not be opened")
	}
	defer file.Close()
	head := make([]byte, 512)
	read, readErr := io.ReadFull(file, head)
	if readErr != nil && readErr != io.ErrUnexpectedEOF && readErr != io.EOF {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "requirement attachment could not be read")
	}
	if read == 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "requirement attachment is empty")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "requirement attachment could not be read")
	}
	contentType := http.DetectContentType(head[:read])
	result, err := service.ProcurementRequirements().UploadFile(ctx, procurementRequirementModel.FileUploadInput{Reader: file, Filename: upload.Filename, ContentType: contentType, Size: upload.Size})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.UploadFileRes{Endpoint: result.Endpoint, Bucket: result.Bucket, Path: result.Path, OriginalName: result.OriginalName, ContentType: result.ContentType, Size: result.Size}, nil
}

func (c *ControllerV1) CreateResponse(ctx context.Context, req *v1.CreateResponseReq) (*v1.CreateResponseRes, error) {
	user, err := currentVerifiedUser(ctx)
	if err != nil {
		return nil, err
	}
	if user.Role != 2 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "supplier authorization is required")
	}
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "response is required")
	}
	result, err := service.ProcurementRequirements().CreateResponse(ctx, procurementRequirementModel.ResponseInput{
		RequirementID: req.RequirementID, SupplierUserID: uint64(user.ID), Message: req.Message,
		AvailableQuantity: req.AvailableQuantity, LeadTimeDays: req.LeadTimeDays,
	})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.CreateResponseRes{RequirementResponseInfo: toAPIResponse(result, false)}, nil
}

func (c *ControllerV1) UpdateResponseStatus(ctx context.Context, req *v1.UpdateResponseStatusReq) (*v1.UpdateResponseStatusRes, error) {
	user, err := currentVerifiedUser(ctx)
	if err != nil {
		return nil, err
	}
	if user.Role != 1 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "buyer authorization is required")
	}
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "response status is required")
	}
	result, err := service.ProcurementRequirements().UpdateResponseStatus(ctx, uint64(user.ID), req.ID, req.Status)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.UpdateResponseStatusRes{RequirementResponseInfo: toAPIResponse(result, false)}, nil
}

func currentVerifiedUser(ctx context.Context) (*authUser, error) {
	value := service.Bizctx().Get(ctx)
	if value == nil || value.User == nil {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "authorization is required")
	}
	if value.User.Status != 1 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "verified profile is required")
	}
	return &authUser{ID: value.User.ID, Role: value.User.Role}, nil
}

type authUser struct {
	ID   int64
	Role uint8
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
	return v1.RequirementInfo{
		ID: item.ID, BuyerUserID: item.BuyerUserID,
		CategoryID: item.CategoryID, CategoryName: item.CategoryName, CategoryNameSnapshot: item.CategoryNameSnapshot,
		Name: item.Name, Quantity: item.Quantity, StockUnitID: item.StockUnitID, StockUnitName: item.StockUnitName,
		StockUnitNameSnapshot: item.StockUnitNameSnapshot, Description: item.Description,
		DeliveryAddress: item.DeliveryAddress, ExpectedDeliveryDate: item.ExpectedDeliveryDate, ContactName: item.ContactName,
		ContactPhone: item.ContactPhone, ReviewStatus: item.ReviewStatus, ReviewReason: item.ReviewReason, Status: item.Status,
		PublishedAt: item.PublishedAt, ExpiresAt: item.ExpiresAt, ResponseCount: item.ResponseCount,
		CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt, Attributes: attributes, Files: files,
	}
}

func toAPIResponse(item procurementRequirementModel.ResponseItem, companyAllowed bool) v1.RequirementResponseInfo {
	return v1.RequirementResponseInfo{
		ID: item.ID, RequirementID: item.RequirementID, SupplierUserID: item.SupplierUserID,
		CompanyInfo: companyaccess.Summary(item.SupplierUserID, item.SupplierCompany, companyAllowed), CompanyLocked: !companyAllowed,
		Message: item.Message, AvailableQuantity: item.AvailableQuantity, LeadTimeDays: item.LeadTimeDays,
		Status: item.Status, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}

func toAPIError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, consts.ErrRequirementNotFound), errors.Is(err, consts.ErrRequirementClosed), errors.Is(err, consts.ErrRequirementResponseExists), errors.Is(err, consts.ErrRequirementResponseNotFound):
		return gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
	case errors.Is(err, consts.ErrCategoryNotFound), errors.Is(err, consts.ErrStockUnitNotFound):
		return gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
	}
	var parameterError *consts.InvalidParameterError
	if errors.As(err, &parameterError) {
		return gerror.NewCode(gcode.CodeInvalidParameter, parameterError.Error())
	}
	glog.Errorf(ctx, "procurement requirements internal error: %+v", err)
	return gerror.NewCode(gcode.CodeInternalError, "internal server error")
}
