package adminattributes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"

	mysqlDriver "github.com/go-sql-driver/mysql"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	"github.com/suxiangdong/indu-scope-backend/internal/dao"
	attributeModel "github.com/suxiangdong/indu-scope-backend/internal/model/attribute"
	"github.com/suxiangdong/indu-scope-backend/internal/model/do"
	"github.com/suxiangdong/indu-scope-backend/internal/model/entity"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

type sAdminAttributes struct{}

var _ service.IAdminAttributes = (*sAdminAttributes)(nil)

func New() service.IAdminAttributes {
	return &sAdminAttributes{}
}

func init() {
	service.RegisterAdminAttributes(New())
}

func (s *sAdminAttributes) List(ctx context.Context) (attributeModel.AttributeListResult, error) {
	cols := dao.Attributes.Columns()
	var records []*entity.Attributes
	if err := dao.Attributes.Ctx(ctx).Fields(
		cols.Id, cols.Code, cols.Name, cols.ValueType, cols.InputType,
		cols.Unit, cols.Placeholder, cols.ValidationRules, cols.HelpText,
		cols.Status, cols.CreatedAt, cols.UpdatedAt,
	).OrderAsc(cols.Code).OrderAsc(cols.Id).Scan(&records); err != nil {
		return attributeModel.AttributeListResult{}, gerror.Wrap(err, "list attributes")
	}

	items := make([]attributeModel.AttributeItem, 0, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		optionCount, err := dao.AttributeOptions.Ctx(ctx).
			Where(do.AttributeOptions{AttributeId: record.Id}).Count()
		if err != nil {
			return attributeModel.AttributeListResult{}, gerror.Wrap(err, "count attribute options")
		}
		item := toAttributeItem(record)
		item.OptionCount = uint(optionCount)
		groupCount, err := dao.AttributeGroupAttributes.Ctx(ctx).
			Where(do.AttributeGroupAttributes{AttributeId: record.Id}).Count()
		if err != nil {
			return attributeModel.AttributeListResult{}, gerror.Wrap(err, "count attribute groups")
		}
		item.GroupCount = uint(groupCount)
		item.GroupIDs, err = s.listAttributeGroupIDs(ctx, record.Id)
		if err != nil {
			return attributeModel.AttributeListResult{}, err
		}
		items = append(items, item)
	}
	return attributeModel.AttributeListResult{Items: items}, nil
}

func (s *sAdminAttributes) Create(ctx context.Context, input attributeModel.AttributeInput) (attributeModel.AttributeItem, error) {
	input = normalizeAttributeInput(input)
	if err := validateAttributeInput(input); err != nil {
		return attributeModel.AttributeItem{}, err
	}
	if err := s.validateGroupIDs(ctx, input.GroupIDs); err != nil {
		return attributeModel.AttributeItem{}, err
	}
	id, err := dao.Attributes.Ctx(ctx).Data(attributeData(input)).InsertAndGetId()
	if err != nil {
		if isDuplicateKey(err) {
			return attributeModel.AttributeItem{}, consts.ErrAttributeAlreadyExists
		}
		return attributeModel.AttributeItem{}, gerror.Wrap(err, "create attribute")
	}
	attributeID := uint64(id)
	if err := s.syncAttributeGroups(ctx, attributeID, input.GroupIDs); err != nil {
		_, _ = dao.Attributes.Ctx(ctx).Where(do.Attributes{Id: attributeID}).Delete()
		return attributeModel.AttributeItem{}, err
	}
	return s.loadAttribute(ctx, attributeID)
}

func (s *sAdminAttributes) Update(ctx context.Context, id uint64, input attributeModel.AttributeInput) (attributeModel.AttributeItem, error) {
	if id == 0 {
		return attributeModel.AttributeItem{}, &consts.InvalidParameterError{Message: "id is required"}
	}
	input = normalizeAttributeInput(input)
	if err := validateAttributeInput(input); err != nil {
		return attributeModel.AttributeItem{}, err
	}
	if input.GroupIDs != nil {
		if err := s.validateGroupIDs(ctx, input.GroupIDs); err != nil {
			return attributeModel.AttributeItem{}, err
		}
	}
	if count, err := dao.Attributes.Ctx(ctx).Where(do.Attributes{Id: id}).Count(); err != nil {
		return attributeModel.AttributeItem{}, gerror.Wrap(err, "query attribute")
	} else if count == 0 {
		return attributeModel.AttributeItem{}, consts.ErrAttributeNotFound
	}
	if _, err := dao.Attributes.Ctx(ctx).Where(do.Attributes{Id: id}).Data(attributeData(input)).Update(); err != nil {
		if isDuplicateKey(err) {
			return attributeModel.AttributeItem{}, consts.ErrAttributeAlreadyExists
		}
		return attributeModel.AttributeItem{}, gerror.Wrap(err, "update attribute")
	}
	if input.GroupIDs != nil {
		if err := s.syncAttributeGroups(ctx, id, input.GroupIDs); err != nil {
			return attributeModel.AttributeItem{}, err
		}
	}
	return s.loadAttribute(ctx, id)
}

func (s *sAdminAttributes) Delete(ctx context.Context, id uint64) error {
	if id == 0 {
		return &consts.InvalidParameterError{Message: "id is required"}
	}
	if count, err := dao.Attributes.Ctx(ctx).Where(do.Attributes{Id: id}).Count(); err != nil {
		return gerror.Wrap(err, "query attribute")
	} else if count == 0 {
		return consts.ErrAttributeNotFound
	}
	if count, err := dao.CategoryAttributes.Ctx(ctx).Where(do.CategoryAttributes{AttributeId: id}).Count(); err != nil {
		return gerror.Wrap(err, "check attribute categories")
	} else if count > 0 {
		return consts.ErrAttributeHasCategories
	}
	if count, err := dao.AttributeGroupAttributes.Ctx(ctx).Where(do.AttributeGroupAttributes{AttributeId: id}).Count(); err != nil {
		return gerror.Wrap(err, "check attribute groups")
	} else if count > 0 {
		return consts.ErrAttributeHasCategories
	}
	if count, err := dao.ProductAttributeValues.Ctx(ctx).Where(do.ProductAttributeValues{AttributeId: id}).Count(); err != nil {
		return gerror.Wrap(err, "check attribute product values")
	} else if count > 0 {
		return consts.ErrAttributeHasProducts
	}
	result, err := dao.Attributes.Ctx(ctx).Where(do.Attributes{Id: id}).Delete()
	if err != nil {
		return gerror.Wrap(err, "delete attribute")
	}
	if affected, err := result.RowsAffected(); err != nil {
		return gerror.Wrap(err, "check deleted attribute")
	} else if affected == 0 {
		return consts.ErrAttributeNotFound
	}
	return nil
}

