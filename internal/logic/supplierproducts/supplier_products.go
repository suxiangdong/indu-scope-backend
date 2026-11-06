package supplierproducts

import (
	"context"
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	"github.com/suxiangdong/indu-scope-backend/internal/dao"
	"github.com/suxiangdong/indu-scope-backend/internal/model/do"
	"github.com/suxiangdong/indu-scope-backend/internal/model/entity"
	supplierProductModel "github.com/suxiangdong/indu-scope-backend/internal/model/supplierproduct"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

type sSupplierProducts struct{}

var _ service.ISupplierProducts = (*sSupplierProducts)(nil)

func New() service.ISupplierProducts {
	return &sSupplierProducts{}
}

func init() {
	service.RegisterSupplierProducts(New())
}

func (s *sSupplierProducts) FormOptions(ctx context.Context, categoryID *uint64) (supplierProductModel.FormOptions, error) {
	categories, err := s.listCategories(ctx)
	if err != nil {
		return supplierProductModel.FormOptions{}, err
	}
	result := supplierProductModel.FormOptions{Categories: categories, Attributes: []supplierProductModel.AttributeItem{}}
	if categoryID == nil || *categoryID == 0 {
		return result, nil
	}
	if err := s.ensureLeafCategory(ctx, *categoryID); err != nil {
		return supplierProductModel.FormOptions{}, err
	}

	bindings, err := service.AdminAttributes().ListCategoryBindings(ctx, categoryID, nil)
	if err != nil {
		return supplierProductModel.FormOptions{}, gerror.Wrap(err, "list category attributes")
	}
	for _, binding := range bindings.Items {
		if binding.GroupID != nil {
			activeGroupCount, err := dao.AttributeGroups.Ctx(ctx).Where(do.AttributeGroups{Id: *binding.GroupID, Status: 1}).Count()
			if err != nil {
				return supplierProductModel.FormOptions{}, gerror.Wrap(err, "check attribute group")
			}
			if activeGroupCount == 0 {
				continue
			}
		}
		attribute, err := s.loadActiveAttribute(ctx, binding.AttributeID)
		if err != nil {
			if err == nilAttribute {
				continue
			}
			return supplierProductModel.FormOptions{}, err
		}
		options, err := s.listAttributeOptions(ctx, attribute.Id)
		if err != nil {
			return supplierProductModel.FormOptions{}, err
		}
		result.Attributes = append(result.Attributes, supplierProductModel.AttributeItem{
			ID: attribute.Id, Code: attribute.Code, Name: attribute.Name,
			ValueType: attribute.ValueType, InputType: attribute.InputType,
			Unit: attribute.Unit, Placeholder: attribute.Placeholder,
			HelpText: attribute.HelpText, IsRequired: binding.IsRequired,
			SortOrder: binding.SortOrder, GroupDisplayName: firstNonEmpty(binding.GroupDisplayName, binding.GroupName),
			Options: options,
		})
	}
	return result, nil
}

func (s *sSupplierProducts) Create(ctx context.Context, input supplierProductModel.CreateInput) (supplierProductModel.ProductItem, error) {
	if input.SupplierUserID == 0 {
		return supplierProductModel.ProductItem{}, consts.ErrSupplierRequired
	}
	input.Name = strings.TrimSpace(input.Name)
	input.DescriptionHTML = strings.TrimSpace(input.DescriptionHTML)
	input.ConditionCode = strings.TrimSpace(input.ConditionCode)
	input.StockUnitCode = strings.TrimSpace(input.StockUnitCode)
	if input.CategoryID == 0 {
		return supplierProductModel.ProductItem{}, &consts.InvalidParameterError{Message: "category_id is required"}
	}
	if input.Name == "" || len([]rune(input.Name)) > 255 {
		return supplierProductModel.ProductItem{}, &consts.InvalidParameterError{Message: "name is required and must be no longer than 255 characters"}
	}
	if input.StockQuantity < 0 || math.IsNaN(input.StockQuantity) || math.IsInf(input.StockQuantity, 0) {
		return supplierProductModel.ProductItem{}, &consts.InvalidParameterError{Message: "stock_quantity must be greater than or equal to 0"}
	}
	if input.StockUnitCode == "" {
		input.StockUnitCode = "piece"
	}
	if len([]rune(input.StockUnitCode)) > 32 {
		return supplierProductModel.ProductItem{}, &consts.InvalidParameterError{Message: "stock_unit_code must be no longer than 32 characters"}
	}
	if input.ConditionCode == "" {
		input.ConditionCode = "new"
	}
	if len([]rune(input.ConditionCode)) > 32 {
		return supplierProductModel.ProductItem{}, &consts.InvalidParameterError{Message: "condition_code must be no longer than 32 characters"}
	}
	if err := s.ensureLeafCategory(ctx, input.CategoryID); err != nil {
		return supplierProductModel.ProductItem{}, err
	}

	form, err := s.FormOptions(ctx, &input.CategoryID)
	if err != nil {
		return supplierProductModel.ProductItem{}, err
	}
	definitions := make(map[uint64]supplierProductModel.AttributeItem, len(form.Attributes))
	for _, definition := range form.Attributes {
		definitions[definition.ID] = definition
	}
	values := make(map[uint64]supplierProductModel.AttributeValueInput, len(input.Attributes))
	for _, value := range input.Attributes {
		if value.AttributeID == 0 {
			return supplierProductModel.ProductItem{}, &consts.InvalidParameterError{Message: "attribute_id is required"}
		}
		if _, exists := definitions[value.AttributeID]; !exists {
			return supplierProductModel.ProductItem{}, &consts.InvalidParameterError{Message: "attribute is not assigned to the selected category"}
		}
		if _, exists := values[value.AttributeID]; exists {
			return supplierProductModel.ProductItem{}, &consts.InvalidParameterError{Message: "an attribute can only be submitted once"}
		}
		value.Value = strings.TrimSpace(value.Value)
		values[value.AttributeID] = value
	}
	for _, definition := range form.Attributes {
		value, exists := values[definition.ID]
		if !exists || (value.Value == "" && value.OptionID == 0 && len(value.OptionIDs) == 0) {
			if definition.IsRequired == 1 {
				return supplierProductModel.ProductItem{}, &consts.InvalidParameterError{Message: "请填写必填属性：" + definition.Name}
			}
			continue
		}
		if err := validateAttributeValue(definition, value); err != nil {
			return supplierProductModel.ProductItem{}, err
		}
	}

	var product supplierProductModel.ProductItem
	err = dao.Products.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		id, err := tx.Model(dao.Products.Table()).Data(g.Map{
			"supplier_user_id": input.SupplierUserID,
			"category_id":      input.CategoryID,
			"name":             input.Name,
			"description_html": input.DescriptionHTML,
			"condition_code":   input.ConditionCode,
			"stock_quantity":   input.StockQuantity,
			"stock_unit_code":  input.StockUnitCode,
			"stock_updated_at": time.Now(),
			"stock_updated_by": input.SupplierUserID,
			"status":           0,
		}).InsertAndGetId()
		if err != nil {
			return gerror.Wrap(err, "create product")
		}
		product = supplierProductModel.ProductItem{
			ID: uint64(id), SupplierUserID: input.SupplierUserID, CategoryID: input.CategoryID,
			Name: input.Name, DescriptionHTML: input.DescriptionHTML, ConditionCode: input.ConditionCode,
			StockQuantity: input.StockQuantity, StockUnitCode: input.StockUnitCode, Status: 0,
		}
		for _, value := range values {
			definition := definitions[value.AttributeID]
			if value.Value == "" && value.OptionID == 0 && len(value.OptionIDs) == 0 {
				continue
			}
			data, optionIDs := attributeValueData(definition, value)
			data["product_id"] = id
			data["attribute_id"] = value.AttributeID
			valueID, err := tx.Model(dao.ProductAttributeValues.Table()).Data(data).InsertAndGetId()
			if err != nil {
				return gerror.Wrap(err, "create product attribute value")
			}
			for _, optionID := range optionIDs {
				if _, err := tx.Model(dao.ProductAttributeValueOptions.Table()).Data(g.Map{
					"value_id": valueID, "option_id": optionID,
				}).Insert(); err != nil {
					return gerror.Wrap(err, "create product attribute option value")
				}
			}
		}
		return nil
	})
	if err != nil {
		return supplierProductModel.ProductItem{}, err
	}
	return product, nil
}

