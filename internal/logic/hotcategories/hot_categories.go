package hotcategories

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	"github.com/suxiangdong/indu-scope-backend/internal/dao"
	hotcategoryModel "github.com/suxiangdong/indu-scope-backend/internal/model/hotcategory"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

type sHotCategories struct{}

var _ service.IHotCategories = (*sHotCategories)(nil)

func New() service.IHotCategories { return &sHotCategories{} }

func init() { service.RegisterHotCategories(New()) }

func (s *sHotCategories) List(ctx context.Context, filter hotcategoryModel.ListFilter) (hotcategoryModel.ListResult, error) {
	query := dao.HotCategories.Ctx(ctx).As("hot").Fields("hot.id", "hot.category_id", "hot.sort_order", "hot.status", "hot.created_at", "hot.updated_at")
	if filter.Status != nil {
		query = query.Where("hot.status", *filter.Status)
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		categoryIDs, err := s.categoryIDsByKeyword(ctx, keyword)
		if err != nil {
			return hotcategoryModel.ListResult{}, err
		}
		if len(categoryIDs) == 0 {
			return hotcategoryModel.ListResult{Items: []hotcategoryModel.Item{}}, nil
		}
		query = query.WhereIn("hot.category_id", categoryIDs)
	}
	var records []hotCategoryRecord
	if err := query.OrderAsc("hot.sort_order").OrderAsc("hot.id").Scan(&records); err != nil {
		return hotcategoryModel.ListResult{}, gerror.Wrap(err, "list hot categories")
	}
	if err := s.hydrateCategoryNames(ctx, records); err != nil {
		return hotcategoryModel.ListResult{}, err
	}
	items := make([]hotcategoryModel.Item, 0, len(records))
	for _, record := range records {
		items = append(items, hotcategoryModel.Item{ID: record.ID, CategoryID: record.CategoryID, Name: record.Name, SortOrder: record.SortOrder, Status: record.Status, CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt})
	}
	return hotcategoryModel.ListResult{Items: items}, nil
}

func (s *sHotCategories) ListPublic(ctx context.Context) (hotcategoryModel.ListResult, error) {
	var records []hotCategoryRecord
	if err := dao.HotCategories.Ctx(ctx).Fields("id", "category_id", "sort_order", "status").
		Where("status", 1).OrderAsc("sort_order").OrderAsc("id").Scan(&records); err != nil {
		return hotcategoryModel.ListResult{}, gerror.Wrap(err, "list public hot categories")
	}
	if err := s.hydrateCategoryNames(ctx, records); err != nil {
		return hotcategoryModel.ListResult{}, err
	}
	items := make([]hotcategoryModel.Item, 0, len(records))
	for _, record := range records {
		if record.CategoryStatus != 1 {
			continue
		}
		items = append(items, hotcategoryModel.Item{ID: record.ID, CategoryID: record.CategoryID, Name: record.Name, SortOrder: record.SortOrder, Status: 1})
	}
	return hotcategoryModel.ListResult{Items: items}, nil
}

func (s *sHotCategories) Create(ctx context.Context, categoryID uint64, sortOrder uint, status uint8) (hotcategoryModel.Item, error) {
	if err := validateInput(categoryID, status); err != nil {
		return hotcategoryModel.Item{}, err
	}
	if err := s.ensureCategory(ctx, categoryID); err != nil {
		return hotcategoryModel.Item{}, err
	}
	if count, err := dao.HotCategories.Ctx(ctx).Where("category_id", categoryID).Count(); err != nil {
		return hotcategoryModel.Item{}, gerror.Wrap(err, "check hot category")
	} else if count > 0 {
		return hotcategoryModel.Item{}, consts.ErrHotCategoryAlreadyExists
	}
	id, err := dao.HotCategories.Ctx(ctx).Data(g.Map{"category_id": categoryID, "sort_order": sortOrder, "status": status}).InsertAndGetId()
	if err != nil {
		return hotcategoryModel.Item{}, gerror.Wrap(err, "create hot category")
	}
	return s.loadByID(ctx, uint64(id))
}