func (s *sAdminAttributes) ListOptions(ctx context.Context, attributeID uint64) (attributeModel.AttributeOptionListResult, error) {
	if err := s.ensureAttributeExists(ctx, attributeID); err != nil {
		return attributeModel.AttributeOptionListResult{}, err
	}
	cols := dao.AttributeOptions.Columns()
	var records []*entity.AttributeOptions
	if err := dao.AttributeOptions.Ctx(ctx).Fields(
		cols.Id, cols.AttributeId, cols.Code, cols.Name, cols.SortOrder,
		cols.Status, cols.CreatedAt, cols.UpdatedAt,
	).Where(do.AttributeOptions{AttributeId: attributeID}).
		OrderAsc(cols.SortOrder).OrderAsc(cols.Id).Scan(&records); err != nil {
		return attributeModel.AttributeOptionListResult{}, gerror.Wrap(err, "list attribute options")
	}
	items := make([]attributeModel.AttributeOptionItem, 0, len(records))
	for _, record := range records {
		if record != nil {
			items = append(items, toAttributeOptionItem(record))
		}
	}
	return attributeModel.AttributeOptionListResult{Items: items}, nil
}

func (s *sAdminAttributes) CreateOption(ctx context.Context, attributeID uint64, input attributeModel.AttributeOptionInput) (attributeModel.AttributeOptionItem, error) {
	if err := s.ensureAttributeExists(ctx, attributeID); err != nil {
		return attributeModel.AttributeOptionItem{}, err
	}
	input = normalizeOptionInput(input)
	if err := validateOptionInput(input); err != nil {
		return attributeModel.AttributeOptionItem{}, err
	}
	id, err := dao.AttributeOptions.Ctx(ctx).Data(g.Map{
		"attribute_id": attributeID,
		"code":         input.Code,
		"name":         input.Name,
		"sort_order":   input.SortOrder,
		"status":       input.Status,
	}).InsertAndGetId()
	if err != nil {
		if isDuplicateKey(err) {
			return attributeModel.AttributeOptionItem{}, consts.ErrAttributeOptionExists
		}
		return attributeModel.AttributeOptionItem{}, gerror.Wrap(err, "create attribute option")
	}
	return s.loadOption(ctx, uint64(id))
}

func (s *sAdminAttributes) UpdateOption(ctx context.Context, id uint64, input attributeModel.AttributeOptionInput) (attributeModel.AttributeOptionItem, error) {
	if id == 0 {
		return attributeModel.AttributeOptionItem{}, &consts.InvalidParameterError{Message: "id is required"}
	}
	input = normalizeOptionInput(input)
	if err := validateOptionInput(input); err != nil {
		return attributeModel.AttributeOptionItem{}, err
	}
	record, err := s.loadOptionEntity(ctx, id)
	if err != nil {
		return attributeModel.AttributeOptionItem{}, err
	}
	optionCols := dao.AttributeOptions.Columns()
	if count, err := dao.AttributeOptions.Ctx(ctx).
		Where(do.AttributeOptions{AttributeId: record.AttributeId, Code: input.Code}).
		Wheref("%s <> ?", optionCols.Id, id).Count(); err != nil {
		return attributeModel.AttributeOptionItem{}, gerror.Wrap(err, "check attribute option code")
	} else if count > 0 {
		return attributeModel.AttributeOptionItem{}, consts.ErrAttributeOptionExists
	}
	if _, err := dao.AttributeOptions.Ctx(ctx).Where(do.AttributeOptions{Id: id}).Data(g.Map{
		"code": input.Code, "name": input.Name,
		"sort_order": input.SortOrder, "status": input.Status,
	}).Update(); err != nil {
		if isDuplicateKey(err) {
			return attributeModel.AttributeOptionItem{}, consts.ErrAttributeOptionExists
		}
		return attributeModel.AttributeOptionItem{}, gerror.Wrap(err, "update attribute option")
	}
	return s.loadOption(ctx, id)
}

func (s *sAdminAttributes) DeleteOption(ctx context.Context, id uint64) error {
	if id == 0 {
		return &consts.InvalidParameterError{Message: "id is required"}
	}
	if _, err := s.loadOptionEntity(ctx, id); err != nil {
		return err
	}
	if count, err := dao.ProductAttributeValues.Ctx(ctx).
		Where(do.ProductAttributeValues{ValueOptionId: id}).Count(); err != nil {
		return gerror.Wrap(err, "check option product values")
	} else if count > 0 {
		return consts.ErrAttributeOptionInUse
	}
	if count, err := dao.ProductAttributeValueOptions.Ctx(ctx).
		Where(do.ProductAttributeValueOptions{OptionId: id}).Count(); err != nil {
		return gerror.Wrap(err, "check option product value selections")
	} else if count > 0 {
		return consts.ErrAttributeOptionInUse
	}
	result, err := dao.AttributeOptions.Ctx(ctx).Where(do.AttributeOptions{Id: id}).Delete()
	if err != nil {
		return gerror.Wrap(err, "delete attribute option")
	}
	if affected, err := result.RowsAffected(); err != nil {
		return gerror.Wrap(err, "check deleted attribute option")
	} else if affected == 0 {
		return consts.ErrAttributeOptionNotFound
	}
	return nil
}

func (s *sAdminAttributes) ListCategoryBindings(ctx context.Context, categoryID, attributeID *uint64) (attributeModel.CategoryAttributeListResult, error) {
	if categoryID != nil && *categoryID == 0 {
		return attributeModel.CategoryAttributeListResult{}, &consts.InvalidParameterError{Message: "category_id is required"}
	}
	if attributeID != nil && *attributeID == 0 {
		return attributeModel.CategoryAttributeListResult{}, &consts.InvalidParameterError{Message: "attribute_id is required"}
	}
	items, err := s.listEffectiveCategoryBindings(ctx, categoryID)
	if err != nil {
		return attributeModel.CategoryAttributeListResult{}, err
	}
	if attributeID == nil {
		return attributeModel.CategoryAttributeListResult{Items: items}, nil
	}
	filtered := make([]attributeModel.CategoryAttributeItem, 0, len(items))
	for _, item := range items {
		if item.AttributeID == *attributeID {
			filtered = append(filtered, item)
		}
	}
	return attributeModel.CategoryAttributeListResult{Items: filtered}, nil
}

