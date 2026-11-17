package productcatalog

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	"github.com/suxiangdong/indu-scope-backend/internal/dao"
	"github.com/suxiangdong/indu-scope-backend/internal/model/do"
	"github.com/suxiangdong/indu-scope-backend/internal/model/entity"
	productCatalogModel "github.com/suxiangdong/indu-scope-backend/internal/model/productcatalog"
	"github.com/suxiangdong/indu-scope-backend/internal/pagination"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
	"github.com/suxiangdong/indu-scope-backend/internal/storage"
)

type sProductCatalog struct{}

var _ service.IProductCatalog = (*sProductCatalog)(nil)

func New() service.IProductCatalog { return &sProductCatalog{} }

func init() { service.RegisterProductCatalog(New()) }

func (s *sProductCatalog) List(ctx context.Context, buyerUserID uint64, filter productCatalogModel.ListFilter) (productCatalogModel.ListResult, error) {
	if buyerUserID == 0 {
		return productCatalogModel.ListResult{}, consts.ErrTokenInvalid
	}
	filter.Page, filter.PageSize = pagination.Normalize(filter.Page, filter.PageSize)

	query := s.publishedProductQuery(ctx, filter)
	total, err := query.Count()
	if err != nil {
		return productCatalogModel.ListResult{}, gerror.Wrap(err, "count buyer products")
	}

	var records []productRecord
	offset := (filter.Page - 1) * filter.PageSize
	if err := s.publishedProductQuery(ctx, filter).
		Fields("id", "supplier_user_id", "category_id", "name", "stock_quantity", "stock_unit_id", "status", "created_at").
		OrderDesc("created_at").OrderDesc("id").
		Limit(int(offset), int(filter.PageSize)).Scan(&records); err != nil {
		return productCatalogModel.ListResult{}, gerror.Wrap(err, "list buyer products")
	}
	if err := s.hydrateProductRecords(ctx, records); err != nil {
		return productCatalogModel.ListResult{}, err
	}

	favorited, err := s.favoriteProductIDs(ctx, buyerUserID, records)
	if err != nil {
		return productCatalogModel.ListResult{}, err
	}
	media, err := s.loadProductMedia(ctx, records)
	if err != nil {
		return productCatalogModel.ListResult{}, err
	}
	ossClient, _ := storage.OSSClient()
	items := make([]productCatalogModel.ProductItem, 0, len(records))
	for _, record := range records {
		item := record.item()
		item.Favorited = favorited[record.ID]

		item.ImageCount = media.Counts[record.ID]
		if cover, ok := media.Covers[record.ID]; ok && cover.ObjectKey != "" {
			item.CoverURL = signImageURL(ossClient, cover)
		}
		items = append(items, item)
	}

	return productCatalogModel.ListResult{
		Items: items, Total: uint64(total), Page: filter.Page, PageSize: filter.PageSize,
	}, nil
}