var nilAttribute = gerror.New("attribute is inactive or missing")

func (s *sSupplierProducts) listCategories(ctx context.Context) ([]supplierProductModel.CategoryItem, error) {
	cols := dao.Categories.Columns()
	var records []*entity.Categories
	if err := dao.Categories.Ctx(ctx).Fields(cols.Id, cols.ParentId, cols.Name, cols.SortOrder).
		Where(do.Categories{Status: 1}).OrderAsc(cols.ParentId).OrderAsc(cols.SortOrder).OrderAsc(cols.Id).Scan(&records); err != nil {
		return nil, gerror.Wrap(err, "list active categories")
	}
	items := make([]supplierProductModel.CategoryItem, 0, len(records))
	for _, record := range records {
		if record != nil {
			items = append(items, supplierProductModel.CategoryItem{ID: record.Id, ParentID: record.ParentId, Name: record.Name, SortOrder: record.SortOrder})
		}
	}
	return items, nil
}

func (s *sSupplierProducts) ensureLeafCategory(ctx context.Context, categoryID uint64) error {
	count, err := dao.Categories.Ctx(ctx).Where(do.Categories{Id: categoryID, Status: 1}).Count()
	if err != nil {
		return gerror.Wrap(err, "check category")
	}
	if count == 0 {
		return consts.ErrCategoryNotFound
	}
	children, err := dao.Categories.Ctx(ctx).Where(do.Categories{ParentId: categoryID, Status: 1}).Count()
	if err != nil {
		return gerror.Wrap(err, "check category children")
	}
	if children > 0 {
		return &consts.InvalidParameterError{Message: "category_id must be a leaf category"}
	}
	return nil
}

