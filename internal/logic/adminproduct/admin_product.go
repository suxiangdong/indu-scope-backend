package adminproduct

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	"github.com/suxiangdong/indu-scope-backend/internal/dao"
	adminProductModel "github.com/suxiangdong/indu-scope-backend/internal/model/adminproduct"
	"github.com/suxiangdong/indu-scope-backend/internal/model/do"
	"github.com/suxiangdong/indu-scope-backend/internal/model/entity"
	"github.com/suxiangdong/indu-scope-backend/internal/pagination"
	"github.com/suxiangdong/indu-scope-backend/internal/productstatus"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
	"github.com/suxiangdong/indu-scope-backend/internal/storage"
)

type sAdminProducts struct{}

var _ service.IAdminProducts = (*sAdminProducts)(nil)

func New() service.IAdminProducts { return &sAdminProducts{} }

func init() { service.RegisterAdminProducts(New()) }

func (s *sAdminProducts) List(ctx context.Context, filter adminProductModel.ListFilter) (adminProductModel.ListResult, error) {
	filter.Page, filter.PageSize = pagination.Normalize(filter.Page, filter.PageSize)
	total, err := s.productQuery(ctx, filter).Count()
	if err != nil {
		return adminProductModel.ListResult{}, gerror.Wrap(err, "count admin products")
	}

	var records []productRecord
	offset := (filter.Page - 1) * filter.PageSize
	if err := s.productQuery(ctx, filter).Fields("id", "supplier_user_id", "category_id", "name", "stock_quantity", "stock_unit_id", "status", "created_at", "updated_at").OrderDesc("created_at").OrderDesc("id").Limit(int(offset), int(filter.PageSize)).Scan(&records); err != nil {
		return adminProductModel.ListResult{}, gerror.Wrap(err, "list admin products")
	}
	if err := s.hydrateProductRecords(ctx, records); err != nil {
		return adminProductModel.ListResult{}, err
	}
	media, err := s.loadProductMedia(ctx, records)
	if err != nil {
		return adminProductModel.ListResult{}, err
	}

	ossClient, _ := storage.OSSClient()
	items := make([]adminProductModel.ProductItem, 0, len(records))
	for _, record := range records {
		item := record.item()
		item.ImageCount = media.Counts[record.ID]
		if image, ok := media.Covers[record.ID]; ok && image.ObjectKey != "" {
			item.CoverURL = signImageURL(ossClient, image)
		}
		items = append(items, item)
	}
	return adminProductModel.ListResult{Items: items, Total: uint64(total), Page: filter.Page, PageSize: filter.PageSize}, nil
}

func (s *sAdminProducts) Detail(ctx context.Context, id uint64) (adminProductModel.ProductItem, error) {
	if id == 0 {
		return adminProductModel.ProductItem{}, &consts.InvalidParameterError{Message: "id is required"}
	}
	var product entity.Products
	if err := dao.Products.Ctx(ctx).Cache(gdb.CacheOption{}).Where("id", id).Scan(&product); err != nil {
		return adminProductModel.ProductItem{}, gerror.Wrap(err, "query product detail")
	}
	if product.Id == 0 {
		return adminProductModel.ProductItem{}, consts.ErrProductNotFound
	}
	var supplier struct {
		CompanyName string `orm:"company_name"`
	}
	if err := dao.Users.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("company_name").Where(do.Users{Id: product.SupplierUserId}).Scan(&supplier); err != nil {
		return adminProductModel.ProductItem{}, gerror.Wrap(err, "query product supplier")
	}
	var category struct{ Name string }
	if err := dao.Categories.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("name").Where(do.Categories{Id: product.CategoryId}).Scan(&category); err != nil {
		return adminProductModel.ProductItem{}, gerror.Wrap(err, "query product category")
	}
	var unit struct{ Name string }
	if err := dao.StockUnits.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("name").Where("id", product.StockUnitId).Scan(&unit); err != nil {
		return adminProductModel.ProductItem{}, gerror.Wrap(err, "query product stock unit")
	}
	record := productRecord{ID: product.Id, SupplierUserID: product.SupplierUserId, SupplierCompany: supplier.CompanyName,
		CategoryID: product.CategoryId, CategoryName: category.Name, Name: product.Name, StockQuantity: product.StockQuantity,
		StockUnitID: product.StockUnitId, StockUnitName: unit.Name, Status: uint8(product.Status), CreatedAt: product.CreatedAt, UpdatedAt: product.UpdatedAt}
	item := record.item()
	ossClient, _ := storage.OSSClient()
	images, err := s.loadImages(ctx, ossClient, id)
	if err != nil {
		return adminProductModel.ProductItem{}, err
	}
	item.Images = images
	item.ImageCount = uint(len(images))
	if len(images) > 0 {
		item.CoverURL = images[0].URL
	}
	attributes, err := s.loadAttributes(ctx, id)
	if err != nil {
		return adminProductModel.ProductItem{}, err
	}
	item.Attributes = attributes
	return item, nil
}

