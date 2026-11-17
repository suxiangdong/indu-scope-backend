package adminstockunit

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	"github.com/suxiangdong/indu-scope-backend/internal/dao"
	stockUnitModel "github.com/suxiangdong/indu-scope-backend/internal/model/stockunit"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

type sAdminStockUnit struct{}

var _ service.IAdminStockUnit = (*sAdminStockUnit)(nil)

func New() service.IAdminStockUnit { return &sAdminStockUnit{} }

func init() { service.RegisterAdminStockUnit(New()) }

func (s *sAdminStockUnit) List(ctx context.Context) (stockUnitModel.ListResult, error) {
	var records []stockUnitRecord
	if err := dao.StockUnits.Ctx(ctx).
		Fields("id", "code", "name", "sort_order", "status", "created_at", "updated_at").
		OrderAsc("sort_order").OrderAsc("id").Scan(&records); err != nil {
		return stockUnitModel.ListResult{}, gerror.Wrap(err, "list stock units")
	}
	items := make([]stockUnitModel.Item, 0, len(records))
	for _, record := range records {
		items = append(items, record.item())
	}
	return stockUnitModel.ListResult{Items: items}, nil
}

func (s *sAdminStockUnit) Create(ctx context.Context, input stockUnitModel.Input) (stockUnitModel.Item, error) {
	input = normalizeInput(input)
	if err := validateInput(input); err != nil {
		return stockUnitModel.Item{}, err
	}
	if count, err := dao.StockUnits.Ctx(ctx).Where("code", input.Code).Count(); err != nil {
		return stockUnitModel.Item{}, gerror.Wrap(err, "check stock unit code")
	} else if count > 0 {
		return stockUnitModel.Item{}, consts.ErrStockUnitAlreadyExists
	}
	if count, err := dao.StockUnits.Ctx(ctx).Where("name", input.Name).Count(); err != nil {
		return stockUnitModel.Item{}, gerror.Wrap(err, "check stock unit name")
	} else if count > 0 {
		return stockUnitModel.Item{}, consts.ErrStockUnitAlreadyExists
	}
	id, err := dao.StockUnits.Ctx(ctx).Data(g.Map{
		"code": input.Code, "name": input.Name, "sort_order": input.SortOrder, "status": input.Status,
	}).InsertAndGetId()
	if err != nil {
		return stockUnitModel.Item{}, gerror.Wrap(err, "create stock unit")
	}
	return s.loadByID(ctx, uint64(id))
}

func (s *sAdminStockUnit) Update(ctx context.Context, id uint64, input stockUnitModel.Input) (stockUnitModel.Item, error) {
	if id == 0 {
		return stockUnitModel.Item{}, &consts.InvalidParameterError{Message: "id is required"}
	}
	input = normalizeInput(input)
	if err := validateInput(input); err != nil {
		return stockUnitModel.Item{}, err
	}
	if count, err := dao.StockUnits.Ctx(ctx).Where("id", id).Count(); err != nil {
		return stockUnitModel.Item{}, gerror.Wrap(err, "query stock unit")
	} else if count == 0 {
		return stockUnitModel.Item{}, consts.ErrStockUnitNotFound
	}
	if count, err := dao.StockUnits.Ctx(ctx).Where("code", input.Code).Where("id != ?", id).Count(); err != nil {
		return stockUnitModel.Item{}, gerror.Wrap(err, "check stock unit code")
	} else if count > 0 {
		return stockUnitModel.Item{}, consts.ErrStockUnitAlreadyExists
	}
	if count, err := dao.StockUnits.Ctx(ctx).Where("name", input.Name).Where("id != ?", id).Count(); err != nil {
		return stockUnitModel.Item{}, gerror.Wrap(err, "check stock unit name")
	} else if count > 0 {
		return stockUnitModel.Item{}, consts.ErrStockUnitAlreadyExists
	}
	if _, err := dao.StockUnits.Ctx(ctx).Cache(gdb.CacheOption{}).Where("id", id).Data(g.Map{
		"code": input.Code, "name": input.Name, "sort_order": input.SortOrder, "status": input.Status,
	}).Update(); err != nil {
		return stockUnitModel.Item{}, gerror.Wrap(err, "update stock unit")
	}
	return s.loadByID(ctx, id)
}

