package favorites

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	"github.com/suxiangdong/indu-scope-backend/internal/dao"
	"github.com/suxiangdong/indu-scope-backend/internal/model/entity"
	favoriteModel "github.com/suxiangdong/indu-scope-backend/internal/model/favorite"
	"github.com/suxiangdong/indu-scope-backend/internal/pagination"
	"github.com/suxiangdong/indu-scope-backend/internal/selectcache"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
	"github.com/suxiangdong/indu-scope-backend/internal/storage"
)

type sFavorites struct{}

var _ service.IFavorites = (*sFavorites)(nil)

func New() service.IFavorites { return &sFavorites{} }

func init() { service.RegisterFavorites(New()) }

func (s *sFavorites) List(ctx context.Context, userID uint64, page, pageSize uint) (favoriteModel.ListResult, error) {
	if userID == 0 {
		return favoriteModel.ListResult{}, consts.ErrTokenInvalid
	}
	page, pageSize = pagination.Normalize(page, pageSize)
	base := dao.ProductFavorites.Ctx(ctx).Where("user_id", userID)
	total, err := base.Count()
	if err != nil {
		return favoriteModel.ListResult{}, gerror.Wrap(err, "count product favorites")
	}

	var records []favoriteRecord
	offset := (page - 1) * pageSize
	if err := dao.ProductFavorites.Ctx(ctx).
		Fields("id AS favorite_id", "created_at AS favorited_at", "product_id").
		Where("user_id", userID).OrderDesc("created_at").OrderDesc("id").
		Limit(int(offset), int(pageSize)).Scan(&records); err != nil {
		return favoriteModel.ListResult{}, gerror.Wrap(err, "list product favorites")
	}
	if err := s.hydrateFavoriteRecords(ctx, records); err != nil {
		return favoriteModel.ListResult{}, err
	}

	ossClient, _ := storage.OSSClient()
	media, err := s.loadProductMedia(ctx, records)
	if err != nil {
		return favoriteModel.ListResult{}, err
	}
	items := make([]favoriteModel.ProductItem, 0, len(records))
	for _, record := range records {
		item := record.item()
		item.ImageCount = media.Counts[record.ProductID]
		if cover, ok := media.Covers[record.ProductID]; ok && cover.ObjectKey != "" {
			item.CoverURL = signCoverURL(ossClient, cover)
		}
		items = append(items, item)
	}
	return favoriteModel.ListResult{Items: items, Total: uint64(total), Page: page, PageSize: pageSize}, nil
}

func (s *sFavorites) Toggle(ctx context.Context, userID, productID uint64) (favoriteModel.State, error) {
	if userID == 0 {
		return favoriteModel.State{}, consts.ErrTokenInvalid
	}
	if productID == 0 {
		return favoriteModel.State{}, &consts.InvalidParameterError{Message: "product_id is required"}
	}
	return toggleFavorite(ctx, g.DB(), userID, productID)
}

// A user's stable row serializes toggles even when the favorite does not exist.
// Locking only the favorite row would leave the first toggle unprotected.
func toggleFavorite(ctx context.Context, db gdb.DB, userID, productID uint64) (favoriteModel.State, error) {
	state := favoriteModel.State{ProductID: productID}
	err := db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var user struct{ ID uint64 }
		if err := tx.Model("users").Fields("id").Where("id", userID).LockUpdate().Scan(&user); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return consts.ErrTokenInvalid
			}
			return gerror.Wrap(err, "lock favorite owner")
		}
		var favorite favoriteStateRecord
		// The owner lock is acquired before this transaction's first snapshot
		// read. Do not lock a missing favorite: InnoDB gap locks can deadlock
		// inserts for different users whose keys fall in the same index gap.
		err := tx.Model("product_favorites").Fields("id", "product_id").
			Where(g.Map{"user_id": userID, "product_id": productID}).Scan(&favorite)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return gerror.Wrap(err, "query product favorite")
		}
		product, err := readProductAvailability(tx.Model("products p"), productID)
		if err != nil {
			return err
		}
		if product.ID == 0 {
			return consts.ErrProductNotFound
		}
		if favorite.ID != 0 {
			_, err := tx.Model("product_favorites").Where("id", favorite.ID).Delete()
			return gerror.Wrap(err, "remove product favorite")
		}
		if product.Status != 1 || product.SupplierStatus != 1 || product.CategoryStatus != 1 || product.UnitStatus != 1 {
			return consts.ErrProductUnavailable
		}
		id, err := tx.Model("product_favorites").Data(g.Map{"user_id": userID, "product_id": productID}).InsertAndGetId()
		if err != nil {
			return gerror.Wrap(err, "save product favorite")
		}
		state.Favorited, state.FavoriteID = true, uint64(id)
		return nil
	})
	return state, err
}