func (s *sHotCategories) Update(ctx context.Context, id, categoryID uint64, sortOrder uint, status uint8) (hotcategoryModel.Item, error) {
	if id == 0 {
		return hotcategoryModel.Item{}, &consts.InvalidParameterError{Message: "id is required"}
	}
	if err := validateInput(categoryID, status); err != nil {
		return hotcategoryModel.Item{}, err
	}
	if err := s.ensureCategory(ctx, categoryID); err != nil {
		return hotcategoryModel.Item{}, err
	}
	if count, err := dao.HotCategories.Ctx(ctx).Where("id", id).Count(); err != nil {
		return hotcategoryModel.Item{}, gerror.Wrap(err, "query hot category")
	} else if count == 0 {
		return hotcategoryModel.Item{}, consts.ErrHotCategoryNotFound
	}
	if count, err := dao.HotCategories.Ctx(ctx).Where("category_id", categoryID).Where("id != ?", id).Count(); err != nil {
		return hotcategoryModel.Item{}, gerror.Wrap(err, "check hot category category")
	} else if count > 0 {
		return hotcategoryModel.Item{}, consts.ErrHotCategoryAlreadyExists
	}
	if _, err := dao.HotCategories.Ctx(ctx).Cache(gdb.CacheOption{}).Where("id", id).Data(g.Map{"category_id": categoryID, "sort_order": sortOrder, "status": status}).Update(); err != nil {
		return hotcategoryModel.Item{}, gerror.Wrap(err, "update hot category")
	}
	return s.loadByID(ctx, id)
}

func (s *sHotCategories) Delete(ctx context.Context, id uint64) error {
	if id == 0 {
		return &consts.InvalidParameterError{Message: "id is required"}
	}
	result, err := dao.HotCategories.Ctx(ctx).Where("id", id).Delete()
	if err != nil {
		return gerror.Wrap(err, "delete hot category")
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return gerror.Wrap(err, "check deleted hot category")
	}
	if affected == 0 {
		return consts.ErrHotCategoryNotFound
	}
	return nil
}

func (s *sHotCategories) UpdateStatus(ctx context.Context, id uint64, status uint8) (hotcategoryModel.Item, error) {
	if id == 0 {
		return hotcategoryModel.Item{}, &consts.InvalidParameterError{Message: "id is required"}
	}
	if status > 1 {
		return hotcategoryModel.Item{}, &consts.InvalidParameterError{Message: "status must be 0 or 1"}
	}
	if count, err := dao.HotCategories.Ctx(ctx).Where("id", id).Count(); err != nil {
		return hotcategoryModel.Item{}, gerror.Wrap(err, "query hot category status")
	} else if count == 0 {
		return hotcategoryModel.Item{}, consts.ErrHotCategoryNotFound
	}
	if _, err := dao.HotCategories.Ctx(ctx).Cache(gdb.CacheOption{}).Where("id", id).Data(g.Map{"status": status}).Update(); err != nil {
		return hotcategoryModel.Item{}, gerror.Wrap(err, "update hot category status")
	}
	return s.loadByID(ctx, id)
}

func (s *sHotCategories) Sort(ctx context.Context, items []hotcategoryModel.SortItem) (hotcategoryModel.ListResult, error) {
	if len(items) == 0 {
		return hotcategoryModel.ListResult{}, &consts.InvalidParameterError{Message: "items must not be empty"}
	}
	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		seen := make(map[uint64]struct{}, len(items))
		for _, item := range items {
			if item.ID == 0 {
				return &consts.InvalidParameterError{Message: "id is required"}
			}
			if _, ok := seen[item.ID]; ok {
				return &consts.InvalidParameterError{Message: "duplicate hot category id"}
			}
			seen[item.ID] = struct{}{}
			count, err := tx.Model("hot_categories").Ctx(ctx).Where("id", item.ID).Count()
			if err != nil {
				return gerror.Wrap(err, "query hot category for sort")
			}
			if count == 0 {
				return consts.ErrHotCategoryNotFound
			}
			_, err = tx.Model("hot_categories").Ctx(ctx).Where("id", item.ID).Data(g.Map{"sort_order": item.SortOrder}).Update()
			if err != nil {
				return gerror.Wrap(err, "sort hot categories")
			}
		}
		return nil
	})
	if err != nil {
		return hotcategoryModel.ListResult{}, err
	}
	return s.List(ctx, hotcategoryModel.ListFilter{})
}