func (s *sProductCatalog) Detail(ctx context.Context, buyerUserID, productID uint64) (productCatalogModel.ProductDetail, error) {
	if buyerUserID == 0 {
		return productCatalogModel.ProductDetail{}, consts.ErrTokenInvalid
	}
	if productID == 0 {
		return productCatalogModel.ProductDetail{}, &consts.InvalidParameterError{Message: "id is required"}
	}
	var product entity.Products
	if err := dao.Products.Ctx(ctx).Cache(gdb.CacheOption{}).Where("id", productID).Scan(&product); err != nil {
		return productCatalogModel.ProductDetail{}, gerror.Wrap(err, "query buyer product availability")
	}
	if product.Id == 0 {
		return productCatalogModel.ProductDetail{}, consts.ErrProductNotFound
	}
	var supplier entity.Users
	if err := dao.Users.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("id", "status", "company_name").Where(do.Users{Id: product.SupplierUserId}).Scan(&supplier); err != nil {
		return productCatalogModel.ProductDetail{}, gerror.Wrap(err, "query buyer product supplier")
	}
	var category entity.Categories
	if err := dao.Categories.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("id", "status", "name").Where(do.Categories{Id: product.CategoryId}).Scan(&category); err != nil {
		return productCatalogModel.ProductDetail{}, gerror.Wrap(err, "query buyer product category")
	}
	var unit struct {
		Id     uint64
		Code   string
		Name   string
		Status uint8
	}
	if err := dao.StockUnits.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("id", "code", "name", "status").Where("id", product.StockUnitId).Scan(&unit); err != nil {
		return productCatalogModel.ProductDetail{}, gerror.Wrap(err, "query buyer product stock unit")
	}
	if product.Status != 1 || supplier.Status != 1 || category.Status != 1 || unit.Status != 1 {
		return productCatalogModel.ProductDetail{}, consts.ErrProductUnavailable
	}
	supplierProfile, err := service.UserProfile().Profile(ctx, product.SupplierUserId)
	if err != nil {
		return productCatalogModel.ProductDetail{}, gerror.Wrap(err, "query buyer product supplier contacts")
	}
	record := productRecord{ID: product.Id, SupplierUserID: product.SupplierUserId, SupplierCompany: supplier.CompanyName,
		CategoryID: product.CategoryId, CategoryName: category.Name, Name: product.Name, StockQuantity: product.StockQuantity,
		StockUnitID: product.StockUnitId, StockUnitCode: unit.Code, StockUnitName: unit.Name, Status: uint8(product.Status), CreatedAt: product.CreatedAt}
	adminItem, err := service.AdminProducts().Detail(ctx, productID)
	if err != nil {
		return productCatalogModel.ProductDetail{}, err
	}
	favorited, err := s.favoriteProductIDs(ctx, buyerUserID, []productRecord{record})
	if err != nil {
		return productCatalogModel.ProductDetail{}, err
	}
	item := record.item()
	item.Favorited = favorited[productID]
	item.ImageCount = uint(len(adminItem.Images))
	if len(adminItem.Images) > 0 {
		item.CoverURL = adminItem.Images[0].URL
	}
	detail := productCatalogModel.ProductDetail{ProductItem: item, Images: make([]productCatalogModel.ProductImage, 0, len(adminItem.Images)), Attributes: make([]productCatalogModel.AttributeValue, 0, len(adminItem.Attributes)), SupplierContacts: supplierProfile.Contacts}
	for _, image := range adminItem.Images {
		detail.Images = append(detail.Images, productCatalogModel.ProductImage{ID: image.ID, URL: image.URL, Path: image.Path, OriginalName: image.OriginalName, ContentType: image.ContentType, Size: image.Size, SortOrder: image.SortOrder})
	}
	for _, attribute := range adminItem.Attributes {
		detail.Attributes = append(detail.Attributes, productCatalogModel.AttributeValue{AttributeID: attribute.AttributeID, Code: attribute.Code, Name: attribute.Name, Unit: attribute.Unit, Value: attribute.Value})
	}
	return detail, nil
}

func (s *sProductCatalog) publishedProductQuery(ctx context.Context, filter productCatalogModel.ListFilter) *gdb.Model {
	model := dao.Products.Ctx(ctx).
		Where("status", 1).
		Where("supplier_user_id IN (SELECT id FROM users WHERE status = 1)").
		Where("category_id IN (SELECT id FROM categories WHERE status = 1)").
		Where("stock_unit_id IN (SELECT id FROM stock_units WHERE status = 1)")
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		pattern := "%" + keyword + "%"
		model = model.Where("(name LIKE ? OR supplier_user_id IN (SELECT id FROM users WHERE status = 1 AND company_name LIKE ?) OR category_id IN (SELECT id FROM categories WHERE status = 1 AND name LIKE ?))", pattern, pattern, pattern)
	}
	if filter.CategoryID != 0 {
		model = model.Where("category_id", filter.CategoryID)
	}
	return model
}

func (s *sProductCatalog) hydrateProductRecords(ctx context.Context, records []productRecord) error {
	for index := range records {
		var supplier struct {
			CompanyName string `orm:"company_name"`
		}
		if err := dao.Users.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("company_name").Where("id", records[index].SupplierUserID).Scan(&supplier); err != nil {
			return gerror.Wrap(err, "query buyer product supplier")
		}
		var category struct{ Name string }
		if err := dao.Categories.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("name").Where("id", records[index].CategoryID).Scan(&category); err != nil {
			return gerror.Wrap(err, "query buyer product category")
		}
		var unit struct {
			Code string
			Name string
		}
		if err := dao.StockUnits.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("code", "name").Where("id", records[index].StockUnitID).Scan(&unit); err != nil {
			return gerror.Wrap(err, "query buyer product stock unit")
		}
		records[index].SupplierCompany = supplier.CompanyName
		records[index].CategoryName = category.Name
		records[index].StockUnitCode = unit.Code
		records[index].StockUnitName = unit.Name
	}
	return nil
}