func (s *sAdminAttributes) SaveCategoryBinding(ctx context.Context, input attributeModel.CategoryAttributeInput) (attributeModel.CategoryAttributeItem, error) {
	if input.CategoryID == 0 || input.AttributeID == 0 {
		return attributeModel.CategoryAttributeItem{}, &consts.InvalidParameterError{Message: "category_id and attribute_id are required"}
	}
	if err := validateFlag(input.IsRequired, "is_required"); err != nil {
		return attributeModel.CategoryAttributeItem{}, err
	}
	if err := validateFlag(input.IsFilterable, "is_filterable"); err != nil {
		return attributeModel.CategoryAttributeItem{}, err
	}
	if err := validateFlag(input.IsSearchable, "is_searchable"); err != nil {
		return attributeModel.CategoryAttributeItem{}, err
	}
	if err := s.ensureCategoryExists(ctx, input.CategoryID); err != nil {
		return attributeModel.CategoryAttributeItem{}, err
	}
	if count, err := dao.Categories.Ctx(ctx).
		Where(do.Categories{ParentId: input.CategoryID}).Count(); err != nil {
		return attributeModel.CategoryAttributeItem{}, gerror.Wrap(err, "check category children")
	} else if count > 0 {
		return attributeModel.CategoryAttributeItem{}, consts.ErrCategoryHasChildren
	}
	if err := s.ensureAttributeExists(ctx, input.AttributeID); err != nil {
		return attributeModel.CategoryAttributeItem{}, err
	}
	groupID, err := s.resolveCategoryGroup(ctx, input.CategoryID, input.AttributeID, input.GroupID, input.GroupName)
	if err != nil {
		return attributeModel.CategoryAttributeItem{}, err
	}
	group, err := s.loadAttributeGroupEntity(ctx, groupID)
	if err != nil {
		return attributeModel.CategoryAttributeItem{}, err
	}
	data := g.Map{
		"category_id": input.CategoryID, "attribute_id": input.AttributeID,
		"group_id":    groupID,
		"is_required": input.IsRequired, "is_filterable": input.IsFilterable,
		"is_searchable": input.IsSearchable, "sort_order": input.SortOrder,
		"group_name": group.Name,
	}
	query := dao.CategoryAttributes.Ctx(ctx).Where(do.CategoryAttributes{
		CategoryId: input.CategoryID, AttributeId: input.AttributeID,
	})
	var record *entity.CategoryAttributes
	if err := query.Scan(&record); err != nil {
		return attributeModel.CategoryAttributeItem{}, gerror.Wrap(err, "query category attribute")
	}
	if record == nil {
		if _, err := dao.CategoryAttributes.Ctx(ctx).Data(data).Insert(); err != nil {
			if isDuplicateKey(err) {
				return attributeModel.CategoryAttributeItem{}, consts.ErrCategoryAttributeExists
			}
			return attributeModel.CategoryAttributeItem{}, gerror.Wrap(err, "create category attribute")
		}
	} else if _, err := query.Data(data).Update(); err != nil {
		return attributeModel.CategoryAttributeItem{}, gerror.Wrap(err, "update category attribute")
	}
	return s.loadCategoryBindingByIDs(ctx, input.CategoryID, input.AttributeID)
}

func (s *sAdminAttributes) DeleteCategoryBinding(ctx context.Context, categoryID, attributeID uint64) error {
	if categoryID == 0 || attributeID == 0 {
		return &consts.InvalidParameterError{Message: "category_id and attribute_id are required"}
	}
	result, err := dao.CategoryAttributes.Ctx(ctx).Where(do.CategoryAttributes{
		CategoryId: categoryID, AttributeId: attributeID,
	}).Delete()
	if err != nil {
		return gerror.Wrap(err, "delete category attribute")
	}
	if affected, err := result.RowsAffected(); err != nil {
		return gerror.Wrap(err, "check deleted category attribute")
	} else if affected == 0 {
		return consts.ErrCategoryAttributeNotFound
	}
	return nil
}

func (s *sAdminAttributes) listEffectiveCategoryBindings(ctx context.Context, categoryID *uint64) ([]attributeModel.CategoryAttributeItem, error) {
	bindingCols := dao.CategoryAttributeGroupBindings.Columns()
	bindingQuery := dao.CategoryAttributeGroupBindings.Ctx(ctx).Fields(
		bindingCols.CategoryId, bindingCols.GroupId, bindingCols.SortOrder, bindingCols.Status,
		bindingCols.CreatedAt, bindingCols.UpdatedAt,
	).Where(do.CategoryAttributeGroupBindings{Status: 1}).OrderAsc(bindingCols.CategoryId).OrderAsc(bindingCols.SortOrder).OrderAsc(bindingCols.GroupId)
	if categoryID != nil {
		bindingQuery = bindingQuery.Where(do.CategoryAttributeGroupBindings{CategoryId: *categoryID})
	}
	var bindings []*entity.CategoryAttributeGroupBindings
	if err := bindingQuery.Scan(&bindings); err != nil {
		return nil, gerror.Wrap(err, "list category group bindings")
	}

	type candidate struct {
		item        attributeModel.CategoryAttributeItem
		groupOrder  uint
		attrOrder   uint
		hasOverride bool
	}
	candidates := make(map[string]candidate)
	for _, binding := range bindings {
		if binding == nil {
			continue
		}
		group, err := s.loadAttributeGroupEntity(ctx, binding.GroupId)
		if err != nil {
			return nil, err
		}
		members, err := s.listGroupAttributeEntities(ctx, binding.GroupId)
		if err != nil {
			return nil, err
		}
		for _, member := range members {
			if member == nil {
				continue
			}
			item, hasOverride, err := s.buildCategoryAttributeItem(ctx, binding.CategoryId, binding, group, member)
			if err != nil {
				return nil, err
			}
			key := fmt.Sprintf("%d:%d", binding.CategoryId, member.AttributeId)
			current, exists := candidates[key]
			if !exists || (hasOverride && !current.hasOverride) ||
				(hasOverride == current.hasOverride && (binding.SortOrder < current.groupOrder ||
					(binding.SortOrder == current.groupOrder && member.SortOrder < current.attrOrder))) {
				candidates[key] = candidate{item: item, groupOrder: binding.SortOrder, attrOrder: member.SortOrder, hasOverride: hasOverride}
			}
		}
	}
	items := make([]attributeModel.CategoryAttributeItem, 0, len(candidates))
	for _, value := range candidates {
		items = append(items, value.item)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].CategoryID != items[j].CategoryID {
			return items[i].CategoryID < items[j].CategoryID
		}
		if items[i].SortOrder != items[j].SortOrder {
			return items[i].SortOrder < items[j].SortOrder
		}
		return items[i].AttributeID < items[j].AttributeID
	})
	return items, nil
}

func (s *sAdminAttributes) buildCategoryAttributeItem(ctx context.Context, categoryID uint64, binding *entity.CategoryAttributeGroupBindings, group *entity.AttributeGroups, member *entity.AttributeGroupAttributes) (attributeModel.CategoryAttributeItem, bool, error) {
	category, err := s.loadCategoryEntity(ctx, categoryID)
	if err != nil {
		return attributeModel.CategoryAttributeItem{}, false, err
	}
	attribute, err := s.loadAttributeEntity(ctx, member.AttributeId)
	if err != nil {
		return attributeModel.CategoryAttributeItem{}, false, err
	}
	override, err := s.loadCategoryOverride(ctx, categoryID, member.AttributeId)
	if err != nil {
		return attributeModel.CategoryAttributeItem{}, false, err
	}
	groupID := group.Id
	item := attributeModel.CategoryAttributeItem{
		CategoryID: category.Id, CategoryName: category.Name,
		AttributeID: attribute.Id, AttributeCode: attribute.Code, AttributeName: attribute.Name,
		GroupID: &groupID, GroupCode: group.Code, GroupName: group.Name,
		IsRequired: 0, IsFilterable: 0, IsSearchable: 0,
		SortOrder: member.SortOrder, CreatedAt: member.CreatedAt, UpdatedAt: member.UpdatedAt,
	}
	if override != nil {
		item.IsRequired = uint8(override.IsRequired)
		item.IsFilterable = uint8(override.IsFilterable)
		item.IsSearchable = uint8(override.IsSearchable)
		item.SortOrder = override.SortOrder
		item.CreatedAt = override.CreatedAt
		item.UpdatedAt = override.UpdatedAt
		return item, override.GroupId == group.Id, nil
	}
	_ = binding
	return item, false, nil
}