func (s *sAdminStockUnit) Delete(ctx context.Context, id uint64) error {
	if id == 0 {
		return &consts.InvalidParameterError{Message: "id is required"}
	}
	if count, err := dao.StockUnits.Ctx(ctx).Where("id", id).Count(); err != nil {
		return gerror.Wrap(err, "query stock unit")
	} else if count == 0 {
		return consts.ErrStockUnitNotFound
	}
	if count, err := dao.Products.Ctx(ctx).Where("stock_unit_id", id).Count(); err != nil {
		return gerror.Wrap(err, "check stock unit products")
	} else if count > 0 {
		return consts.ErrStockUnitInUse
	}
	if _, err := dao.StockUnits.Ctx(ctx).Where("id", id).Delete(); err != nil {
		return gerror.Wrap(err, "delete stock unit")
	}
	return nil
}

func (s *sAdminStockUnit) ListCategories(ctx context.Context, stockUnitID uint64) (stockUnitModel.CategoryListResult, error) {
	if stockUnitID == 0 {
		return stockUnitModel.CategoryListResult{}, &consts.InvalidParameterError{Message: "stock_unit_id is required"}
	}
	var records []stockUnitCategoryRecord
	if err := dao.CategoryStockUnits.Ctx(ctx).
		Fields("category_id", "sort_order", "created_at", "updated_at").
		Where("stock_unit_id", stockUnitID).OrderAsc("sort_order").OrderAsc("category_id").Scan(&records); err != nil {
		return stockUnitModel.CategoryListResult{}, gerror.Wrap(err, "list stock unit categories")
	}
	if err := hydrateCategoryNames(ctx, records); err != nil {
		return stockUnitModel.CategoryListResult{}, err
	}
	items := make([]stockUnitModel.CategoryItem, 0, len(records))
	for _, record := range records {
		items = append(items, record.categoryItem())
	}
	return stockUnitModel.CategoryListResult{Items: items}, nil
}

func (s *sAdminStockUnit) BindCategory(ctx context.Context, stockUnitID, categoryID uint64, sortOrder uint) (stockUnitModel.CategoryItem, error) {
	if stockUnitID == 0 || categoryID == 0 {
		return stockUnitModel.CategoryItem{}, &consts.InvalidParameterError{Message: "stock_unit_id and category_id are required"}
	}
	if count, err := dao.StockUnits.Ctx(ctx).Where("id", stockUnitID).Count(); err != nil {
		return stockUnitModel.CategoryItem{}, gerror.Wrap(err, "query stock unit")
	} else if count == 0 {
		return stockUnitModel.CategoryItem{}, consts.ErrStockUnitNotFound
	}
	if count, err := dao.Categories.Ctx(ctx).Where("id", categoryID).Count(); err != nil {
		return stockUnitModel.CategoryItem{}, gerror.Wrap(err, "query category")
	} else if count == 0 {
		return stockUnitModel.CategoryItem{}, consts.ErrCategoryNotFound
	}
	if count, err := dao.Categories.Ctx(ctx).Where("parent_id", categoryID).Count(); err != nil {
		return stockUnitModel.CategoryItem{}, gerror.Wrap(err, "check category children")
	} else if count > 0 {
		return stockUnitModel.CategoryItem{}, consts.ErrCategoryHasChildren
	}
	if count, err := dao.CategoryStockUnits.Ctx(ctx).Where(g.Map{"stock_unit_id": stockUnitID, "category_id": categoryID}).Count(); err != nil {
		return stockUnitModel.CategoryItem{}, gerror.Wrap(err, "check stock unit category binding")
	} else if count > 0 {
		return stockUnitModel.CategoryItem{}, consts.ErrStockUnitCategoryExists
	}
	if _, err := dao.CategoryStockUnits.Ctx(ctx).Data(g.Map{
		"stock_unit_id": stockUnitID, "category_id": categoryID, "sort_order": sortOrder,
	}).Insert(); err != nil {
		return stockUnitModel.CategoryItem{}, gerror.Wrap(err, "bind stock unit category")
	}
	return s.loadCategoryBinding(ctx, stockUnitID, categoryID)
}

