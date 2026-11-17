package procurementrequirements

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/suxiangdong/indu-scope-backend/internal/attributevalue"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	"github.com/suxiangdong/indu-scope-backend/internal/dao"
	procurementRequirementModel "github.com/suxiangdong/indu-scope-backend/internal/model/procurementrequirement"
	supplierProductModel "github.com/suxiangdong/indu-scope-backend/internal/model/supplierproduct"
	"github.com/suxiangdong/indu-scope-backend/internal/pagination"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
	"github.com/suxiangdong/indu-scope-backend/internal/storage"
)

const (
	requirementAttachmentPurpose = "procurement_requirement_attachment"
	maxRequirementAttachmentSize = 10 * 1024 * 1024
	maxRequirementAttachments    = 10
	defaultRequirementLifetime   = 30 * 24 * time.Hour
)

type sProcurementRequirements struct{}

var _ service.IProcurementRequirements = (*sProcurementRequirements)(nil)

func New() service.IProcurementRequirements { return &sProcurementRequirements{} }

func init() { service.RegisterProcurementRequirements(New()) }

func (s *sProcurementRequirements) FormOptions(ctx context.Context, categoryID *uint64) (procurementRequirementModel.FormOptions, error) {
	return service.SupplierProducts().FormOptions(ctx, categoryID)
}

func (s *sProcurementRequirements) List(ctx context.Context, filter procurementRequirementModel.ListFilter) (procurementRequirementModel.ListResult, error) {
	if filter.UserID == 0 {
		return procurementRequirementModel.ListResult{}, consts.ErrTokenInvalid
	}
	if err := s.expireRequirements(ctx); err != nil {
		return procurementRequirementModel.ListResult{}, err
	}
	filter.Page, filter.PageSize = pagination.Normalize(filter.Page, filter.PageSize)
	query, err := s.requirementQuery(ctx, filter)
	if err != nil {
		return procurementRequirementModel.ListResult{}, err
	}
	total, err := query.Count()
	if err != nil {
		return procurementRequirementModel.ListResult{}, gerror.Wrap(err, "count procurement requirements")
	}
	var records []requirementRecord
	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Fields(requirementFields()...).OrderDesc("created_at").OrderDesc("id").Limit(int(offset), int(filter.PageSize)).Scan(&records); err != nil {
		return procurementRequirementModel.ListResult{}, gerror.Wrap(err, "list procurement requirements")
	}
	if err := hydrateRequirementRecords(ctx, records); err != nil {
		return procurementRequirementModel.ListResult{}, err
	}
	items := make([]procurementRequirementModel.RequirementItem, 0, len(records))
	for _, record := range records {
		items = append(items, record.item())
	}
	return procurementRequirementModel.ListResult{Items: items, Total: uint64(total), Page: filter.Page, PageSize: filter.PageSize}, nil
}

func (s *sProcurementRequirements) AdminList(ctx context.Context, filter procurementRequirementModel.AdminListFilter) (procurementRequirementModel.ListResult, error) {
	if err := s.expireRequirements(ctx); err != nil {
		return procurementRequirementModel.ListResult{}, err
	}
	filter.Page, filter.PageSize = pagination.Normalize(filter.Page, filter.PageSize)
	query, err := s.adminRequirementQuery(ctx, filter)
	if err != nil {
		return procurementRequirementModel.ListResult{}, err
	}
	total, err := query.Count()
	if err != nil {
		return procurementRequirementModel.ListResult{}, gerror.Wrap(err, "count admin procurement requirements")
	}
	var records []requirementRecord
	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Fields(requirementFields()...).OrderDesc("created_at").OrderDesc("id").Limit(int(offset), int(filter.PageSize)).Scan(&records); err != nil {
		return procurementRequirementModel.ListResult{}, gerror.Wrap(err, "list admin procurement requirements")
	}
	if err := hydrateRequirementRecords(ctx, records); err != nil {
		return procurementRequirementModel.ListResult{}, err
	}
	items := make([]procurementRequirementModel.RequirementItem, 0, len(records))
	for _, record := range records {
		items = append(items, record.item())
	}
	return procurementRequirementModel.ListResult{Items: items, Total: uint64(total), Page: filter.Page, PageSize: filter.PageSize}, nil
}

func (s *sProcurementRequirements) AdminDetail(ctx context.Context, id uint64) (procurementRequirementModel.Detail, error) {
	if id == 0 {
		return procurementRequirementModel.Detail{}, &consts.InvalidParameterError{Message: "id is required"}
	}
	if err := s.expireRequirements(ctx); err != nil {
		return procurementRequirementModel.Detail{}, err
	}
	record, err := loadRequirementRecord(ctx, id)
	if err != nil {
		return procurementRequirementModel.Detail{}, err
	}
	item := record.item()
	attributes, err := s.loadAttributes(ctx, id)
	if err != nil {
		return procurementRequirementModel.Detail{}, err
	}
	item.Attributes = attributes
	files, err := s.loadFiles(ctx, id)
	if err != nil {
		return procurementRequirementModel.Detail{}, err
	}
	item.Files = files
	responses, err := s.loadResponses(ctx, id, 0, 1)
	if err != nil {
		return procurementRequirementModel.Detail{}, err
	}
	return procurementRequirementModel.Detail{Requirement: item, Responses: responses}, nil
}

func (s *sProcurementRequirements) Review(ctx context.Context, adminUserID, id uint64, reviewStatus uint8, reason string) (procurementRequirementModel.Detail, error) {
	if adminUserID == 0 {
		return procurementRequirementModel.Detail{}, consts.ErrAdminTokenInvalid
	}
	if id == 0 {
		return procurementRequirementModel.Detail{}, &consts.InvalidParameterError{Message: "id is required"}
	}
	if reviewStatus != 1 && reviewStatus != 2 {
		return procurementRequirementModel.Detail{}, &consts.InvalidParameterError{Message: "review_status must be 1 or 2"}
	}
	reason = strings.TrimSpace(reason)
	if len([]rune(reason)) > 500 {
		return procurementRequirementModel.Detail{}, &consts.InvalidParameterError{Message: "reason must be no longer than 500 characters"}
	}
	if reviewStatus == 2 && reason == "" {
		return procurementRequirementModel.Detail{}, &consts.InvalidParameterError{Message: "reason is required when rejecting a requirement"}
	}
	if _, err := loadRequirementRecord(ctx, id); err != nil {
		return procurementRequirementModel.Detail{}, gerror.Wrap(err, "query procurement requirement for review")
	}
	now := time.Now()
	data := g.Map{"review_status": reviewStatus, "review_reason": nullableString(reason), "reviewed_by": adminUserID, "reviewed_at": now}
	if reviewStatus == 1 {
		data["status"] = 1
		data["published_at"] = now
		data["expires_at"] = now.Add(defaultRequirementLifetime)
	} else {
		data["status"] = 5
		data["published_at"] = nil
		data["expires_at"] = nil
	}
	result, err := dao.ProcurementRequirements.Ctx(ctx).Cache(gdb.CacheOption{}).
		Where("id", id).Where("deleted_at IS NULL").Where("review_status", 0).Data(data).Update()
	if err != nil {
		return procurementRequirementModel.Detail{}, gerror.Wrap(err, "review procurement requirement")
	}
	if affected, err := result.RowsAffected(); err != nil {
		return procurementRequirementModel.Detail{}, gerror.Wrap(err, "check procurement requirement review")
	} else if affected == 0 {
		return procurementRequirementModel.Detail{}, consts.ErrRequirementClosed
	}
	return s.AdminDetail(ctx, id)
}