func (s *sAdminAttributes) loadCategoryBindingByIDs(ctx context.Context, categoryID, attributeID uint64) (attributeModel.CategoryAttributeItem, error) {
	items, err := s.listEffectiveCategoryBindings(ctx, &categoryID)
	if err != nil {
		return attributeModel.CategoryAttributeItem{}, err
	}
	for _, item := range items {
		if item.AttributeID == attributeID {
			return item, nil
		}
	}
	return attributeModel.CategoryAttributeItem{}, consts.ErrCategoryAttributeNotFound
}

func (s *sAdminAttributes) loadCategoryEntity(ctx context.Context, id uint64) (*entity.Categories, error) {
	cols := dao.Categories.Columns()
	var category *entity.Categories
	if err := dao.Categories.Ctx(ctx).Fields(cols.Id, cols.Name).Where(do.Categories{Id: id}).Scan(&category); err != nil {
		return nil, gerror.Wrap(err, "query category")
	}
	if category == nil {
		return nil, consts.ErrCategoryNotFound
	}
	return category, nil
}

func (s *sAdminAttributes) loadAttributeEntity(ctx context.Context, id uint64) (*entity.Attributes, error) {
	cols := dao.Attributes.Columns()
	var attribute *entity.Attributes
	if err := dao.Attributes.Ctx(ctx).Fields(cols.Id, cols.Code, cols.Name).Where(do.Attributes{Id: id}).Scan(&attribute); err != nil {
		return nil, gerror.Wrap(err, "query attribute")
	}
	if attribute == nil {
		return nil, consts.ErrAttributeNotFound
	}
	return attribute, nil
}

func (s *sAdminAttributes) loadCategoryOverride(ctx context.Context, categoryID, attributeID uint64) (*entity.CategoryAttributes, error) {
	cols := dao.CategoryAttributes.Columns()
	var override *entity.CategoryAttributes
	if err := dao.CategoryAttributes.Ctx(ctx).Fields(
		cols.CategoryId, cols.AttributeId, cols.GroupId, cols.IsRequired, cols.IsFilterable,
		cols.IsSearchable, cols.SortOrder, cols.GroupName, cols.CreatedAt, cols.UpdatedAt,
	).Where(do.CategoryAttributes{CategoryId: categoryID, AttributeId: attributeID}).Scan(&override); err != nil {
		return nil, gerror.Wrap(err, "query category attribute override")
	}
	return override, nil
}

func (s *sAdminAttributes) loadAttribute(ctx context.Context, id uint64) (attributeModel.AttributeItem, error) {
	cols := dao.Attributes.Columns()
	var record *entity.Attributes
	if err := dao.Attributes.Ctx(ctx).Fields(
		cols.Id, cols.Code, cols.Name, cols.ValueType, cols.InputType,
		cols.Unit, cols.Placeholder, cols.ValidationRules, cols.HelpText,
		cols.Status, cols.CreatedAt, cols.UpdatedAt,
	).Where(do.Attributes{Id: id}).Scan(&record); err != nil {
		return attributeModel.AttributeItem{}, gerror.Wrap(err, "query attribute")
	}
	if record == nil {
		return attributeModel.AttributeItem{}, consts.ErrAttributeNotFound
	}
	item := toAttributeItem(record)
	count, err := dao.AttributeOptions.Ctx(ctx).Where(do.AttributeOptions{AttributeId: id}).Count()
	if err != nil {
		return attributeModel.AttributeItem{}, gerror.Wrap(err, "count attribute options")
	}
	item.OptionCount = uint(count)
	groupCount, err := dao.AttributeGroupAttributes.Ctx(ctx).Where(do.AttributeGroupAttributes{AttributeId: id}).Count()
	if err != nil {
		return attributeModel.AttributeItem{}, gerror.Wrap(err, "count attribute groups")
	}
	item.GroupCount = uint(groupCount)
	item.GroupIDs, err = s.listAttributeGroupIDs(ctx, id)
	if err != nil {
		return attributeModel.AttributeItem{}, err
	}
	return item, nil
}

func (s *sAdminAttributes) loadOption(ctx context.Context, id uint64) (attributeModel.AttributeOptionItem, error) {
	record, err := s.loadOptionEntity(ctx, id)
	if err != nil {
		return attributeModel.AttributeOptionItem{}, err
	}
	return toAttributeOptionItem(record), nil
}

func (s *sAdminAttributes) loadOptionEntity(ctx context.Context, id uint64) (*entity.AttributeOptions, error) {
	cols := dao.AttributeOptions.Columns()
	var record *entity.AttributeOptions
	if err := dao.AttributeOptions.Ctx(ctx).Fields(
		cols.Id, cols.AttributeId, cols.Code, cols.Name, cols.SortOrder,
		cols.Status, cols.CreatedAt, cols.UpdatedAt,
	).Where(do.AttributeOptions{Id: id}).Scan(&record); err != nil {
		return nil, gerror.Wrap(err, "query attribute option")
	}
	if record == nil {
		return nil, consts.ErrAttributeOptionNotFound
	}
	return record, nil
}

func (s *sAdminAttributes) ListGroups(ctx context.Context) (attributeModel.AttributeGroupListResult, error) {
	cols := dao.AttributeGroups.Columns()
	var records []*entity.AttributeGroups
	if err := dao.AttributeGroups.Ctx(ctx).Fields(
		cols.Id, cols.Code, cols.Name, cols.SortOrder, cols.Status, cols.CreatedAt, cols.UpdatedAt,
	).OrderAsc(cols.SortOrder).OrderAsc(cols.Id).Scan(&records); err != nil {
		return attributeModel.AttributeGroupListResult{}, gerror.Wrap(err, "list attribute groups")
	}
	items := make([]attributeModel.AttributeGroupItem, 0, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		item, err := s.loadAttributeGroup(ctx, record.Id)
		if err != nil {
			return attributeModel.AttributeGroupListResult{}, err
		}
		items = append(items, item)
	}
	return attributeModel.AttributeGroupListResult{Items: items}, nil
}