func (s *sAdminStockUnit) UnbindCategory(ctx context.Context, stockUnitID, categoryID uint64) error {
	if stockUnitID == 0 || categoryID == 0 {
		return &consts.InvalidParameterError{Message: "stock_unit_id and category_id are required"}
	}
	result, err := dao.CategoryStockUnits.Ctx(ctx).Where(g.Map{"stock_unit_id": stockUnitID, "category_id": categoryID}).Delete()
	if err != nil {
		return gerror.Wrap(err, "unbind stock unit category")
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return gerror.Wrap(err, "check unbound stock unit category")
	}
	if affected == 0 {
		return consts.ErrStockUnitCategoryNotFound
	}
	return nil
}

type stockUnitRecord struct {
	ID        uint64
	Code      string
	Name      string
	SortOrder uint
	Status    uint8
	CreatedAt time.Time
	UpdatedAt time.Time
}

type stockUnitCategoryRecord struct {
	CategoryID   uint64
	CategoryName string
	SortOrder    uint
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (r stockUnitRecord) item() stockUnitModel.Item {
	return stockUnitModel.Item{ID: r.ID, Code: r.Code, Name: r.Name, SortOrder: r.SortOrder, Status: r.Status, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

func (r stockUnitCategoryRecord) categoryItem() stockUnitModel.CategoryItem {
	return stockUnitModel.CategoryItem{CategoryID: r.CategoryID, CategoryName: r.CategoryName, SortOrder: r.SortOrder, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

func (s *sAdminStockUnit) loadByID(ctx context.Context, id uint64) (stockUnitModel.Item, error) {
	var record stockUnitRecord
	if err := dao.StockUnits.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("id", "code", "name", "sort_order", "status", "created_at", "updated_at").Where("id", id).Scan(&record); err != nil {
		return stockUnitModel.Item{}, gerror.Wrap(err, "query stock unit")
	}
	if record.ID == 0 {
		return stockUnitModel.Item{}, consts.ErrStockUnitNotFound
	}
	return record.item(), nil
}

func (s *sAdminStockUnit) loadCategoryBinding(ctx context.Context, stockUnitID, categoryID uint64) (stockUnitModel.CategoryItem, error) {
	var record stockUnitCategoryRecord
	if err := dao.CategoryStockUnits.Ctx(ctx).
		Fields("category_id", "sort_order", "created_at", "updated_at").
		Where("stock_unit_id", stockUnitID).Where("category_id", categoryID).Scan(&record); err != nil {
		return stockUnitModel.CategoryItem{}, gerror.Wrap(err, "query stock unit category binding")
	}
	if record.CategoryID == 0 {
		return stockUnitModel.CategoryItem{}, consts.ErrStockUnitCategoryNotFound
	}
	records := []stockUnitCategoryRecord{record}
	if err := hydrateCategoryNames(ctx, records); err != nil {
		return stockUnitModel.CategoryItem{}, err
	}
	record = records[0]
	return record.categoryItem(), nil
}

func hydrateCategoryNames(ctx context.Context, records []stockUnitCategoryRecord) error {
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
		ID   uint64
		Name string
	}
	if err := dao.Categories.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("id", "name").WhereIn("id", ids).Scan(&categories); err != nil {
		return gerror.Wrap(err, "query stock unit category names")
	}
	names := make(map[uint64]string, len(categories))
	for _, category := range categories {
		names[category.ID] = category.Name
	}
	for index := range records {
		records[index].CategoryName = names[records[index].CategoryID]
	}
	return nil
}

func normalizeInput(input stockUnitModel.Input) stockUnitModel.Input {
	input.Code = strings.ToLower(strings.TrimSpace(input.Code))
	input.Name = strings.TrimSpace(input.Name)
	return input
}

func validateInput(input stockUnitModel.Input) error {
	if input.Code == "" || len([]rune(input.Code)) > 32 {
		return &consts.InvalidParameterError{Message: "code must contain 1 to 32 characters"}
	}
	if input.Name == "" || len([]rune(input.Name)) > 32 {
		return &consts.InvalidParameterError{Message: "name must contain 1 to 32 characters"}
	}
	if input.Status > 1 {
		return &consts.InvalidParameterError{Message: "status must be 0 or 1"}
	}
	return nil
}