func (s *sFavorites) Status(ctx context.Context, userID, productID uint64) (favoriteModel.State, error) {
	if userID == 0 {
		return favoriteModel.State{}, consts.ErrTokenInvalid
	}
	if productID == 0 {
		return favoriteModel.State{}, &consts.InvalidParameterError{Message: "product_id is required"}
	}
	product, err := s.loadProductAvailability(ctx, productID)
	if err != nil {
		return favoriteModel.State{}, err
	}
	if product.ID == 0 {
		return favoriteModel.State{}, consts.ErrProductNotFound
	}
	if product.Status != 1 || product.SupplierStatus != 1 || product.CategoryStatus != 1 || product.UnitStatus != 1 {
		return favoriteModel.State{}, consts.ErrProductUnavailable
	}
	favorite, err := s.loadFavoriteState(ctx, userID, productID)
	if err != nil {
		return favoriteModel.State{}, err
	}
	return favoriteModel.State{ProductID: productID, Favorited: favorite.ID != 0, FavoriteID: favorite.ID}, nil
}

func (s *sFavorites) loadFavoriteState(ctx context.Context, userID, productID uint64) (favoriteStateRecord, error) {
	var favorite favoriteStateRecord
	err := dao.ProductFavorites.Ctx(ctx).Cache(gdb.CacheOption{}).Where(g.Map{"user_id": userID, "product_id": productID}).Fields("id", "product_id").Scan(&favorite)
	if errors.Is(err, sql.ErrNoRows) {
		return favoriteStateRecord{}, nil
	}
	return favorite, gerror.Wrap(err, "query product favorite")
}

type productAvailability struct {
	ID             uint64
	Status         uint8
	SupplierStatus uint8
	CategoryStatus uint8
	UnitStatus     uint8
}

func (s *sFavorites) loadProductAvailability(ctx context.Context, productID uint64) (productAvailability, error) {
	return readProductAvailability(g.DB().Model("products p").Ctx(ctx), productID)
}

func readProductAvailability(model *gdb.Model, productID uint64) (productAvailability, error) {
	var product productAvailability
	err := model.LeftJoin("users supplier", "supplier.id = p.supplier_user_id").
		LeftJoin("categories category", "category.id = p.category_id").
		LeftJoin("stock_units unit", "unit.id = p.stock_unit_id").
		Fields("p.id, p.status, COALESCE(supplier.status, 0) AS supplier_status, COALESCE(category.status, 0) AS category_status, COALESCE(unit.status, 0) AS unit_status").
		Where("p.id", productID).Scan(&product)
	if errors.Is(err, sql.ErrNoRows) {
		return productAvailability{}, nil
	}
	return product, gerror.Wrap(err, "query favorite product availability")
}

type favoriteRecord struct {
	FavoriteID      uint64
	FavoritedAt     time.Time
	ProductID       uint64
	ProductName     string
	SupplierUserID  uint64
	SupplierCompany string
	CategoryID      uint64
	CategoryName    string
	StockQuantity   float64
	StockUnitName   string
	Status          uint8
}

func (s *sFavorites) hydrateFavoriteRecords(ctx context.Context, records []favoriteRecord) error {
	for index := range records {
		var product entity.Products
		if err := dao.Products.Ctx(ctx).Cache(selectcache.Query(dao.Products.Table(), records[index].ProductID)).Where("id", records[index].ProductID).Scan(&product); err != nil {
			return gerror.Wrap(err, "query favorite product")
		}
		if product.Id == 0 {
			continue
		}
		records[index].ProductName = product.Name
		records[index].SupplierUserID = product.SupplierUserId
		records[index].CategoryID = product.CategoryId
		records[index].StockQuantity = product.StockQuantity
		records[index].Status = uint8(product.Status)

		var supplier struct {
			ID          uint64
			CompanyName string `orm:"company_name"`
			Status      uint8
		}
		if err := dao.Users.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("id", "company_name", "status").Where("id", product.SupplierUserId).Scan(&supplier); err != nil {
			return gerror.Wrap(err, "query favorite supplier")
		}
		records[index].SupplierCompany = supplier.CompanyName

		var category struct {
			ID     uint64
			Name   string
			Status uint8
		}
		if err := dao.Categories.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("id", "name", "status").Where("id", product.CategoryId).Scan(&category); err != nil {
			return gerror.Wrap(err, "query favorite category")
		}
		records[index].CategoryName = category.Name

		var unit struct {
			ID     uint64
			Name   string
			Status uint8
		}
		if err := dao.StockUnits.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("id", "name", "status").Where("id", product.StockUnitId).Scan(&unit); err != nil {
			return gerror.Wrap(err, "query favorite stock unit")
		}
		records[index].StockUnitName = unit.Name
	}
	return nil
}