func (s *sAdminProducts) UpdateStatus(ctx context.Context, id uint64, status uint8) (adminProductModel.ProductItem, error) {
	if id == 0 {
		return adminProductModel.ProductItem{}, &consts.InvalidParameterError{Message: "id is required"}
	}
	if err := productstatus.Change(dao.Products.Ctx(ctx).Cache(gdb.CacheOption{}).Where("id", id), status); err != nil {
		return adminProductModel.ProductItem{}, err
	}
	return s.Detail(ctx, id)
}

func (s *sAdminProducts) Review(ctx context.Context, id uint64) (adminProductModel.ProductItem, error) {
	if id == 0 {
		return adminProductModel.ProductItem{}, &consts.InvalidParameterError{Message: "id is required"}
	}
	if err := productstatus.Approve(dao.Products.Ctx(ctx).Cache(gdb.CacheOption{}).Where("id", id)); err != nil {
		return adminProductModel.ProductItem{}, err
	}
	return s.Detail(ctx, id)
}

func (s *sAdminProducts) productQuery(ctx context.Context, filter adminProductModel.ListFilter) *gdb.Model {
	model := dao.Products.Ctx(ctx)
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		pattern := "%" + keyword + "%"
		model = model.Where("(name LIKE ? OR supplier_user_id IN (SELECT id FROM users WHERE company_name LIKE ?))", pattern, pattern)
	}
	if filter.CategoryID != nil && *filter.CategoryID != 0 {
		model = model.Where("category_id", *filter.CategoryID)
	}
	if filter.SupplierUserID != nil && *filter.SupplierUserID != 0 {
		model = model.Where("supplier_user_id", *filter.SupplierUserID)
	}
	if filter.Status != nil {
		model = model.Where("status", *filter.Status)
	}
	return model
}

func (s *sAdminProducts) hydrateProductRecords(ctx context.Context, records []productRecord) error {
	for index := range records {
		var supplier struct {
			CompanyName string `orm:"company_name"`
		}
		if err := dao.Users.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("company_name").Where("id", records[index].SupplierUserID).Scan(&supplier); err != nil {
			return gerror.Wrap(err, "query admin product supplier")
		}
		var category struct{ Name string }
		if err := dao.Categories.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("name").Where("id", records[index].CategoryID).Scan(&category); err != nil {
			return gerror.Wrap(err, "query admin product category")
		}
		var unit struct{ Name string }
		if err := dao.StockUnits.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("name").Where("id", records[index].StockUnitID).Scan(&unit); err != nil {
			return gerror.Wrap(err, "query admin product stock unit")
		}
		records[index].SupplierCompany = supplier.CompanyName
		records[index].CategoryName = category.Name
		records[index].StockUnitName = unit.Name
	}
	return nil
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
	StockUnitName   string
	Status          uint8
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (r productRecord) item() adminProductModel.ProductItem {
	return adminProductModel.ProductItem{
		ID: r.ID, SupplierUserID: r.SupplierUserID, SupplierCompany: r.SupplierCompany,
		CategoryID: r.CategoryID, CategoryName: r.CategoryName, Name: r.Name,
		StockQuantity: r.StockQuantity, StockUnitID: r.StockUnitID, StockUnitName: r.StockUnitName,
		Status: r.Status, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		Images: []adminProductModel.ProductImage{}, Attributes: []adminProductModel.AttributeValue{},
	}
}

type imageRecord struct {
	ProductID    uint64 `orm:"product_id"`
	ID           uint64
	UserFileID   uint64 `orm:"user_file_id"`
	Endpoint     string
	Bucket       string
	ObjectKey    string `orm:"object_key"`
	OriginalName string `orm:"original_name"`
	ContentType  string `orm:"content_type"`
	FileSize     int64  `orm:"file_size"`
	SortOrder    uint   `orm:"sort_order"`
	CreatedAt    time.Time
}

type productMedia struct {
	Covers map[uint64]imageRecord
	Counts map[uint64]uint
}

func (s *sAdminProducts) loadProductMedia(ctx context.Context, records []productRecord) (productMedia, error) {
	media := productMedia{Covers: make(map[uint64]imageRecord), Counts: make(map[uint64]uint)}
	if len(records) == 0 {
		return media, nil
	}
	ids := make([]uint64, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.ID)
	}
	images, err := s.loadImageRecords(ctx, ids)
	if err != nil {
		return media, gerror.Wrap(err, "query admin product media")
	}
	for _, image := range images {
		media.Counts[image.ProductID]++
		if _, exists := media.Covers[image.ProductID]; !exists {
			media.Covers[image.ProductID] = image
		}
	}
	return media, nil
}