func (s *sProcurementRequirements) Detail(ctx context.Context, userID uint64, role uint8, id uint64) (procurementRequirementModel.Detail, error) {
	if userID == 0 {
		return procurementRequirementModel.Detail{}, consts.ErrTokenInvalid
	}
	if id == 0 {
		return procurementRequirementModel.Detail{}, &consts.InvalidParameterError{Message: "id is required"}
	}
	if err := s.expireRequirements(ctx); err != nil {
		return procurementRequirementModel.Detail{}, err
	}
	record, err := loadRequirementRecord(ctx, id)
	if err != nil {
		return procurementRequirementModel.Detail{}, err
	}
	if record.BuyerUserID != userID {
		if role != 2 {
			return procurementRequirementModel.Detail{}, consts.ErrRequirementNotFound
		}
		if record.ReviewStatus != 1 || record.Status != 1 || (record.ExpiresAt != nil && !record.ExpiresAt.After(time.Now())) {
			return procurementRequirementModel.Detail{}, consts.ErrRequirementNotFound
		}
	}
	if role == 2 && (record.ReviewStatus != 1 || record.Status != 1 || (record.ExpiresAt != nil && !record.ExpiresAt.After(time.Now()))) {
		return procurementRequirementModel.Detail{}, consts.ErrRequirementNotFound
	}
	if err := validateRequirementReferences(ctx, record); err != nil {
		return procurementRequirementModel.Detail{}, err
	}
	item := record.item()
	attributes, err := s.loadAttributes(ctx, id)
	if err != nil {
		return procurementRequirementModel.Detail{}, err
	}
	item.Attributes = attributes
	files, err := s.loadFiles(ctx, id)
	if err != nil {
		return procurementRequirementModel.Detail{}, err
	}
	item.Files = files
	responses, err := s.loadResponses(ctx, id, userID, role)
	if err != nil {
		return procurementRequirementModel.Detail{}, err
	}
	buyerProfile, err := service.UserProfile().Profile(ctx, record.BuyerUserID)
	if err != nil {
		return procurementRequirementModel.Detail{}, gerror.Wrap(err, "query procurement requirement buyer contacts")
	}
	return procurementRequirementModel.Detail{Requirement: item, Responses: responses, BuyerContacts: buyerProfile.Contacts}, nil
}

type preparedRequirement struct {
	Input         procurementRequirementModel.CreateInput
	ExpectedDate  any
	CategoryName  string
	StockUnitName string
	Definitions   map[uint64]supplierProductModel.AttributeItem
	Values        map[uint64]procurementRequirementModel.AttributeValueInput
	Files         []fileRecord
}

func (s *sProcurementRequirements) prepareInput(ctx context.Context, input procurementRequirementModel.CreateInput) (preparedRequirement, error) {
	prepared := preparedRequirement{Input: input}
	if input.BuyerUserID == 0 {
		return prepared, consts.ErrTokenInvalid
	}
	prepared.Input.Name = strings.TrimSpace(input.Name)
	prepared.Input.Description = strings.TrimSpace(input.Description)
	prepared.Input.DeliveryAddress = strings.TrimSpace(input.DeliveryAddress)
	prepared.Input.ContactName = strings.TrimSpace(input.ContactName)
	prepared.Input.ContactPhone = strings.TrimSpace(input.ContactPhone)
	if input.CategoryID == 0 {
		return prepared, &consts.InvalidParameterError{Message: "category_id is required"}
	}
	if prepared.Input.Name == "" || len([]rune(prepared.Input.Name)) > 255 {
		return prepared, &consts.InvalidParameterError{Message: "name is required and must be no longer than 255 characters"}
	}
	if prepared.Input.Description != "" && len([]rune(prepared.Input.Description)) > 2000 {
		return prepared, &consts.InvalidParameterError{Message: "description must be no longer than 2000 characters"}
	}
	if input.Quantity <= 0 || math.IsNaN(input.Quantity) || math.IsInf(input.Quantity, 0) {
		return prepared, &consts.InvalidParameterError{Message: "quantity must be greater than 0"}
	}
	if input.StockUnitID == 0 {
		return prepared, &consts.InvalidParameterError{Message: "stock_unit_id is required"}
	}
	if prepared.Input.DeliveryAddress == "" || len([]rune(prepared.Input.DeliveryAddress)) > 500 {
		return prepared, &consts.InvalidParameterError{Message: "delivery_address is required and must be no longer than 500 characters"}
	}
	if prepared.Input.ContactName == "" || len([]rune(prepared.Input.ContactName)) > 50 {
		return prepared, &consts.InvalidParameterError{Message: "contact_name is required and must be no longer than 50 characters"}
	}
	if prepared.Input.ContactPhone == "" || len([]rune(prepared.Input.ContactPhone)) > 30 {
		return prepared, &consts.InvalidParameterError{Message: "contact_phone is required and must be no longer than 30 characters"}
	}
	if strings.TrimSpace(input.ExpectedDeliveryDate) != "" {
		parsed, err := time.Parse("2006-01-02", strings.TrimSpace(input.ExpectedDeliveryDate))
		if err != nil {
			return prepared, &consts.InvalidParameterError{Message: "expected_delivery_date must be formatted as YYYY-MM-DD"}
		}
		prepared.ExpectedDate = parsed.Format("2006-01-02")
	}
	form, err := s.FormOptions(ctx, &input.CategoryID)
	if err != nil {
		return prepared, err
	}
	prepared.CategoryName, err = categoryName(form.Categories, input.CategoryID)
	if err != nil {
		return prepared, err
	}
	prepared.StockUnitName, err = stockUnitName(form.StockUnits, input.StockUnitID)
	if err != nil {
		return prepared, err
	}
	prepared.Definitions = make(map[uint64]supplierProductModel.AttributeItem, len(form.Attributes))
	for _, definition := range form.Attributes {
		prepared.Definitions[definition.ID] = definition
	}
	prepared.Values = make(map[uint64]procurementRequirementModel.AttributeValueInput, len(input.Attributes))
	for _, value := range input.Attributes {
		if value.AttributeID == 0 {
			return prepared, &consts.InvalidParameterError{Message: "attribute_id is required"}
		}
		definition, exists := prepared.Definitions[value.AttributeID]
		if !exists {
			return prepared, &consts.InvalidParameterError{Message: "attribute is not assigned to the selected category"}
		}
		if _, exists := prepared.Values[value.AttributeID]; exists {
			return prepared, &consts.InvalidParameterError{Message: "an attribute can only be submitted once"}
		}
		value.Value = strings.TrimSpace(value.Value)
		if value.Value == "" && value.OptionID == 0 && len(value.OptionIDs) == 0 {
			if definition.IsRequired == 1 {
				return prepared, &consts.InvalidParameterError{Message: "请填写必填属性：" + definition.Name}
			}
			continue
		}
		if err := attributevalue.Validate(definition, value); err != nil {
			return prepared, err
		}
		prepared.Values[value.AttributeID] = value
	}
	for _, definition := range form.Attributes {
		if _, exists := prepared.Values[definition.ID]; !exists && definition.IsRequired == 1 {
			return prepared, &consts.InvalidParameterError{Message: "请填写必填属性：" + definition.Name}
		}
	}
	prepared.Files, err = s.loadPendingFiles(ctx, input.BuyerUserID, input.FilePaths)
	if err != nil {
		return prepared, err
	}
	return prepared, nil
}