func (s *sAdminAttributes) CreateGroup(ctx context.Context, input attributeModel.AttributeGroupInput) (attributeModel.AttributeGroupItem, error) {
	input = normalizeGroupInput(input)
	if err := validateGroupInput(input); err != nil {
		return attributeModel.AttributeGroupItem{}, err
	}
	id, err := dao.AttributeGroups.Ctx(ctx).Data(groupData(input)).InsertAndGetId()
	if err != nil {
		if isDuplicateKey(err) {
			return attributeModel.AttributeGroupItem{}, consts.ErrAttributeGroupExists
		}
		return attributeModel.AttributeGroupItem{}, gerror.Wrap(err, "create attribute group")
	}
	return s.loadAttributeGroup(ctx, uint64(id))
}

func (s *sAdminAttributes) UpdateGroup(ctx context.Context, id uint64, input attributeModel.AttributeGroupInput) (attributeModel.AttributeGroupItem, error) {
	if id == 0 {
		return attributeModel.AttributeGroupItem{}, &consts.InvalidParameterError{Message: "id is required"}
	}
	input = normalizeGroupInput(input)
	if err := validateGroupInput(input); err != nil {
		return attributeModel.AttributeGroupItem{}, err
	}
	if _, err := s.loadAttributeGroupEntity(ctx, id); err != nil {
		return attributeModel.AttributeGroupItem{}, err
	}
	cols := dao.AttributeGroups.Columns()
	if count, err := dao.AttributeGroups.Ctx(ctx).Where(do.AttributeGroups{Code: input.Code}).Wheref("%s <> ?", cols.Id, id).Count(); err != nil {
		return attributeModel.AttributeGroupItem{}, gerror.Wrap(err, "check attribute group code")
	} else if count > 0 {
		return attributeModel.AttributeGroupItem{}, consts.ErrAttributeGroupExists
	}
	if count, err := dao.AttributeGroups.Ctx(ctx).Where(do.AttributeGroups{Name: input.Name}).Wheref("%s <> ?", cols.Id, id).Count(); err != nil {
		return attributeModel.AttributeGroupItem{}, gerror.Wrap(err, "check attribute group name")
	} else if count > 0 {
		return attributeModel.AttributeGroupItem{}, consts.ErrAttributeGroupExists
	}
	if _, err := dao.AttributeGroups.Ctx(ctx).Where(do.AttributeGroups{Id: id}).Data(groupData(input)).Update(); err != nil {
		if isDuplicateKey(err) {
			return attributeModel.AttributeGroupItem{}, consts.ErrAttributeGroupExists
		}
		return attributeModel.AttributeGroupItem{}, gerror.Wrap(err, "update attribute group")
	}
	if _, err := dao.CategoryAttributes.Ctx(ctx).Where(do.CategoryAttributes{GroupId: id}).Data(g.Map{"group_name": input.Name}).Update(); err != nil {
		return attributeModel.AttributeGroupItem{}, gerror.Wrap(err, "sync category attribute override group name")
	}
	return s.loadAttributeGroup(ctx, id)
}

func (s *sAdminAttributes) DeleteGroup(ctx context.Context, id uint64) error {
	if id == 0 {
		return &consts.InvalidParameterError{Message: "id is required"}
	}
	if _, err := s.loadAttributeGroupEntity(ctx, id); err != nil {
		return err
	}
	if count, err := dao.CategoryAttributeGroupBindings.Ctx(ctx).Where(do.CategoryAttributeGroupBindings{GroupId: id}).Count(); err != nil {
		return gerror.Wrap(err, "check category group bindings")
	} else if count > 0 {
		return consts.ErrAttributeGroupInUse
	}
	if count, err := dao.AttributeGroupAttributes.Ctx(ctx).Where(do.AttributeGroupAttributes{GroupId: id}).Count(); err != nil {
		return gerror.Wrap(err, "check attribute group attributes")
	} else if count > 0 {
		return consts.ErrAttributeGroupInUse
	}
	if count, err := dao.CategoryAttributes.Ctx(ctx).Where(do.CategoryAttributes{GroupId: id}).Count(); err != nil {
		return gerror.Wrap(err, "check category attribute overrides")
	} else if count > 0 {
		return consts.ErrAttributeGroupInUse
	}
	result, err := dao.AttributeGroups.Ctx(ctx).Where(do.AttributeGroups{Id: id}).Delete()
	if err != nil {
		return gerror.Wrap(err, "delete attribute group")
	}
	if affected, err := result.RowsAffected(); err != nil {
		return gerror.Wrap(err, "check deleted attribute group")
	} else if affected == 0 {
		return consts.ErrAttributeGroupNotFound
	}
	return nil
}

func (s *sAdminAttributes) ListGroupAttributes(ctx context.Context, groupID uint64) (attributeModel.AttributeGroupAttributeListResult, error) {
	if err := s.ensureAttributeGroupExists(ctx, groupID); err != nil {
		return attributeModel.AttributeGroupAttributeListResult{}, err
	}
	records, err := s.listGroupAttributeEntities(ctx, groupID)
	if err != nil {
		return attributeModel.AttributeGroupAttributeListResult{}, err
	}
	items := make([]attributeModel.AttributeGroupAttributeItem, 0, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		attribute, err := s.loadAttributeEntity(ctx, record.AttributeId)
		if err != nil {
			return attributeModel.AttributeGroupAttributeListResult{}, err
		}
		fullAttribute, err := s.loadAttribute(ctx, record.AttributeId)
		if err != nil {
			return attributeModel.AttributeGroupAttributeListResult{}, err
		}
		items = append(items, attributeModel.AttributeGroupAttributeItem{
			GroupID: record.GroupId, AttributeID: attribute.Id, Code: attribute.Code, Name: attribute.Name,
			ValueType: fullAttribute.ValueType, InputType: fullAttribute.InputType, SortOrder: record.SortOrder,
			Status: fullAttribute.Status, CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
		})
	}
	return attributeModel.AttributeGroupAttributeListResult{Items: items}, nil
}

func (s *sAdminAttributes) SaveGroupAttribute(ctx context.Context, input attributeModel.AttributeGroupAttributeInput) (attributeModel.AttributeGroupAttributeItem, error) {
	if input.GroupID == 0 || input.AttributeID == 0 {
		return attributeModel.AttributeGroupAttributeItem{}, &consts.InvalidParameterError{Message: "group_id and attribute_id are required"}
	}
	if err := s.ensureAttributeGroupExists(ctx, input.GroupID); err != nil {
		return attributeModel.AttributeGroupAttributeItem{}, err
	}
	if err := s.ensureAttributeExists(ctx, input.AttributeID); err != nil {
		return attributeModel.AttributeGroupAttributeItem{}, err
	}
	_, err := dao.AttributeGroupAttributes.Ctx(ctx).Data(g.Map{
		"group_id": input.GroupID, "attribute_id": input.AttributeID, "sort_order": input.SortOrder,
	}).OnDuplicate(g.Map{"sort_order": input.SortOrder}).Insert()
	if err != nil {
		return attributeModel.AttributeGroupAttributeItem{}, gerror.Wrap(err, "save attribute group attribute")
	}
	items, err := s.ListGroupAttributes(ctx, input.GroupID)
	if err != nil {
		return attributeModel.AttributeGroupAttributeItem{}, err
	}
	for _, item := range items.Items {
		if item.AttributeID == input.AttributeID {
			return item, nil
		}
	}
	return attributeModel.AttributeGroupAttributeItem{}, consts.ErrAttributeNotFound
}