func (s *sHotCategories) ensureCategory(ctx context.Context, categoryID uint64) error {
	if categoryID == 0 {
		return &consts.InvalidParameterError{Message: "category_id is required"}
	}
	count, err := dao.Categories.Ctx(ctx).Where(g.Map{"id": categoryID, "status": 1}).Count()
	if err != nil {
		return gerror.Wrap(err, "query category for hot category")
	}
	if count == 0 {
		return consts.ErrCategoryNotFound
	}
	children, err := dao.Categories.Ctx(ctx).Where(g.Map{"parent_id": categoryID, "status": 1}).Count()
	if err != nil {
		return gerror.Wrap(err, "check hot category children")
	}
	if children > 0 {
		return &consts.InvalidParameterError{Message: "category_id must be a leaf category"}
	}
	return nil
}

func (s *sHotCategories) loadByID(ctx context.Context, id uint64) (hotcategoryModel.Item, error) {
	var record hotCategoryRecord
	if err := dao.HotCategories.Ctx(ctx).Cache(gdb.CacheOption{}).
		Fields("id", "category_id", "sort_order", "status", "created_at", "updated_at").Where("id", id).Scan(&record); err != nil {
		return hotcategoryModel.Item{}, gerror.Wrap(err, "query hot category")
	}
	if record.ID == 0 {
		return hotcategoryModel.Item{}, consts.ErrHotCategoryNotFound
	}
	records := []hotCategoryRecord{record}
	if err := s.hydrateCategoryNames(ctx, records); err != nil {
		return hotcategoryModel.Item{}, err
	}
	record = records[0]
	return hotcategoryModel.Item{ID: record.ID, CategoryID: record.CategoryID, Name: record.Name, SortOrder: record.SortOrder, Status: record.Status, CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt}, nil
}

type hotCategoryRecord struct {
	ID             uint64
	CategoryID     uint64
	Name           string
	CategoryStatus uint8
	SortOrder      uint
	Status         uint8
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (s *sHotCategories) categoryIDsByKeyword(ctx context.Context, keyword string) ([]uint64, error) {
	var categories []struct{ ID uint64 }
	if err := dao.Categories.Ctx(ctx).Fields("id").WhereLike("name", "%"+keyword+"%").Scan(&categories); err != nil {
		return nil, gerror.Wrap(err, "search hot category names")
	}
	ids := make([]uint64, 0, len(categories))
	for _, category := range categories {
		ids = append(ids, category.ID)
	}
	return ids, nil
}

func (s *sHotCategories) hydrateCategoryNames(ctx context.Context, records []hotCategoryRecord) error {
	if len(records) == 0 {
		return nil
	}
	ids := make([]uint64, 0, len(records))
	seen := make(map[uint64]struct{}, len(records))
	for _, record := range records {
		if _, ok := seen[record.CategoryID]; !ok {
			seen[record.CategoryID] = struct{}{}
			ids = append(ids, record.CategoryID)
		}
	}
	var categories []struct {
		ID     uint64
		Name   string
		Status uint8
	}
	if err := dao.Categories.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("id", "name", "status").WhereIn("id", ids).Scan(&categories); err != nil {
		return gerror.Wrap(err, "query hot category names")
	}
	byID := make(map[uint64]struct {
		Name   string
		Status uint8
	}, len(categories))
	for _, category := range categories {
		byID[category.ID] = struct {
			Name   string
			Status uint8
		}{category.Name, category.Status}
	}
	for index := range records {
		if category, ok := byID[records[index].CategoryID]; ok {
			records[index].Name = category.Name
			records[index].CategoryStatus = category.Status
		}
	}
	return nil
}

func validateInput(categoryID uint64, status uint8) error {
	if categoryID == 0 {
		return &consts.InvalidParameterError{Message: "category_id is required"}
	}
	if status > 1 {
		return &consts.InvalidParameterError{Message: "status must be 0 or 1"}
	}
	return nil
}