func insertRequirementRelations(tx gdb.TX, requirementID uint64, prepared preparedRequirement) error {
	for _, file := range prepared.Files {
		if _, err := tx.Model("procurement_requirement_files").Data(g.Map{
			"requirement_id": requirementID, "user_file_id": file.ID, "sort_order": file.SortOrder,
		}).Insert(); err != nil {
			return gerror.Wrap(err, "save procurement requirement file")
		}
	}
	for _, value := range prepared.Values {
		definition := prepared.Definitions[value.AttributeID]
		data, optionIDs, err := attributevalue.Data(definition, value)
		if err != nil {
			return err
		}
		data["requirement_id"] = requirementID
		data["attribute_id"] = value.AttributeID
		valueID, err := tx.Model("procurement_requirement_attribute_values").Data(data).InsertAndGetId()
		if err != nil {
			return gerror.Wrap(err, "save procurement requirement attribute value")
		}
		for _, optionID := range optionIDs {
			if _, err := tx.Model("procurement_requirement_attribute_value_options").Data(g.Map{"value_id": valueID, "option_id": optionID}).Insert(); err != nil {
				return gerror.Wrap(err, "save procurement requirement attribute option")
			}
		}
	}
	return nil
}

func (s *sProcurementRequirements) Create(ctx context.Context, input procurementRequirementModel.CreateInput) (procurementRequirementModel.RequirementItem, error) {
	prepared, err := s.prepareInput(ctx, input)
	if err != nil {
		return procurementRequirementModel.RequirementItem{}, err
	}
	var requirementID uint64
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		id, err := tx.Model("procurement_requirements").Data(g.Map{
			"buyer_user_id": prepared.Input.BuyerUserID, "category_id": prepared.Input.CategoryID, "category_name_snapshot": prepared.CategoryName,
			"name": prepared.Input.Name, "quantity": prepared.Input.Quantity, "stock_unit_id": prepared.Input.StockUnitID, "stock_unit_name_snapshot": prepared.StockUnitName,
			"description": nullableString(prepared.Input.Description), "delivery_address": prepared.Input.DeliveryAddress,
			"expected_delivery_date": prepared.ExpectedDate, "contact_name": prepared.Input.ContactName, "contact_phone": prepared.Input.ContactPhone,
			"review_status": 0, "status": 0,
		}).InsertAndGetId()
		if err != nil {
			return gerror.Wrap(err, "create procurement requirement")
		}
		requirementID = uint64(id)
		return insertRequirementRelations(tx, requirementID, prepared)
	})
	if err != nil {
		return procurementRequirementModel.RequirementItem{}, err
	}
	detail, err := s.Detail(ctx, input.BuyerUserID, 1, requirementID)
	if err != nil {
		return procurementRequirementModel.RequirementItem{}, err
	}
	return detail.Requirement, nil
}

func (s *sProcurementRequirements) UpdateStatus(ctx context.Context, userID, id uint64, status uint8) (procurementRequirementModel.RequirementItem, error) {
	if userID == 0 {
		return procurementRequirementModel.RequirementItem{}, consts.ErrTokenInvalid
	}
	if id == 0 {
		return procurementRequirementModel.RequirementItem{}, &consts.InvalidParameterError{Message: "id is required"}
	}
	if status != 1 && status != 2 {
		return procurementRequirementModel.RequirementItem{}, &consts.InvalidParameterError{Message: "requirement status must be 1 (online) or 2 (offline)"}
	}
	current, err := loadRequirementRecord(ctx, id)
	if err != nil {
		return procurementRequirementModel.RequirementItem{}, gerror.Wrap(err, "query procurement requirement status")
	}
	if current.BuyerUserID != userID {
		return procurementRequirementModel.RequirementItem{}, consts.ErrRequirementNotFound
	}
	if err := validateRequirementReferences(ctx, current); err != nil {
		return procurementRequirementModel.RequirementItem{}, err
	}
	if current.ReviewStatus != 1 || current.Status == 3 || current.Status == 4 {
		return procurementRequirementModel.RequirementItem{}, consts.ErrRequirementClosed
	}
	if status == 1 && current.ExpiresAt != nil && !current.ExpiresAt.After(time.Now()) {
		return procurementRequirementModel.RequirementItem{}, consts.ErrRequirementClosed
	}
	if current.Status == status {
		detail, err := s.Detail(ctx, userID, 1, id)
		return detail.Requirement, err
	}
	result, err := dao.ProcurementRequirements.Ctx(ctx).Cache(gdb.CacheOption{}).
		Where(g.Map{"id": id, "buyer_user_id": userID, "status": current.Status}).
		Where("review_status", 1).
		Where("deleted_at IS NULL").
		Data(g.Map{"status": status}).Update()
	if err != nil {
		return procurementRequirementModel.RequirementItem{}, gerror.Wrap(err, "update procurement requirement status")
	}
	if affected, err := result.RowsAffected(); err != nil {
		return procurementRequirementModel.RequirementItem{}, gerror.Wrap(err, "check procurement requirement status")
	} else if affected == 0 {
		var latest struct{ Status uint8 }
		if err := dao.ProcurementRequirements.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("status").Where(g.Map{"id": id, "buyer_user_id": userID}).Scan(&latest); err != nil {
			return procurementRequirementModel.RequirementItem{}, gerror.Wrap(err, "check current procurement requirement status")
		}
		if latest.Status == status {
			detail, err := s.Detail(ctx, userID, 1, id)
			return detail.Requirement, err
		}
		return procurementRequirementModel.RequirementItem{}, consts.ErrRequirementClosed
	}
	detail, err := s.Detail(ctx, userID, 1, id)
	return detail.Requirement, err
}