func (s *sAdminAttributes) DeleteGroupAttribute(ctx context.Context, groupID, attributeID uint64) error {
	if groupID == 0 || attributeID == 0 {
		return &consts.InvalidParameterError{Message: "group_id and attribute_id are required"}
	}
	if err := s.ensureAttributeGroupExists(ctx, groupID); err != nil {
		return err
	}
	if count, err := dao.CategoryAttributeGroupBindings.Ctx(ctx).Where(do.CategoryAttributeGroupBindings{GroupId: groupID}).Count(); err != nil {
		return gerror.Wrap(err, "check category group bindings")
	} else if count > 0 {
		return consts.ErrAttributeGroupInUse
	}
	result, err := dao.AttributeGroupAttributes.Ctx(ctx).Where(do.AttributeGroupAttributes{GroupId: groupID, AttributeId: attributeID}).Delete()
	if err != nil {
		return gerror.Wrap(err, "delete attribute group attribute")
	}
	if affected, err := result.RowsAffected(); err != nil {
		return gerror.Wrap(err, "check deleted attribute group attribute")
	} else if affected == 0 {
		return consts.ErrAttributeGroupAttributeNotFound
	}
	return nil
}

func (s *sAdminAttributes) ListCategoryGroups(ctx context.Context, categoryID uint64) (attributeModel.CategoryGroupBindingListResult, error) {
	if err := s.ensureCategoryExists(ctx, categoryID); err != nil {
		return attributeModel.CategoryGroupBindingListResult{}, err
	}
	cols := dao.CategoryAttributeGroupBindings.Columns()
	var records []*entity.CategoryAttributeGroupBindings
	if err := dao.CategoryAttributeGroupBindings.Ctx(ctx).Fields(
		cols.CategoryId, cols.GroupId, cols.SortOrder, cols.Status, cols.CreatedAt, cols.UpdatedAt,
	).Where(do.CategoryAttributeGroupBindings{CategoryId: categoryID}).OrderAsc(cols.SortOrder).OrderAsc(cols.GroupId).Scan(&records); err != nil {
		return attributeModel.CategoryGroupBindingListResult{}, gerror.Wrap(err, "list category groups")
	}
	category, err := s.loadCategoryEntity(ctx, categoryID)
	if err != nil {
		return attributeModel.CategoryGroupBindingListResult{}, err
	}
	items := make([]attributeModel.CategoryGroupBindingItem, 0, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		group, err := s.loadAttributeGroup(ctx, record.GroupId)
		if err != nil {
			return attributeModel.CategoryGroupBindingListResult{}, err
		}
		items = append(items, attributeModel.CategoryGroupBindingItem{
			CategoryID: category.Id, CategoryName: category.Name, GroupID: group.ID, GroupCode: group.Code,
			GroupName: group.Name, SortOrder: record.SortOrder, Status: uint8(record.Status),
			AttributeCount: group.AttributeCount, CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
		})
	}
	return attributeModel.CategoryGroupBindingListResult{Items: items}, nil
}

func (s *sAdminAttributes) BindCategoryGroup(ctx context.Context, input attributeModel.CategoryGroupBindingInput) (attributeModel.CategoryGroupBindingItem, error) {
	if input.CategoryID == 0 || input.GroupID == 0 {
		return attributeModel.CategoryGroupBindingItem{}, &consts.InvalidParameterError{Message: "category_id and group_id are required"}
	}
	if err := s.ensureCategoryExists(ctx, input.CategoryID); err != nil {
		return attributeModel.CategoryGroupBindingItem{}, err
	}
	if err := s.ensureLeafCategory(ctx, input.CategoryID); err != nil {
		return attributeModel.CategoryGroupBindingItem{}, err
	}
	if err := s.ensureAttributeGroupExists(ctx, input.GroupID); err != nil {
		return attributeModel.CategoryGroupBindingItem{}, err
	}
	if err := validateFlag(input.Status, "status"); err != nil {
		return attributeModel.CategoryGroupBindingItem{}, err
	}
	query := dao.CategoryAttributeGroupBindings.Ctx(ctx).Where(do.CategoryAttributeGroupBindings{CategoryId: input.CategoryID, GroupId: input.GroupID})
	var record *entity.CategoryAttributeGroupBindings
	if err := query.Scan(&record); err != nil {
		return attributeModel.CategoryGroupBindingItem{}, gerror.Wrap(err, "query category group binding")
	}
	data := g.Map{"category_id": input.CategoryID, "group_id": input.GroupID, "sort_order": input.SortOrder, "status": input.Status}
	if record == nil {
		if _, err := dao.CategoryAttributeGroupBindings.Ctx(ctx).Data(data).Insert(); err != nil {
			if isDuplicateKey(err) {
				return attributeModel.CategoryGroupBindingItem{}, consts.ErrCategoryGroupBindingExists
			}
			return attributeModel.CategoryGroupBindingItem{}, gerror.Wrap(err, "create category group binding")
		}
	} else if _, err := query.Data(data).Update(); err != nil {
		return attributeModel.CategoryGroupBindingItem{}, gerror.Wrap(err, "update category group binding")
	}
	items, err := s.ListCategoryGroups(ctx, input.CategoryID)
	if err != nil {
		return attributeModel.CategoryGroupBindingItem{}, err
	}
	for _, item := range items.Items {
		if item.GroupID == input.GroupID {
			return item, nil
		}
	}
	return attributeModel.CategoryGroupBindingItem{}, consts.ErrAttributeGroupNotFound
}

func (s *sAdminAttributes) UnbindCategoryGroup(ctx context.Context, categoryID, groupID uint64) error {
	if categoryID == 0 || groupID == 0 {
		return &consts.InvalidParameterError{Message: "category_id and group_id are required"}
	}
	result, err := dao.CategoryAttributeGroupBindings.Ctx(ctx).Where(do.CategoryAttributeGroupBindings{CategoryId: categoryID, GroupId: groupID}).Delete()
	if err != nil {
		return gerror.Wrap(err, "delete category group binding")
	}
	if affected, err := result.RowsAffected(); err != nil {
		return gerror.Wrap(err, "check deleted category group binding")
	} else if affected == 0 {
		return consts.ErrCategoryGroupBindingNotFound
	}
	return nil
}