func (s *sAdminProducts) loadFirstImage(ctx context.Context, id uint64) (imageRecord, error) {
	images, err := s.loadImageRecords(ctx, []uint64{id})
	if err != nil {
		return imageRecord{}, gerror.Wrap(err, "query product cover")
	}
	if len(images) == 0 {
		return imageRecord{}, nil
	}
	return images[0], nil
}

func (s *sAdminProducts) loadImages(ctx context.Context, ossClient *storage.Client, id uint64) ([]adminProductModel.ProductImage, error) {
	records, err := s.loadImageRecords(ctx, []uint64{id})
	if err != nil {
		return nil, gerror.Wrap(err, "query product images")
	}
	images := make([]adminProductModel.ProductImage, 0, len(records))
	for _, record := range records {
		images = append(images, adminProductModel.ProductImage{
			ID: record.ID, URL: signImageURL(ossClient, record), Path: record.ObjectKey,
			OriginalName: record.OriginalName, ContentType: record.ContentType, Size: record.FileSize, SortOrder: record.SortOrder, CreatedAt: record.CreatedAt,
		})
	}
	return images, nil
}

func (s *sAdminProducts) loadImageRecords(ctx context.Context, productIDs []uint64) ([]imageRecord, error) {
	if len(productIDs) == 0 {
		return []imageRecord{}, nil
	}
	var rows []imageRecord
	if err := dao.ProductImages.Ctx(ctx).
		Fields("id", "product_id", "user_file_id", "sort_order", "created_at").
		WhereIn("product_id", productIDs).
		OrderAsc("product_id").OrderAsc("sort_order").OrderAsc("id").Scan(&rows); err != nil {
		return nil, err
	}
	fileIDs := make([]uint64, 0, len(rows))
	seen := make(map[uint64]struct{}, len(rows))
	for _, row := range rows {
		if row.UserFileID != 0 {
			if _, ok := seen[row.UserFileID]; !ok {
				seen[row.UserFileID] = struct{}{}
				fileIDs = append(fileIDs, row.UserFileID)
			}
		}
	}
	if len(fileIDs) == 0 {
		return rows, nil
	}
	var files []struct {
		ID           uint64
		Endpoint     string
		Bucket       string
		ObjectKey    string `orm:"object_key"`
		OriginalName string `orm:"original_name"`
		ContentType  string `orm:"content_type"`
		FileSize     int64  `orm:"file_size"`
	}
	if err := dao.UserFiles.Ctx(ctx).WhereIn("id", fileIDs).
		Fields("id", "endpoint", "bucket", "object_key", "original_name", "content_type", "file_size").Scan(&files); err != nil {
		return nil, err
	}
	fileByID := make(map[uint64]struct {
		Endpoint, Bucket, ObjectKey, OriginalName, ContentType string
		FileSize                                               int64
	}, len(files))
	for _, file := range files {
		fileByID[file.ID] = struct {
			Endpoint, Bucket, ObjectKey, OriginalName, ContentType string
			FileSize                                               int64
		}{file.Endpoint, file.Bucket, file.ObjectKey, file.OriginalName, file.ContentType, file.FileSize}
	}
	for index := range rows {
		if file, ok := fileByID[rows[index].UserFileID]; ok {
			rows[index].Endpoint = file.Endpoint
			rows[index].Bucket = file.Bucket
			rows[index].ObjectKey = file.ObjectKey
			rows[index].OriginalName = file.OriginalName
			rows[index].ContentType = file.ContentType
			rows[index].FileSize = file.FileSize
		}
	}
	return rows, nil
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

type attributeRecord struct {
	ID            uint64
	AttributeID   uint64
	Code          string
	Name          string
	Unit          string
	ValueNumber   *float64 `orm:"value_number"`
	ValueText     *string  `orm:"value_text"`
	ValueBool     *uint    `orm:"value_bool"`
	ValueOptionID *uint64  `orm:"value_option_id"`
	ValueJSON     *string  `orm:"value_json"`
}

func (s *sAdminProducts) loadAttributes(ctx context.Context, productID uint64) ([]adminProductModel.AttributeValue, error) {
	var records []attributeRecord
	if err := dao.ProductAttributeValues.Ctx(ctx).
		Fields("id", "attribute_id", "value_number", "value_text", "value_bool", "value_option_id", "value_json").
		Where("product_id", productID).OrderAsc("id").Scan(&records); err != nil {
		return nil, gerror.Wrap(err, "query product attributes")
	}
	attributeIDs := make([]uint64, 0, len(records))
	seen := make(map[uint64]struct{}, len(records))
	for _, record := range records {
		if _, ok := seen[record.AttributeID]; !ok {
			seen[record.AttributeID] = struct{}{}
			attributeIDs = append(attributeIDs, record.AttributeID)
		}
	}
	var attributes []struct {
		ID   uint64
		Code string
		Name string
		Unit string
	}
	if len(attributeIDs) > 0 {
		if err := dao.Attributes.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("id", "code", "name", "unit").WhereIn("id", attributeIDs).Scan(&attributes); err != nil {
			return nil, gerror.Wrap(err, "query product attribute definitions")
		}
	}
	attributeByID := make(map[uint64]struct{ Code, Name, Unit string }, len(attributes))
	for _, attribute := range attributes {
		attributeByID[attribute.ID] = struct{ Code, Name, Unit string }{attribute.Code, attribute.Name, attribute.Unit}
	}
	items := make([]adminProductModel.AttributeValue, 0, len(records))
	for _, record := range records {
		definition := attributeByID[record.AttributeID]
		record.Code, record.Name, record.Unit = definition.Code, definition.Name, definition.Unit
		value, err := s.attributeValue(ctx, record)
		if err != nil {
			return nil, err
		}
		items = append(items, adminProductModel.AttributeValue{AttributeID: record.AttributeID, Code: record.Code, Name: record.Name, Unit: record.Unit, Value: value})
	}
	return items, nil
}

func (s *sAdminProducts) attributeValue(ctx context.Context, record attributeRecord) (string, error) {
	if record.ValueOptionID != nil && *record.ValueOptionID != 0 {
		var option struct{ Name string }
		if err := dao.AttributeOptions.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("name").Where("id", *record.ValueOptionID).Scan(&option); err != nil {
			return "", gerror.Wrap(err, "query product attribute option")
		}
		if option.Name != "" {
			return option.Name, nil
		}
	}
	var optionRows []struct {
		OptionID uint64 `orm:"option_id"`
	}
	if err := dao.ProductAttributeValueOptions.Ctx(ctx).Fields("option_id").Where("value_id", record.ID).OrderAsc("option_id").Scan(&optionRows); err != nil {
		return "", gerror.Wrap(err, "query product attribute option ids")
	}
	optionIDs := make([]uint64, 0, len(optionRows))
	for _, row := range optionRows {
		optionIDs = append(optionIDs, row.OptionID)
	}
	var options []struct {
		ID   uint64
		Name string
	}
	if len(optionIDs) > 0 {
		if err := dao.AttributeOptions.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("id", "name").WhereIn("id", optionIDs).Scan(&options); err != nil {
			return "", gerror.Wrap(err, "query product attribute options")
		}
	}
	if len(options) > 0 {
		namesByID := make(map[uint64]string, len(options))
		for _, option := range options {
			namesByID[option.ID] = option.Name
		}
		names := make([]string, 0, len(optionIDs))
		for _, optionID := range optionIDs {
			option := struct{ Name string }{Name: namesByID[optionID]}
			if option.Name != "" {
				names = append(names, option.Name)
			}
		}
		if len(names) > 0 {
			return strings.Join(names, ", "), nil
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