func (s *sProcurementRequirements) CreateResponse(ctx context.Context, input procurementRequirementModel.ResponseInput) (procurementRequirementModel.ResponseItem, error) {
	if input.SupplierUserID == 0 {
		return procurementRequirementModel.ResponseItem{}, consts.ErrTokenInvalid
	}
	if input.RequirementID == 0 {
		return procurementRequirementModel.ResponseItem{}, &consts.InvalidParameterError{Message: "requirement_id is required"}
	}
	input.Message = strings.TrimSpace(input.Message)
	if len([]rune(input.Message)) > 2000 {
		return procurementRequirementModel.ResponseItem{}, &consts.InvalidParameterError{Message: "message must be no longer than 2000 characters"}
	}
	if input.AvailableQuantity != nil && (*input.AvailableQuantity < 0 || math.IsNaN(*input.AvailableQuantity) || math.IsInf(*input.AvailableQuantity, 0)) {
		return procurementRequirementModel.ResponseItem{}, &consts.InvalidParameterError{Message: "available_quantity must be greater than or equal to 0"}
	}
	if input.LeadTimeDays != nil && *input.LeadTimeDays > 36500 {
		return procurementRequirementModel.ResponseItem{}, &consts.InvalidParameterError{Message: "lead_time_days is invalid"}
	}
	if err := s.expireRequirements(ctx); err != nil {
		return procurementRequirementModel.ResponseItem{}, err
	}
	responseID, err := createResponse(ctx, g.DB(), input)
	if err != nil {
		return procurementRequirementModel.ResponseItem{}, err
	}
	return s.loadResponse(ctx, responseID)
}

// Lock the parent before creating any child rows. InnoDB's foreign-key check
// otherwise takes a shared parent lock that concurrent inserts both need to
// upgrade when incrementing response_count.
func createResponse(ctx context.Context, db gdb.DB, input procurementRequirementModel.ResponseInput) (uint64, error) {
	var responseID uint64
	err := db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var requirement struct {
			ID           uint64
			Status       uint8
			ReviewStatus uint8
			ExpiresAt    *time.Time `orm:"expires_at"`
		}
		if err := tx.Model("procurement_requirements").Fields("id", "status", "review_status", "expires_at").
			Where("id", input.RequirementID).Where("deleted_at IS NULL").LockUpdate().Scan(&requirement); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return consts.ErrRequirementNotFound
			}
			return gerror.Wrap(err, "query procurement requirement for response")
		}
		if requirement.ID == 0 {
			return consts.ErrRequirementNotFound
		}
		if requirement.ReviewStatus != 1 || requirement.Status != 1 || (requirement.ExpiresAt != nil && !requirement.ExpiresAt.After(time.Now())) {
			return consts.ErrRequirementClosed
		}
		id, err := tx.Model("procurement_requirement_responses").Data(g.Map{
			"requirement_id": input.RequirementID, "supplier_user_id": input.SupplierUserID,
			"message": nullableString(input.Message), "available_quantity": input.AvailableQuantity, "lead_time_days": input.LeadTimeDays, "status": 0,
		}).InsertAndGetId()
		if err != nil {
			if isDuplicateKeyError(err) {
				return consts.ErrRequirementResponseExists
			}
			return gerror.Wrap(err, "create procurement requirement response")
		}
		responseID = uint64(id)
		result, err := tx.Model("procurement_requirements").Where(g.Map{"id": input.RequirementID, "review_status": 1, "status": 1}).
			Where("expires_at IS NULL OR expires_at > NOW()").Increment("response_count", 1)
		if err != nil {
			return gerror.Wrap(err, "increment procurement requirement response count")
		}
		if affected, err := result.RowsAffected(); err != nil {
			return gerror.Wrap(err, "check procurement requirement response count")
		} else if affected == 0 {
			return consts.ErrRequirementClosed
		}
		return nil
	})
	return responseID, err
}

func (s *sProcurementRequirements) UpdateResponseStatus(ctx context.Context, userID, responseID uint64, status uint8) (procurementRequirementModel.ResponseItem, error) {
	if userID == 0 {
		return procurementRequirementModel.ResponseItem{}, consts.ErrTokenInvalid
	}
	if responseID == 0 {
		return procurementRequirementModel.ResponseItem{}, &consts.InvalidParameterError{Message: "id is required"}
	}
	if status != 1 && status != 2 && status != 3 {
		return procurementRequirementModel.ResponseItem{}, &consts.InvalidParameterError{Message: "response status must be 1, 2 or 3"}
	}
	var record struct {
		ID            uint64
		RequirementID uint64
	}
	if err := dao.ProcurementRequirementResponses.Ctx(ctx).Cache(gdb.CacheOption{}).
		Fields("id", "requirement_id").Where("id", responseID).Scan(&record); err != nil {
		return procurementRequirementModel.ResponseItem{}, gerror.Wrap(err, "query procurement requirement response")
	}
	if record.ID == 0 {
		return procurementRequirementModel.ResponseItem{}, consts.ErrRequirementResponseNotFound
	}
	var requirement struct {
		ID          uint64
		BuyerUserID uint64
	}
	if err := dao.ProcurementRequirements.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("id", "buyer_user_id").Where("id", record.RequirementID).Scan(&requirement); err != nil {
		return procurementRequirementModel.ResponseItem{}, gerror.Wrap(err, "query procurement requirement for response")
	}
	if requirement.ID == 0 || requirement.BuyerUserID != userID {
		return procurementRequirementModel.ResponseItem{}, consts.ErrRequirementResponseNotFound
	}
	if _, err := dao.ProcurementRequirementResponses.Ctx(ctx).Cache(gdb.CacheOption{}).Where("id", responseID).Data(g.Map{"status": status}).Update(); err != nil {
		return procurementRequirementModel.ResponseItem{}, gerror.Wrap(err, "update procurement requirement response status")
	}
	return s.loadResponse(ctx, responseID)
}