func (s *sAdminAttributes) ensureAttributeExists(ctx context.Context, id uint64) error {
	if id == 0 {
		return &consts.InvalidParameterError{Message: "attribute_id is required"}
	}
	count, err := dao.Attributes.Ctx(ctx).Where(do.Attributes{Id: id}).Count()
	if err != nil {
		return gerror.Wrap(err, "query attribute")
	}
	if count == 0 {
		return consts.ErrAttributeNotFound
	}
	return nil
}

func (s *sAdminAttributes) ensureCategoryExists(ctx context.Context, id uint64) error {
	count, err := dao.Categories.Ctx(ctx).Where(do.Categories{Id: id}).Count()
	if err != nil {
		return gerror.Wrap(err, "query category")
	}
	if count == 0 {
		return consts.ErrCategoryNotFound
	}
	return nil
}

func (s *sAdminAttributes) ensureLeafCategory(ctx context.Context, id uint64) error {
	count, err := dao.Categories.Ctx(ctx).Where(do.Categories{ParentId: id}).Count()
	if err != nil {
		return gerror.Wrap(err, "check category children")
	}
	if count > 0 {
		return consts.ErrCategoryHasChildren
	}
	return nil
}

func (s *sAdminAttributes) resolveCategoryGroup(ctx context.Context, categoryID, attributeID uint64, groupID *uint64, groupName string) (uint64, error) {
	if groupID != nil && *groupID != 0 {
		if err := s.ensureCategoryGroupContainsAttribute(ctx, categoryID, *groupID, attributeID); err != nil {
			return 0, err
		}
		return *groupID, nil
	}
	groupName = strings.TrimSpace(groupName)
	if groupName == "" {
		return 0, &consts.InvalidParameterError{Message: "group_id is required for a category attribute override"}
	}
	var group *entity.AttributeGroups
	if err := dao.AttributeGroups.Ctx(ctx).Where(do.AttributeGroups{Name: groupName}).Scan(&group); err != nil {
		return 0, gerror.Wrap(err, "query attribute group")
	}
	if group == nil {
		return 0, consts.ErrAttributeGroupNotFound
	}
	if err := s.ensureCategoryGroupContainsAttribute(ctx, categoryID, group.Id, attributeID); err != nil {
		return 0, err
	}
	return group.Id, nil
}

func (s *sAdminAttributes) ensureCategoryGroupContainsAttribute(ctx context.Context, categoryID, groupID, attributeID uint64) error {
	if err := s.ensureAttributeGroupExists(ctx, groupID); err != nil {
		return err
	}
	if count, err := dao.CategoryAttributeGroupBindings.Ctx(ctx).Where(do.CategoryAttributeGroupBindings{CategoryId: categoryID, GroupId: groupID, Status: 1}).Count(); err != nil {
		return gerror.Wrap(err, "check category group binding")
	} else if count == 0 {
		return consts.ErrCategoryGroupBindingNotFound
	}
	if count, err := dao.AttributeGroupAttributes.Ctx(ctx).Where(do.AttributeGroupAttributes{GroupId: groupID, AttributeId: attributeID}).Count(); err != nil {
		return gerror.Wrap(err, "check group attribute")
	} else if count == 0 {
		return consts.ErrAttributeGroupAttributeNotFound
	}
	return nil
}

func (s *sAdminAttributes) listGroupAttributeEntities(ctx context.Context, groupID uint64) ([]*entity.AttributeGroupAttributes, error) {
	cols := dao.AttributeGroupAttributes.Columns()
	var records []*entity.AttributeGroupAttributes
	if err := dao.AttributeGroupAttributes.Ctx(ctx).Fields(
		cols.GroupId, cols.AttributeId, cols.SortOrder, cols.CreatedAt, cols.UpdatedAt,
	).Where(do.AttributeGroupAttributes{GroupId: groupID}).OrderAsc(cols.SortOrder).OrderAsc(cols.AttributeId).Scan(&records); err != nil {
		return nil, gerror.Wrap(err, "list group attributes")
	}
	return records, nil
}

func (s *sAdminAttributes) loadAttributeGroup(ctx context.Context, id uint64) (attributeModel.AttributeGroupItem, error) {
	group, err := s.loadAttributeGroupEntity(ctx, id)
	if err != nil {
		return attributeModel.AttributeGroupItem{}, err
	}
	attributeCount, err := dao.AttributeGroupAttributes.Ctx(ctx).Where(do.AttributeGroupAttributes{GroupId: id}).Count()
	if err != nil {
		return attributeModel.AttributeGroupItem{}, gerror.Wrap(err, "count group attributes")
	}
	categoryCount, err := dao.CategoryAttributeGroupBindings.Ctx(ctx).Where(do.CategoryAttributeGroupBindings{GroupId: id}).Count()
	if err != nil {
		return attributeModel.AttributeGroupItem{}, gerror.Wrap(err, "count category group bindings")
	}
	return attributeModel.AttributeGroupItem{
		ID: group.Id, Code: group.Code, Name: group.Name, SortOrder: group.SortOrder,
		Status: uint8(group.Status), AttributeCount: uint(attributeCount), CategoryCount: uint(categoryCount),
		CreatedAt: group.CreatedAt, UpdatedAt: group.UpdatedAt,
	}, nil
}

func (s *sAdminAttributes) loadAttributeGroupEntity(ctx context.Context, id uint64) (*entity.AttributeGroups, error) {
	cols := dao.AttributeGroups.Columns()
	var group *entity.AttributeGroups
	if err := dao.AttributeGroups.Ctx(ctx).Fields(
		cols.Id, cols.Code, cols.Name, cols.SortOrder, cols.Status, cols.CreatedAt, cols.UpdatedAt,
	).Where(do.AttributeGroups{Id: id}).Scan(&group); err != nil {
		return nil, gerror.Wrap(err, "query attribute group")
	}
	if group == nil {
		return nil, consts.ErrAttributeGroupNotFound
	}
	return group, nil
}

func (s *sAdminAttributes) ensureAttributeGroupExists(ctx context.Context, id uint64) error {
	if id == 0 {
		return &consts.InvalidParameterError{Message: "group_id is required"}
	}
	if _, err := s.loadAttributeGroupEntity(ctx, id); err != nil {
		return err
	}
	return nil
}

func normalizeAttributeInput(input attributeModel.AttributeInput) attributeModel.AttributeInput {
	input.Code = strings.TrimSpace(input.Code)
	input.Name = strings.TrimSpace(input.Name)
	input.InputType = strings.TrimSpace(input.InputType)
	input.Unit = strings.TrimSpace(input.Unit)
	input.Placeholder = strings.TrimSpace(input.Placeholder)
	input.ValidationRules = strings.TrimSpace(input.ValidationRules)
	input.HelpText = strings.TrimSpace(input.HelpText)
	return input
}

func normalizeOptionInput(input attributeModel.AttributeOptionInput) attributeModel.AttributeOptionInput {
	input.Code = strings.TrimSpace(input.Code)
	input.Name = strings.TrimSpace(input.Name)
	return input
}