func (s *sSupplierProducts) loadActiveAttribute(ctx context.Context, id uint64) (*entity.Attributes, error) {
	cols := dao.Attributes.Columns()
	var record *entity.Attributes
	if err := dao.Attributes.Ctx(ctx).Fields(cols.Id, cols.Code, cols.Name, cols.ValueType, cols.InputType, cols.Unit, cols.Placeholder, cols.HelpText, cols.Status).
		Where(do.Attributes{Id: id, Status: 1}).Scan(&record); err != nil {
		return nil, gerror.Wrap(err, "query attribute")
	}
	if record == nil {
		return nil, nilAttribute
	}
	return record, nil
}

func (s *sSupplierProducts) listAttributeOptions(ctx context.Context, attributeID uint64) ([]supplierProductModel.AttributeOptionItem, error) {
	cols := dao.AttributeOptions.Columns()
	var records []*entity.AttributeOptions
	if err := dao.AttributeOptions.Ctx(ctx).Fields(cols.Id, cols.Code, cols.Name).
		Where(do.AttributeOptions{AttributeId: attributeID, Status: 1}).OrderAsc(cols.SortOrder).OrderAsc(cols.Id).Scan(&records); err != nil {
		return nil, gerror.Wrap(err, "list attribute options")
	}
	items := make([]supplierProductModel.AttributeOptionItem, 0, len(records))
	for _, record := range records {
		if record != nil {
			items = append(items, supplierProductModel.AttributeOptionItem{ID: record.Id, Code: record.Code, Name: record.Name})
		}
	}
	return items, nil
}

func validateAttributeValue(definition supplierProductModel.AttributeItem, value supplierProductModel.AttributeValueInput) error {
	optionInput := len(definition.Options) > 0 || strings.Contains(strings.ToLower(definition.InputType), "select") || strings.Contains(strings.ToLower(definition.InputType), "radio") || strings.Contains(strings.ToLower(definition.InputType), "checkbox")
	if optionInput {
		ids := normalizedOptionIDs(value)
		if len(ids) == 0 {
			return &consts.InvalidParameterError{Message: "请选择属性：" + definition.Name}
		}
		allowed := make(map[uint64]bool, len(definition.Options))
		for _, option := range definition.Options {
			allowed[option.ID] = true
		}
		for _, id := range ids {
			if !allowed[id] {
				return &consts.InvalidParameterError{Message: "属性选项无效：" + definition.Name}
			}
		}
		return nil
	}
	if value.Value == "" {
		return nil
	}
	switch definition.ValueType {
	case 2:
		number, err := strconv.ParseFloat(value.Value, 64)
		if err != nil || math.Trunc(number) != number {
			return &consts.InvalidParameterError{Message: "属性必须是整数：" + definition.Name}
		}
	case 3:
		if number, err := strconv.ParseFloat(value.Value, 64); err != nil || math.IsNaN(number) || math.IsInf(number, 0) {
			return &consts.InvalidParameterError{Message: "属性必须是数字：" + definition.Name}
		}
	case 4:
		if _, err := parseBool(value.Value); err != nil {
			return &consts.InvalidParameterError{Message: "属性必须是布尔值：" + definition.Name}
		}
	case 6:
		if !json.Valid([]byte(value.Value)) {
			return &consts.InvalidParameterError{Message: "属性必须是有效的 JSON：" + definition.Name}
		}
	}
	return nil
}

func attributeValueData(definition supplierProductModel.AttributeItem, value supplierProductModel.AttributeValueInput) (g.Map, []uint64) {
	data := g.Map{"value_number": nil, "value_text": nil, "value_bool": nil, "value_option_id": nil, "value_json": nil}
	ids := normalizedOptionIDs(value)
	if len(ids) > 0 {
		data["value_option_id"] = ids[0]
		return data, ids
	}
	switch definition.ValueType {
	case 2, 3:
		if number, err := strconv.ParseFloat(value.Value, 64); err == nil {
			data["value_number"] = number
		}
	case 4:
		if boolValue, err := parseBool(value.Value); err == nil {
			if boolValue {
				data["value_bool"] = 1
			} else {
				data["value_bool"] = 0
			}
		}
	case 6:
		data["value_json"] = value.Value
	default:
		data["value_text"] = value.Value
	}
	return data, nil
}

func normalizedOptionIDs(value supplierProductModel.AttributeValueInput) []uint64 {
	ids := make([]uint64, 0, len(value.OptionIDs)+1)
	if value.OptionID != 0 {
		ids = append(ids, value.OptionID)
	}
	ids = append(ids, value.OptionIDs...)
	seen := make(map[uint64]struct{}, len(ids))
	result := ids[:0]
	for _, id := range ids {
		if id != 0 {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			result = append(result, id)
		}
	}
	return result
}

func parseBool(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "是":
		return true, nil
	case "0", "false", "no", "否":
		return false, nil
	default:
		return false, gerror.New("invalid boolean")
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