func (s *sProcurementRequirements) UploadFile(ctx context.Context, input procurementRequirementModel.FileUploadInput) (procurementRequirementModel.FileUploadResult, error) {
	value := service.Bizctx().Get(ctx)
	if value == nil || value.User == nil || value.User.Status != 1 || value.User.Role != 1 {
		return procurementRequirementModel.FileUploadResult{}, consts.ErrBuyerRequired
	}
	if input.Reader == nil {
		return procurementRequirementModel.FileUploadResult{}, &consts.InvalidParameterError{Message: "requirement attachment is required"}
	}
	if input.Size <= 0 || input.Size > maxRequirementAttachmentSize {
		return procurementRequirementModel.FileUploadResult{}, &consts.InvalidParameterError{Message: "requirement attachment must be between 1 byte and 10 MB"}
	}
	contentType := strings.ToLower(strings.TrimSpace(input.ContentType))
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		return procurementRequirementModel.FileUploadResult{}, &consts.InvalidParameterError{Message: "requirement attachment must be a JPG, PNG or WEBP image"}
	}
	filename := filepath.Base(strings.TrimSpace(input.Filename))
	if filename == "." || filename == "" {
		filename = "requirement-attachment" + extensionForImageContentType(contentType)
	}
	if filepath.Ext(filename) == "" {
		filename += extensionForImageContentType(contentType)
	}
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return procurementRequirementModel.FileUploadResult{}, gerror.Wrap(err, "generate procurement requirement attachment object key")
	}
	client, err := storage.OSSClient()
	if err != nil {
		return procurementRequirementModel.FileUploadResult{}, gerror.Wrap(err, "initialize oss client")
	}
	objectKey := client.ObjectKey("procurement/requirements", fmt.Sprintf("%d-%d-%s%s", value.User.ID, time.Now().UnixNano(), hex.EncodeToString(nonce), extensionForImageContentType(contentType)))
	etag, err := client.Upload(ctx, objectKey, input.Reader, input.Size, contentType)
	if err != nil {
		_ = storage.DeleteWithTimeout(ctx, client, objectKey)
		return procurementRequirementModel.FileUploadResult{}, gerror.Wrap(err, "upload procurement requirement attachment to oss")
	}
	if _, err := dao.UserFiles.Ctx(ctx).Data(g.Map{
		"user_id": value.User.ID, "purpose": requirementAttachmentPurpose, "endpoint": client.EndpointHost(),
		"bucket": client.Bucket(), "object_key": objectKey, "original_name": filename,
		"content_type": contentType, "file_size": input.Size, "etag": etag,
	}).Insert(); err != nil {
		_ = storage.DeleteWithTimeout(ctx, client, objectKey)
		return procurementRequirementModel.FileUploadResult{}, gerror.Wrap(err, "save procurement requirement attachment metadata")
	}
	return procurementRequirementModel.FileUploadResult{Endpoint: client.EndpointHost(), Bucket: client.Bucket(), Path: objectKey, OriginalName: filename, ContentType: contentType, Size: input.Size}, nil
}

func (s *sProcurementRequirements) requirementQuery(ctx context.Context, filter procurementRequirementModel.ListFilter) (*gdb.Model, error) {
	model := dao.ProcurementRequirements.Ctx(ctx).Where("deleted_at IS NULL")
	buyerIDs, err := activeRequirementBuyerIDs(ctx)
	if err != nil {
		return nil, err
	}
	categoryIDs, err := activeRequirementCategoryIDs(ctx)
	if err != nil {
		return nil, err
	}
	unitIDs, err := activeRequirementStockUnitIDs(ctx)
	if err != nil {
		return nil, err
	}
	model = applyRequirementIDFilter(model, "buyer_user_id", buyerIDs)
	model = applyRequirementIDFilter(model, "category_id", categoryIDs)
	model = applyRequirementIDFilter(model, "stock_unit_id", unitIDs)
	if filter.Scope == "mine" {
		model = model.Where("buyer_user_id", filter.UserID)
	} else {
		model = model.Where(g.Map{"review_status": 1, "status": 1}).Where("expires_at IS NULL OR expires_at > NOW()")
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		pattern := "%" + keyword + "%"
		buyerKeywordIDs, err := requirementBuyerIDsByKeyword(ctx, pattern)
		if err != nil {
			return nil, err
		}
		model = applyRequirementKeywordFilter(model, pattern, buyerKeywordIDs)
	}
	if filter.CategoryID != nil && *filter.CategoryID != 0 {
		model = model.Where("category_id", *filter.CategoryID)
	}
	return model, nil
}

func (s *sProcurementRequirements) expireRequirements(ctx context.Context) error {
	result, err := dao.ProcurementRequirements.Ctx(ctx).Cache(gdb.CacheOption{}).
		Where(g.Map{"review_status": 1, "status": 1}).Where("deleted_at IS NULL").
		Where("expires_at IS NOT NULL AND expires_at <= NOW()").
		Data(g.Map{"status": 4}).Update()
	if err != nil {
		return gerror.Wrap(err, "expire procurement requirements")
	}
	if _, err := result.RowsAffected(); err != nil {
		return gerror.Wrap(err, "check expired procurement requirements")
	}
	return nil
}

func (s *sProcurementRequirements) adminRequirementQuery(ctx context.Context, filter procurementRequirementModel.AdminListFilter) (*gdb.Model, error) {
	model := dao.ProcurementRequirements.Ctx(ctx).Where("deleted_at IS NULL")
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		pattern := "%" + keyword + "%"
		buyerKeywordIDs, err := requirementBuyerIDsByKeyword(ctx, pattern)
		if err != nil {
			return nil, err
		}
		model = applyRequirementKeywordFilter(model, pattern, buyerKeywordIDs)
	}
	if filter.CategoryID != nil && *filter.CategoryID != 0 {
		model = model.Where("category_id", *filter.CategoryID)
	}
	if filter.BuyerUserID != nil && *filter.BuyerUserID != 0 {
		model = model.Where("buyer_user_id", *filter.BuyerUserID)
	}
	if filter.ReviewStatus != nil {
		model = model.Where("review_status", *filter.ReviewStatus)
	}
	if filter.Status != nil {
		model = model.Where("status", *filter.Status)
	}
	return model, nil
}