func (r favoriteRecord) item() favoriteModel.ProductItem {
	return favoriteModel.ProductItem{ID: r.FavoriteID, ProductID: r.ProductID, ProductName: r.ProductName, SupplierUserID: r.SupplierUserID, SupplierCompany: r.SupplierCompany, CategoryID: r.CategoryID, CategoryName: r.CategoryName, StockQuantity: r.StockQuantity, StockUnitName: r.StockUnitName, Status: r.Status, FavoritedAt: r.FavoritedAt}
}

type favoriteStateRecord struct {
	ID        uint64
	ProductID uint64
}

type coverRecord struct {
	ProductID  uint64 `orm:"product_id"`
	UserFileID uint64 `orm:"user_file_id"`
	Endpoint   string
	Bucket     string
	ObjectKey  string `orm:"object_key"`
}

type productMedia struct {
	Covers map[uint64]coverRecord
	Counts map[uint64]uint
}

func (s *sFavorites) loadProductMedia(ctx context.Context, records []favoriteRecord) (productMedia, error) {
	media := productMedia{Covers: make(map[uint64]coverRecord), Counts: make(map[uint64]uint)}
	if len(records) == 0 {
		return media, nil
	}
	ids := make([]uint64, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.ProductID)
	}
	covers, err := loadFavoriteProductImages(ctx, ids)
	if err != nil {
		return media, gerror.Wrap(err, "query favorite product media")
	}
	for _, cover := range covers {
		media.Counts[cover.ProductID]++
		if _, exists := media.Covers[cover.ProductID]; !exists {
			media.Covers[cover.ProductID] = cover
		}
	}
	return media, nil
}

func loadCover(ctx context.Context, productID uint64) (coverRecord, error) {
	covers, err := loadFavoriteProductImages(ctx, []uint64{productID})
	if err != nil {
		return coverRecord{}, gerror.Wrap(err, "query favorite product cover")
	}
	if len(covers) == 0 {
		return coverRecord{}, nil
	}
	return covers[0], nil
}

func loadFavoriteProductImages(ctx context.Context, productIDs []uint64) ([]coverRecord, error) {
	if len(productIDs) == 0 {
		return []coverRecord{}, nil
	}
	var covers []coverRecord
	if err := dao.ProductImages.Ctx(ctx).
		Fields("product_id", "id", "user_file_id", "sort_order").
		WhereIn("product_id", productIDs).
		OrderAsc("product_id").OrderAsc("sort_order").OrderAsc("id").Scan(&covers); err != nil {
		return nil, err
	}
	fileIDs := make([]uint64, 0, len(covers))
	seen := make(map[uint64]struct{}, len(covers))
	for _, cover := range covers {
		if cover.UserFileID != 0 {
			if _, ok := seen[cover.UserFileID]; !ok {
				seen[cover.UserFileID] = struct{}{}
				fileIDs = append(fileIDs, cover.UserFileID)
			}
		}
	}
	if len(fileIDs) == 0 {
		return covers, nil
	}
	var files []struct {
		ID        uint64
		Endpoint  string
		Bucket    string
		ObjectKey string `orm:"object_key"`
	}
	if err := dao.UserFiles.Ctx(ctx).Fields("id", "endpoint", "bucket", "object_key").WhereIn("id", fileIDs).Scan(&files); err != nil {
		return nil, err
	}
	fileByID := make(map[uint64]struct{ Endpoint, Bucket, ObjectKey string }, len(files))
	for _, file := range files {
		fileByID[file.ID] = struct{ Endpoint, Bucket, ObjectKey string }{file.Endpoint, file.Bucket, file.ObjectKey}
	}
	for index := range covers {
		if file, ok := fileByID[covers[index].UserFileID]; ok {
			covers[index].Endpoint, covers[index].Bucket, covers[index].ObjectKey = file.Endpoint, file.Bucket, file.ObjectKey
		}
	}
	return covers, nil
}

func signCoverURL(client *storage.Client, cover coverRecord) string {
	if client == nil || cover.Endpoint == "" || cover.Bucket == "" || cover.ObjectKey == "" {
		return ""
	}
	url, err := client.SignedURLFor(cover.Endpoint, cover.Bucket, cover.ObjectKey, 30*time.Minute)
	if err != nil {
		return ""
	}
	return url
}