func normalizeGroupInput(input attributeModel.AttributeGroupInput) attributeModel.AttributeGroupInput {
	input.Code = strings.TrimSpace(input.Code)
	input.Name = strings.TrimSpace(input.Name)
	return input
}

func validateAttributeInput(input attributeModel.AttributeInput) error {
	if invalidCode(input.Code, 64) {
		return &consts.InvalidParameterError{Message: "code must contain 1 to 64 non-whitespace characters"}
	}
	if input.Name == "" || len(input.Name) > 100 {
		return &consts.InvalidParameterError{Message: "name must contain 1 to 100 characters"}
	}
	if input.ValueType < 1 || input.ValueType > 6 {
		return &consts.InvalidParameterError{Message: "value_type must be between 1 and 6"}
	}
	if input.InputType == "" || len(input.InputType) > 32 {
		return &consts.InvalidParameterError{Message: "input_type must contain 1 to 32 characters"}
	}
	if len(input.Unit) > 32 || len(input.Placeholder) > 255 || len(input.HelpText) > 255 {
		return &consts.InvalidParameterError{Message: "unit, placeholder or help_text is too long"}
	}
	if input.ValidationRules != "" && !json.Valid([]byte(input.ValidationRules)) {
		return &consts.InvalidParameterError{Message: "validation_rules must be valid JSON"}
	}
	return validateFlag(input.Status, "status")
}

func validateOptionInput(input attributeModel.AttributeOptionInput) error {
	if invalidCode(input.Code, 64) {
		return &consts.InvalidParameterError{Message: "code must contain 1 to 64 non-whitespace characters"}
	}
	if input.Name == "" || len(input.Name) > 100 {
		return &consts.InvalidParameterError{Message: "name must contain 1 to 100 characters"}
	}
	return validateFlag(input.Status, "status")
}

func validateGroupInput(input attributeModel.AttributeGroupInput) error {
	if invalidCode(input.Code, 64) {
		return &consts.InvalidParameterError{Message: "code must contain 1 to 64 non-whitespace characters"}
	}
	if input.Name == "" || len(input.Name) > 100 {
		return &consts.InvalidParameterError{Message: "name must contain 1 to 100 characters"}
	}
	return validateFlag(input.Status, "status")
}

func validateFlag(value uint, field string) error {
	if value > 1 {
		return &consts.InvalidParameterError{Message: field + " must be 0 or 1"}
	}
	return nil
}

func invalidCode(value string, max int) bool {
	if value == "" || len(value) > max {
		return true
	}
	return strings.IndexFunc(value, unicode.IsSpace) >= 0
}

func attributeData(input attributeModel.AttributeInput) g.Map {
	return g.Map{
		"code": input.Code, "name": input.Name, "value_type": input.ValueType,
		"input_type": input.InputType, "unit": nullableString(input.Unit),
		"placeholder":      nullableString(input.Placeholder),
		"validation_rules": nullableString(input.ValidationRules),
		"help_text":        nullableString(input.HelpText), "status": input.Status,
	}
}

func groupData(input attributeModel.AttributeGroupInput) g.Map {
	return g.Map{
		"code": input.Code, "name": input.Name,
		"sort_order": input.SortOrder, "status": input.Status,
	}
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func (s *sAdminAttributes) validateGroupIDs(ctx context.Context, groupIDs []uint64) error {
	seen := make(map[uint64]struct{}, len(groupIDs))
	for _, groupID := range groupIDs {
		if groupID == 0 {
			return &consts.InvalidParameterError{Message: "group_ids must contain positive IDs"}
		}
		if _, ok := seen[groupID]; ok {
			return &consts.InvalidParameterError{Message: "group_ids must not contain duplicates"}
		}
		seen[groupID] = struct{}{}
		if err := s.ensureAttributeGroupExists(ctx, groupID); err != nil {
			return err
		}
	}
	return nil
}

func (s *sAdminAttributes) syncAttributeGroups(ctx context.Context, attributeID uint64, groupIDs []uint64) error {
	currentGroupIDs, err := s.listAttributeGroupIDs(ctx, attributeID)
	if err != nil {
		return err
	}
	requested := make(map[uint64]struct{}, len(groupIDs))
	for _, groupID := range groupIDs {
		requested[groupID] = struct{}{}
	}
	for _, groupID := range currentGroupIDs {
		if _, keep := requested[groupID]; keep {
			continue
		}
		count, err := dao.CategoryAttributeGroupBindings.Ctx(ctx).
			Where(do.CategoryAttributeGroupBindings{GroupId: groupID}).Count()
		if err != nil {
			return gerror.Wrap(err, "check removed attribute group bindings")
		}
		if count > 0 {
			return consts.ErrAttributeGroupInUse
		}
	}
	if _, err := dao.AttributeGroupAttributes.Ctx(ctx).Where(do.AttributeGroupAttributes{AttributeId: attributeID}).Delete(); err != nil {
		return gerror.Wrap(err, "clear attribute groups")
	}
	for index, groupID := range groupIDs {
		if _, err := dao.AttributeGroupAttributes.Ctx(ctx).Data(g.Map{
			"group_id": groupID, "attribute_id": attributeID, "sort_order": index,
		}).Insert(); err != nil {
			return gerror.Wrap(err, "save attribute group")
		}
	}
	return nil
}

func (s *sAdminAttributes) listAttributeGroupIDs(ctx context.Context, attributeID uint64) ([]uint64, error) {
	cols := dao.AttributeGroupAttributes.Columns()
	var records []*entity.AttributeGroupAttributes
	if err := dao.AttributeGroupAttributes.Ctx(ctx).Fields(cols.GroupId).
		Where(do.AttributeGroupAttributes{AttributeId: attributeID}).
		OrderAsc(cols.SortOrder).OrderAsc(cols.GroupId).Scan(&records); err != nil {
		return nil, gerror.Wrap(err, "list attribute group IDs")
	}
	ids := make([]uint64, 0, len(records))
	for _, record := range records {
		if record != nil {
			ids = append(ids, record.GroupId)
		}
	}
	return ids, nil
}

func toAttributeItem(record *entity.Attributes) attributeModel.AttributeItem {
	return attributeModel.AttributeItem{
		ID: record.Id, Code: record.Code, Name: record.Name,
		ValueType: uint8(record.ValueType), InputType: record.InputType,
		Unit: record.Unit, Placeholder: record.Placeholder,
		ValidationRules: record.ValidationRules, HelpText: record.HelpText,
		Status: uint8(record.Status), CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
	}
}

func toAttributeOptionItem(record *entity.AttributeOptions) attributeModel.AttributeOptionItem {
	return attributeModel.AttributeOptionItem{
		ID: record.Id, AttributeID: record.AttributeId, Code: record.Code,
		Name: record.Name, SortOrder: record.SortOrder, Status: uint8(record.Status),
		CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
	}
}

func isDuplicateKey(err error) bool {
	var mysqlErr *mysqlDriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