func (s *sProductCatalog) favoriteProductIDs(ctx context.Context, buyerUserID uint64, records []productRecord) (map[uint64]bool, error) {
	result := make(map[uint64]bool, len(records))
	if len(records) == 0 {
		return result, nil
	}
	ids := make([]uint64, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.ID)
	}
	var favoriteRows []struct {
		ProductID uint64 `orm:"product_id"`
	}
	if err := dao.ProductFavorites.Ctx(ctx).
		Fields("product_id").Where("user_id", buyerUserID).WhereIn("product_id", ids).Scan(&favoriteRows); err != nil {
		return nil, gerror.Wrap(err, "query buyer product favorites")
	}
	for _, row := range favoriteRows {
		result[row.ProductID] = true
	}
	return result, nil
}

type productRecord struct {
	ID              uint64
	SupplierUserID  uint64
	SupplierCompany string
	CategoryID      uint64
	CategoryName    string
	Name            string
	StockQuantity   float64
	StockUnitID     uint64
	StockUnitCode   string
	StockUnitName   string
	Status          uint8
	CreatedAt       time.Time
}

func (r productRecord) item() productCatalogModel.ProductItem {
	return productCatalogModel.ProductItem{
		ID: r.ID, SupplierUserID: r.SupplierUserID, SupplierCompany: r.SupplierCompany,
		CategoryID: r.CategoryID, CategoryName: r.CategoryName, Name: r.Name,
		StockQuantity: r.StockQuantity, StockUnitID: r.StockUnitID, StockUnitCode: r.StockUnitCode,
		StockUnitName: r.StockUnitName, Status: r.Status, CreatedAt: r.CreatedAt,
	}
}

type imageRecord struct {
	ProductID  uint64 `orm:"product_id"`
	Endpoint   string
	Bucket     string
	UserFileID uint64 `orm:"user_file_id"`
	ObjectKey  string `orm:"object_key"`
}

type productMedia struct {
	Covers map[uint64]imageRecord
	Counts map[uint64]uint
}

func (s *sProductCatalog) loadProductMedia(ctx context.Context, records []productRecord) (productMedia, error) {
	media := productMedia{Covers: make(map[uint64]imageRecord), Counts: make(map[uint64]uint)}
	if len(records) == 0 {
		return media, nil
	}
	ids := make([]uint64, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.ID)
	}
	images, err := loadProductImages(ctx, ids)
	if err != nil {
		return media, gerror.Wrap(err, "query buyer product media")
	}
	for _, image := range images {
		media.Counts[image.ProductID]++
		if _, exists := media.Covers[image.ProductID]; !exists {
			media.Covers[image.ProductID] = image
		}
	}
	return media, nil
}

func (s *sProductCatalog) loadCover(ctx context.Context, productID uint64) (imageRecord, error) {
	images, err := loadProductImages(ctx, []uint64{productID})
	if err != nil {
		return imageRecord{}, gerror.Wrap(err, "query buyer product cover")
	}
	if len(images) == 0 {
		return imageRecord{}, nil
	}
	return images[0], nil
}

func loadProductImages(ctx context.Context, productIDs []uint64) ([]imageRecord, error) {
	if len(productIDs) == 0 {
		return []imageRecord{}, nil
	}
	var images []imageRecord
	if err := dao.ProductImages.Ctx(ctx).
		Fields("product_id", "id", "user_file_id", "sort_order").
		WhereIn("product_id", productIDs).
		OrderAsc("product_id").OrderAsc("sort_order").OrderAsc("id").Scan(&images); err != nil {
		return nil, err
	}
	fileIDs := make([]uint64, 0, len(images))
	seen := make(map[uint64]struct{}, len(images))
	for _, image := range images {
		if image.UserFileID != 0 {
			if _, ok := seen[image.UserFileID]; !ok {
				seen[image.UserFileID] = struct{}{}
				fileIDs = append(fileIDs, image.UserFileID)
			}
		}
	}
	if len(fileIDs) == 0 {
		return images, nil
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
	for index := range images {
		if file, ok := fileByID[images[index].UserFileID]; ok {
			images[index].Endpoint, images[index].Bucket, images[index].ObjectKey = file.Endpoint, file.Bucket, file.ObjectKey
		}
	}
	return images, nil
}

func signImageURL(client *storage.Client, image imageRecord) string {
	if client == nil || image.Endpoint == "" || image.Bucket == "" || image.ObjectKey == "" {
		return ""
	}
	url, err := client.SignedURLFor(image.Endpoint, image.Bucket, image.ObjectKey, 30*time.Minute)
	if err != nil {
		return ""
	}
	return url
}
