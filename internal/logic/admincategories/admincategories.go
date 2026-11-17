package admincategories

import (
	"context"
	"strings"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	"github.com/suxiangdong/indu-scope-backend/internal/dao"
	categorymodel "github.com/suxiangdong/indu-scope-backend/internal/model/category"
	"github.com/suxiangdong/indu-scope-backend/internal/model/do"
	"github.com/suxiangdong/indu-scope-backend/internal/model/entity"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

type sAdminCategories struct{}

var _ service.IAdminCategories = (*sAdminCategories)(nil)

func New() service.IAdminCategories {
	return &sAdminCategories{}
}

func init() {
	service.RegisterAdminCategories(New())
}

func (s *sAdminCategories) List(ctx context.Context) (categorymodel.CategoryListResult, error) {
	cols := dao.Categories.Columns()
	var records []*entity.Categories
	if err := dao.Categories.Ctx(ctx).Fields(
		cols.Id, cols.ParentId, cols.Name, cols.SortOrder, cols.Status,
		cols.CreatedAt, cols.UpdatedAt,
	).OrderAsc(cols.ParentId).OrderAsc(cols.SortOrder).OrderAsc(cols.Id).Scan(&records); err != nil {
		return categorymodel.CategoryListResult{}, gerror.Wrap(err, "list categories")
	}
	items := make([]categorymodel.CategoryItem, 0, len(records))
	for _, record := range records {
		if record != nil {
			items = append(items, toCategoryItem(record))
		}
	}
	return categorymodel.CategoryListResult{Items: items}, nil
}

func (s *sAdminCategories) Create(ctx context.Context, parentID *uint64, name string, sortOrder uint, status uint) (categorymodel.CategoryItem, error) {
	name = strings.TrimSpace(name)
	parentID = normalizeParentID(parentID)
	if err := validateCategory(name, status); err != nil {
		return categorymodel.CategoryItem{}, err
	}
	var id uint64
	err := dao.Categories.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var err error
		id, err = createCategoryTx(ctx, tx, parentID, name, sortOrder, status)
		return err
	})
	if err != nil {
		return categorymodel.CategoryItem{}, err
	}
	return s.loadByID(ctx, id)
}

func (s *sAdminCategories) Update(ctx context.Context, id uint64, parentID *uint64, name string, sortOrder uint, status uint) (categorymodel.CategoryItem, error) {
	if id == 0 {
		return categorymodel.CategoryItem{}, &consts.InvalidParameterError{Message: "id is required"}
	}
	name = strings.TrimSpace(name)
	parentID = normalizeParentID(parentID)
	if err := validateCategory(name, status); err != nil {
		return categorymodel.CategoryItem{}, err
	}
	if err := dao.Categories.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		return updateCategoryTx(ctx, tx, id, parentID, name, sortOrder, status)
	}); err != nil {
		return categorymodel.CategoryItem{}, err
	}
	return s.loadByID(ctx, id)
}

func (s *sAdminCategories) Delete(ctx context.Context, id uint64) error {
	if id == 0 {
		return &consts.InvalidParameterError{Message: "id is required"}
	}
	return dao.Categories.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		return deleteCategoryTx(ctx, tx, id)
	})
}

func (s *sAdminCategories) loadByID(ctx context.Context, id uint64) (categorymodel.CategoryItem, error) {
	cols := dao.Categories.Columns()
	var record *entity.Categories
	if err := dao.Categories.Ctx(ctx).Fields(
		cols.Id, cols.ParentId, cols.Name, cols.SortOrder, cols.Status,
		cols.CreatedAt, cols.UpdatedAt,
	).Where(do.Categories{Id: id}).Scan(&record); err != nil {
		return categorymodel.CategoryItem{}, gerror.Wrap(err, "query category")
	}
	if record == nil {
		return categorymodel.CategoryItem{}, consts.ErrCategoryNotFound
	}
	return toCategoryItem(record), nil
}

func validateCategory(name string, status uint) error {
	if name == "" || len(name) > 100 {
		return &consts.InvalidParameterError{Message: "name must contain 1 to 100 characters"}
	}
	if status > 1 {
		return &consts.InvalidParameterError{Message: "status must be 0 or 1"}
	}
	return nil
}

func validateParentTx(ctx context.Context, tx gdb.TX, categoryID uint64, parentID *uint64) error {
	// A zero parent ID represents a top-level category. It is stored as 0 in
	// the database, so it must not be looked up as an actual category.
	if parentID == nil || *parentID == 0 {
		return nil
	}
	if categoryID != 0 && *parentID == categoryID {
		return consts.ErrCategoryParentCycle
	}
	if count, err := tx.Model("categories").Ctx(ctx).Where(do.Categories{Id: *parentID}).Count(); err != nil {
		return gerror.Wrap(err, "query category parent")
	} else if count == 0 {
		return consts.ErrCategoryParentNotFound
	}

	visited := make(map[uint64]struct{})
	currentID := *parentID
	for currentID != 0 {
		if categoryID != 0 && currentID == categoryID {
			return consts.ErrCategoryParentCycle
		}
		if _, ok := visited[currentID]; ok {
			return consts.ErrCategoryParentCycle
		}
		visited[currentID] = struct{}{}
		var record *entity.Categories
		if err := tx.Model("categories").Ctx(ctx).Fields(dao.Categories.Columns().Id, dao.Categories.Columns().ParentId).
			Where(do.Categories{Id: currentID}).Scan(&record); err != nil {
			return gerror.Wrap(err, "query category ancestry")
		}
		if record == nil {
			return consts.ErrCategoryParentNotFound
		}
		currentID = record.ParentId
	}
	return nil
}

func ensureSiblingNameAvailableTx(ctx context.Context, tx gdb.TX, categoryID uint64, parentID *uint64, name string) error {
	cols := dao.Categories.Columns()
	query := tx.Model("categories").Ctx(ctx).Wheref("%s = ?", cols.Name, name)
	query = query.Where(do.Categories{ParentId: parentValue(parentID)})
	if categoryID != 0 {
		query = query.Wheref("%s <> ?", cols.Id, categoryID)
	}
	count, err := query.Count()
	if err != nil {
		return gerror.Wrap(err, "check category name")
	}
	if count > 0 {
		return consts.ErrCategoryAlreadyExists
	}
	return nil
}

func parentValue(parentID *uint64) any {
	if parentID == nil {
		return uint64(0)
	}
	return *parentID
}

func normalizeParentID(parentID *uint64) *uint64 {
	if parentID == nil {
		value := uint64(0)
		return &value
	}
	value := *parentID
	return &value
}

func toCategoryItem(record *entity.Categories) categorymodel.CategoryItem {
	parentID := record.ParentId
	return categorymodel.CategoryItem{
		ID: record.Id, ParentID: &parentID, Name: record.Name,
		SortOrder: record.SortOrder, Status: uint8(record.Status),
		CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
	}
}