func activeRequirementBuyerIDs(ctx context.Context) ([]uint64, error) {
	var rows []struct{ ID uint64 }
	if err := dao.Users.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("id").Where(g.Map{"status": 1, "role": 1}).Scan(&rows); err != nil {
		return nil, gerror.Wrap(err, "query active requirement buyers")
	}
	ids := make([]uint64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids, nil
}

func activeRequirementCategoryIDs(ctx context.Context) ([]uint64, error) {
	var rows []struct{ ID uint64 }
	if err := dao.Categories.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("id").Where("status", 1).Scan(&rows); err != nil {
		return nil, gerror.Wrap(err, "query active requirement categories")
	}
	ids := make([]uint64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids, nil
}

func activeRequirementStockUnitIDs(ctx context.Context) ([]uint64, error) {
	var rows []struct{ ID uint64 }
	if err := dao.StockUnits.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("id").Where("status", 1).Scan(&rows); err != nil {
		return nil, gerror.Wrap(err, "query active requirement stock units")
	}
	ids := make([]uint64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids, nil
}

func requirementBuyerIDsByKeyword(ctx context.Context, pattern string) ([]uint64, error) {
	var rows []struct{ ID uint64 }
	if err := dao.Users.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("id").WhereLike("company_name", pattern).Scan(&rows); err != nil {
		return nil, gerror.Wrap(err, "query requirement buyers by keyword")
	}
	ids := make([]uint64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids, nil
}

func applyRequirementIDFilter(model *gdb.Model, column string, ids []uint64) *gdb.Model {
	if len(ids) == 0 {
		return model.Where("1=0")
	}
	return model.WhereIn(column, ids)
}

func applyRequirementKeywordFilter(model *gdb.Model, pattern string, buyerIDs []uint64) *gdb.Model {
	conditions := []string{"name LIKE ?", "category_name_snapshot LIKE ?"}
	args := []any{pattern, pattern}
	if len(buyerIDs) > 0 {
		placeholders := strings.TrimRight(strings.Repeat("?,", len(buyerIDs)), ",")
		conditions = append(conditions, "buyer_user_id IN ("+placeholders+")")
		for _, id := range buyerIDs {
			args = append(args, id)
		}
	}
	return model.Where("("+strings.Join(conditions, " OR ")+")", args...)
}

func hydrateRequirementRecords(ctx context.Context, records []requirementRecord) error {
	buyerIDs := make(map[uint64]struct{}, len(records))
	categoryIDs := make(map[uint64]struct{}, len(records))
	unitIDs := make(map[uint64]struct{}, len(records))
	for _, record := range records {
		buyerIDs[record.BuyerUserID] = struct{}{}
		categoryIDs[record.CategoryID] = struct{}{}
		unitIDs[record.StockUnitID] = struct{}{}
	}
	buyers := make(map[uint64]string, len(buyerIDs))
	for id := range buyerIDs {
		var buyer struct {
			ID          uint64
			CompanyName string
			Status      uint8
			Role        uint8
		}
		if err := dao.Users.Ctx(ctx).Cache(gdb.CacheOption{}).
			Fields("id", "company_name", "status", "role").Where("id", id).Scan(&buyer); err != nil {
			return gerror.Wrap(err, "query requirement buyer")
		}
		buyers[id] = buyer.CompanyName
	}
	categories := make(map[uint64]string, len(categoryIDs))
	for id := range categoryIDs {
		var category struct {
			ID     uint64
			Name   string
			Status uint8
		}
		if err := dao.Categories.Ctx(ctx).Cache(gdb.CacheOption{}).
			Fields("id", "name", "status").Where("id", id).Scan(&category); err != nil {
			return gerror.Wrap(err, "query requirement category")
		}
		categories[id] = category.Name
	}
	units := make(map[uint64]string, len(unitIDs))
	for id := range unitIDs {
		var unit struct {
			ID     uint64
			Name   string
			Status uint8
		}
		if err := dao.StockUnits.Ctx(ctx).Cache(gdb.CacheOption{}).
			Fields("id", "name", "status").Where("id", id).Scan(&unit); err != nil {
			return gerror.Wrap(err, "query requirement stock unit")
		}
		units[id] = unit.Name
	}
	for index := range records {
		records[index].BuyerCompany = buyers[records[index].BuyerUserID]
		records[index].CategoryName = categories[records[index].CategoryID]
		records[index].StockUnitName = units[records[index].StockUnitID]
	}
	return nil
}

func requirementFields() []any {
	return []any{
		"id", "buyer_user_id", "category_id", "category_name_snapshot", "name", "quantity",
		"stock_unit_id", "stock_unit_name_snapshot", "description", "delivery_address", "expected_delivery_date", "contact_name",
		"contact_phone", "review_status", "review_reason", "status", "published_at", "expires_at", "response_count", "created_at", "updated_at",
	}
}

type requirementRecord struct {
	ID                    uint64
	BuyerUserID           uint64
	BuyerCompany          string
	CategoryID            uint64
	CategoryName          string
	CategoryNameSnapshot  string
	Name                  string
	Quantity              float64
	StockUnitID           uint64
	StockUnitName         string
	StockUnitNameSnapshot string
	Description           string
	DeliveryAddress       string
	ExpectedDeliveryDate  *time.Time `orm:"expected_delivery_date"`
	ContactName           string
	ContactPhone          string
	ReviewStatus          uint8
	ReviewReason          string
	Status                uint8
	PublishedAt           *time.Time `orm:"published_at"`
	ExpiresAt             *time.Time `orm:"expires_at"`
	ResponseCount         uint
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// loadRequirementRecord intentionally reads the requirement row by its primary
// key and resolves the related display names in separate primary-key queries.
// This keeps the hot single-record path stable so GoFrame can reuse the query
// cache for the requirement itself and for each referenced entity.
func loadRequirementRecord(ctx context.Context, id uint64) (requirementRecord, error) {
	var record requirementRecord
	if err := dao.ProcurementRequirements.Ctx(ctx).Cache(gdb.CacheOption{}).
		Fields("id", "buyer_user_id", "category_id", "category_name_snapshot", "name", "quantity", "stock_unit_id", "stock_unit_name_snapshot", "description", "delivery_address", "expected_delivery_date", "contact_name", "contact_phone", "review_status", "review_reason", "status", "published_at", "expires_at", "response_count", "created_at", "updated_at").
		Where("id", id).Where("deleted_at IS NULL").Scan(&record); err != nil {
		return requirementRecord{}, gerror.Wrap(err, "query procurement requirement")
	}
	if record.ID == 0 {
		return requirementRecord{}, consts.ErrRequirementNotFound
	}

	var buyer struct {
		ID          uint64
		CompanyName string
		Status      uint8
		Role        uint8
	}
	if err := dao.Users.Ctx(ctx).Cache(gdb.CacheOption{}).
		Fields("id", "company_name", "status", "role").Where("id", record.BuyerUserID).Scan(&buyer); err != nil {
		return requirementRecord{}, gerror.Wrap(err, "query procurement requirement buyer")
	}
	record.BuyerCompany = buyer.CompanyName

	var category struct {
		ID     uint64
		Name   string
		Status uint8
	}
	if err := dao.Categories.Ctx(ctx).Cache(gdb.CacheOption{}).
		Fields("id", "name", "status").Where("id", record.CategoryID).Scan(&category); err != nil {
		return requirementRecord{}, gerror.Wrap(err, "query procurement requirement category")
	}
	if category.ID != 0 {
		record.CategoryName = category.Name
	}

	var unit struct {
		ID     uint64
		Name   string
		Status uint8
	}
	if err := dao.StockUnits.Ctx(ctx).Cache(gdb.CacheOption{}).
		Fields("id", "name", "status").Where("id", record.StockUnitID).Scan(&unit); err != nil {
		return requirementRecord{}, gerror.Wrap(err, "query procurement requirement stock unit")
	}
	if unit.ID != 0 {
		record.StockUnitName = unit.Name
	}
	return record, nil
}

func validateRequirementReferences(ctx context.Context, record requirementRecord) error {
	var buyer struct {
		ID     uint64
		Status uint8
		Role   uint8
	}
	if err := dao.Users.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("id", "status", "role").Where("id", record.BuyerUserID).Scan(&buyer); err != nil {
		return gerror.Wrap(err, "query procurement requirement buyer status")
	}
	var category struct {
		ID     uint64
		Status uint8
	}
	if err := dao.Categories.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("id", "status").Where("id", record.CategoryID).Scan(&category); err != nil {
		return gerror.Wrap(err, "query procurement requirement category status")
	}
	var unit struct {
		ID     uint64
		Status uint8
	}
	if err := dao.StockUnits.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("id", "status").Where("id", record.StockUnitID).Scan(&unit); err != nil {
		return gerror.Wrap(err, "query procurement requirement stock unit status")
	}
	if buyer.ID == 0 || buyer.Status != 1 || buyer.Role != 1 || category.ID == 0 || category.Status != 1 || unit.ID == 0 || unit.Status != 1 {
		return consts.ErrRequirementNotFound
	}
	return nil
}

func (r requirementRecord) item() procurementRequirementModel.RequirementItem {
	return procurementRequirementModel.RequirementItem{
		ID: r.ID, BuyerUserID: r.BuyerUserID, BuyerCompany: r.BuyerCompany, CategoryID: r.CategoryID,
		CategoryName: r.CategoryName, CategoryNameSnapshot: r.CategoryNameSnapshot, Name: r.Name, Quantity: r.Quantity,
		StockUnitID: r.StockUnitID, StockUnitName: r.StockUnitName, StockUnitNameSnapshot: r.StockUnitNameSnapshot,
		Description: r.Description, DeliveryAddress: r.DeliveryAddress, ExpectedDeliveryDate: r.ExpectedDeliveryDate,
		ContactName: r.ContactName, ContactPhone: r.ContactPhone, ReviewStatus: r.ReviewStatus, ReviewReason: r.ReviewReason,
		Status: r.Status, PublishedAt: r.PublishedAt, ExpiresAt: r.ExpiresAt, ResponseCount: r.ResponseCount,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt, Attributes: []procurementRequirementModel.AttributeValueItem{}, Files: []procurementRequirementModel.FileItem{},
	}
}

type attributeValueRecord struct {
	ID            uint64
	AttributeID   uint64
	AttributeName string
	ValueNumber   *float64 `orm:"value_number"`
	ValueText     *string  `orm:"value_text"`
	ValueBool     *uint    `orm:"value_bool"`
	ValueOptionID *uint64  `orm:"value_option_id"`
	ValueJSON     *string  `orm:"value_json"`
}

func (s *sProcurementRequirements) loadAttributes(ctx context.Context, requirementID uint64) ([]procurementRequirementModel.AttributeValueItem, error) {
	var records []attributeValueRecord
	if err := dao.ProcurementRequirementAttributeValues.Ctx(ctx).
		Fields("id", "attribute_id", "value_number", "value_text", "value_bool", "value_option_id", "value_json").
		Where("requirement_id", requirementID).OrderAsc("id").Scan(&records); err != nil {
		return nil, gerror.Wrap(err, "query procurement requirement attributes")
	}
	items := make([]procurementRequirementModel.AttributeValueItem, 0, len(records))
	for _, record := range records {
		value, err := attributeRecordValue(ctx, record)
		if err != nil {
			return nil, err
		}
		var attribute struct {
			ID   uint64
			Name string
		}
		if err := dao.Attributes.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("id", "name").Where("id", record.AttributeID).Scan(&attribute); err != nil {
			return nil, gerror.Wrap(err, "query procurement requirement attribute definition")
		}
		item := procurementRequirementModel.AttributeValueItem{AttributeID: record.AttributeID, AttributeName: attribute.Name, Value: value}
		if record.ValueOptionID != nil {
			item.OptionID = *record.ValueOptionID
		}
		var options []struct{ ID uint64 }
		if err := dao.ProcurementRequirementAttributeValueOptions.Ctx(ctx).Fields("option_id AS id").Where("value_id", record.ID).OrderAsc("option_id").Scan(&options); err != nil {
			return nil, gerror.Wrap(err, "query procurement requirement attribute options")
		}
		for _, option := range options {
			item.OptionIDs = append(item.OptionIDs, option.ID)
		}
		items = append(items, item)
	}
	return items, nil
}

func attributeRecordValue(ctx context.Context, record attributeValueRecord) (string, error) {
	if record.ValueOptionID != nil && *record.ValueOptionID != 0 {
		var option struct{ Name string }
		if err := dao.AttributeOptions.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("name").Where("id", *record.ValueOptionID).Scan(&option); err != nil {
			return "", gerror.Wrap(err, "query procurement requirement attribute option")
		}
		if option.Name != "" {
			return option.Name, nil
		}
	}
	if record.ValueText != nil {
		return *record.ValueText, nil
	}
	if record.ValueNumber != nil {
		return strconv.FormatFloat(*record.ValueNumber, 'f', -1, 64), nil
	}
	if record.ValueBool != nil {
		if *record.ValueBool == 1 {
			return "是", nil
		}
		return "否", nil
	}
	if record.ValueJSON != nil {
		return *record.ValueJSON, nil
	}
	return "", nil
}

type fileRecord struct {
	ID           uint64
	Endpoint     string
	Bucket       string
	ObjectKey    string `orm:"object_key"`
	OriginalName string `orm:"original_name"`
	ContentType  string `orm:"content_type"`
	FileSize     int64  `orm:"file_size"`
	SortOrder    uint   `orm:"sort_order"`
	CreatedAt    time.Time
}

func (s *sProcurementRequirements) loadFiles(ctx context.Context, requirementID uint64) ([]procurementRequirementModel.FileItem, error) {
	var links []struct {
		ID         uint64
		UserFileID uint64    `orm:"user_file_id"`
		SortOrder  uint      `orm:"sort_order"`
		CreatedAt  time.Time `orm:"created_at"`
	}
	if err := dao.ProcurementRequirementFiles.Ctx(ctx).
		Fields("id", "user_file_id", "sort_order", "created_at").
		Where("requirement_id", requirementID).OrderAsc("sort_order").OrderAsc("id").Scan(&links); err != nil {
		return nil, gerror.Wrap(err, "query procurement requirement files")
	}
	client, _ := storage.OSSClient()
	items := make([]procurementRequirementModel.FileItem, 0, len(links))
	for _, link := range links {
		var file fileRecord
		if err := dao.UserFiles.Ctx(ctx).Cache(gdb.CacheOption{}).
			Fields("id", "endpoint", "bucket", "object_key", "original_name", "content_type", "file_size").
			Where("id", link.UserFileID).Scan(&file); err != nil {
			return nil, gerror.Wrap(err, "query procurement requirement file metadata")
		}
		url := ""
		if client != nil && file.ObjectKey != "" {
			url, _ = client.SignedURLFor(file.Endpoint, file.Bucket, file.ObjectKey, 30*time.Minute)
		}
		items = append(items, procurementRequirementModel.FileItem{ID: link.ID, Endpoint: file.Endpoint, Bucket: file.Bucket, Path: file.ObjectKey, URL: url, OriginalName: file.OriginalName, ContentType: file.ContentType, Size: file.FileSize, SortOrder: link.SortOrder, CreatedAt: link.CreatedAt})
	}
	return items, nil
}

type responseRecord struct {
	ID                uint64
	RequirementID     uint64
	SupplierUserID    uint64
	SupplierCompany   string
	Message           string
	AvailableQuantity *float64 `orm:"available_quantity"`
	LeadTimeDays      *uint    `orm:"lead_time_days"`
	Status            uint8
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (s *sProcurementRequirements) loadResponses(ctx context.Context, requirementID, userID uint64, role uint8) ([]procurementRequirementModel.ResponseItem, error) {
	query := dao.ProcurementRequirementResponses.Ctx(ctx).
		Where("requirement_id", requirementID)
	if role == 2 {
		query = query.Where("supplier_user_id", userID)
	}
	var records []responseRecord
	if err := query.Fields("id", "requirement_id", "supplier_user_id", "message", "available_quantity", "lead_time_days", "status", "created_at", "updated_at").OrderDesc("created_at").OrderDesc("id").Scan(&records); err != nil {
		return nil, gerror.Wrap(err, "query procurement requirement responses")
	}
	items := make([]procurementRequirementModel.ResponseItem, 0, len(records))
	for _, record := range records {
		var supplier struct {
			ID          uint64
			CompanyName string
			Status      uint8
			Role        uint8
		}
		if err := dao.Users.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("id", "company_name", "status", "role").Where("id", record.SupplierUserID).Scan(&supplier); err != nil {
			return nil, gerror.Wrap(err, "query procurement requirement response supplier")
		}
		if supplier.ID == 0 || supplier.Status != 1 || supplier.Role != 2 {
			continue
		}
		record.SupplierCompany = supplier.CompanyName
		items = append(items, responseItem(record))
	}
	return items, nil
}

func (s *sProcurementRequirements) loadResponse(ctx context.Context, id uint64) (procurementRequirementModel.ResponseItem, error) {
	var record responseRecord
	if err := dao.ProcurementRequirementResponses.Ctx(ctx).Cache(gdb.CacheOption{}).
		Where("id", id).
		Fields("id", "requirement_id", "supplier_user_id", "message", "available_quantity", "lead_time_days", "status", "created_at", "updated_at").Scan(&record); err != nil {
		return procurementRequirementModel.ResponseItem{}, gerror.Wrap(err, "query procurement requirement response")
	}
	if record.ID == 0 {
		return procurementRequirementModel.ResponseItem{}, consts.ErrRequirementResponseNotFound
	}
	var supplier struct {
		ID          uint64
		CompanyName string
		Status      uint8
		Role        uint8
	}
	if err := dao.Users.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("id", "company_name", "status", "role").Where("id", record.SupplierUserID).Scan(&supplier); err != nil {
		return procurementRequirementModel.ResponseItem{}, gerror.Wrap(err, "query procurement requirement response supplier")
	}
	if supplier.ID == 0 || supplier.Status != 1 || supplier.Role != 2 {
		return procurementRequirementModel.ResponseItem{}, consts.ErrRequirementResponseNotFound
	}
	record.SupplierCompany = supplier.CompanyName
	return responseItem(record), nil
}

func responseItem(record responseRecord) procurementRequirementModel.ResponseItem {
	return procurementRequirementModel.ResponseItem{ID: record.ID, RequirementID: record.RequirementID, SupplierUserID: record.SupplierUserID, SupplierCompany: record.SupplierCompany, Message: record.Message, AvailableQuantity: record.AvailableQuantity, LeadTimeDays: record.LeadTimeDays, Status: record.Status, CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt}
}

func (s *sProcurementRequirements) loadPendingFiles(ctx context.Context, userID uint64, paths []string) ([]fileRecord, error) {
	if len(paths) > maxRequirementAttachments {
		return nil, &consts.InvalidParameterError{Message: "最多上传10个需求附件"}
	}
	files := make([]fileRecord, 0, len(paths))
	seen := make(map[string]struct{}, len(paths))
	for index, rawPath := range paths {
		path := strings.TrimSpace(rawPath)
		if path == "" {
			return nil, &consts.InvalidParameterError{Message: "requirement attachment path is invalid"}
		}
		if _, exists := seen[path]; exists {
			return nil, &consts.InvalidParameterError{Message: "requirement attachment cannot be duplicated"}
		}
		seen[path] = struct{}{}
		var file fileRecord
		if err := dao.UserFiles.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("id", "endpoint", "bucket", "object_key", "original_name", "content_type", "file_size").Where(g.Map{"user_id": userID, "purpose": requirementAttachmentPurpose, "object_key": path}).Scan(&file); err != nil {
			return nil, gerror.Wrap(err, "query procurement requirement attachment metadata")
		}
		if file.ObjectKey == "" {
			return nil, &consts.InvalidParameterError{Message: "requirement attachment must be uploaded before submitting"}
		}
		file.SortOrder = uint(index)
		files = append(files, file)
	}
	return files, nil
}

func categoryName(categories []supplierProductModel.CategoryItem, id uint64) (string, error) {
	for _, category := range categories {
		if category.ID == id {
			return category.Name, nil
		}
	}
	return "", consts.ErrCategoryNotFound
}

func stockUnitName(units []supplierProductModel.StockUnitItem, id uint64) (string, error) {
	for _, unit := range units {
		if unit.ID == id {
			return unit.Name, nil
		}
	}
	return "", consts.ErrStockUnitNotFound
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func isDuplicateKeyError(err error) bool {
	var mysqlErr *mysqlDriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

func extensionForImageContentType(contentType string) string {
	switch contentType {
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	default:
		return ".jpg"
	}
}
