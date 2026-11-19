package supplierproducts

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/suxiangdong/indu-scope-backend/internal/attributevalue"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	"github.com/suxiangdong/indu-scope-backend/internal/dao"
	"github.com/suxiangdong/indu-scope-backend/internal/model/do"
	"github.com/suxiangdong/indu-scope-backend/internal/model/entity"
	supplierProductModel "github.com/suxiangdong/indu-scope-backend/internal/model/supplierproduct"
	"github.com/suxiangdong/indu-scope-backend/internal/pagination"
	"github.com/suxiangdong/indu-scope-backend/internal/productstatus"
	"github.com/suxiangdong/indu-scope-backend/internal/selectcache"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
	"github.com/suxiangdong/indu-scope-backend/internal/storage"
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
	result := supplierProductModel.FormOptions{Categories: categories, Attributes: []supplierProductModel.AttributeItem{}, StockUnits: []supplierProductModel.StockUnitItem{}}
	if categoryID == nil || *categoryID == 0 {
		units, err := s.listStockUnits(ctx, nil)
		if err != nil {
			return supplierProductModel.FormOptions{}, err
		}
		result.StockUnits = units
		return result, nil
	}
	if err := s.ensureLeafCategory(ctx, *categoryID); err != nil {
		return supplierProductModel.FormOptions{}, err
	}
	units, err := s.listStockUnits(ctx, categoryID)
	if err != nil {
		return supplierProductModel.FormOptions{}, err
	}
	result.StockUnits = units

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

func (s *sSupplierProducts) List(ctx context.Context, supplierUserID uint64, page, pageSize uint) (supplierProductModel.ListResult, error) {
	if supplierUserID == 0 {
		return supplierProductModel.ListResult{}, consts.ErrSupplierRequired
	}
	page, pageSize = pagination.Normalize(page, pageSize)
	base := dao.Products.Ctx(ctx).Where(do.Products{SupplierUserId: supplierUserID})
	total, err := base.Count()
	if err != nil {
		return supplierProductModel.ListResult{}, gerror.Wrap(err, "count supplier products")
	}
	var records []productListRecord
	offset := (page - 1) * pageSize
	if err := dao.Products.Ctx(ctx).
		Fields("id", "supplier_user_id", "category_id", "name", "stock_quantity", "stock_unit_id", "status", "created_at").
		Where("supplier_user_id", supplierUserID).
		OrderDesc("created_at").OrderDesc("id").
		Limit(int(offset), int(pageSize)).Scan(&records); err != nil {
		return supplierProductModel.ListResult{}, gerror.Wrap(err, "list supplier products")
	}
	if err := s.hydrateProductRecordNames(ctx, records); err != nil {
		return supplierProductModel.ListResult{}, err
	}
	media, err := s.loadProductMedia(ctx, records)
	if err != nil {
		return supplierProductModel.ListResult{}, err
	}
	ossClient, _ := storage.OSSClient()
	items := make([]supplierProductModel.ProductItem, 0, len(records))
	for _, record := range records {
		item := record.item()
		item.ImageCount = media.Counts[record.ID]
		if cover, ok := media.Covers[record.ID]; ok && cover.ObjectKey != "" {
			item.CoverURL = signProductImageURL(ossClient, cover)
		}
		items = append(items, item)
	}
	return supplierProductModel.ListResult{Items: items, Total: uint64(total), Page: page, PageSize: pageSize}, nil
}

func (s *sSupplierProducts) Detail(ctx context.Context, supplierUserID, productID uint64) (supplierProductModel.ProductItem, error) {
	if supplierUserID == 0 {
		return supplierProductModel.ProductItem{}, consts.ErrSupplierRequired
	}
	if productID == 0 {
		return supplierProductModel.ProductItem{}, &consts.InvalidParameterError{Message: "id is required"}
	}
	var product entity.Products
	if err := dao.Products.Ctx(ctx).Cache(selectcache.Query(dao.Products.Table(), productID)).
		Where("id", productID).Scan(&product); err != nil {
		return supplierProductModel.ProductItem{}, gerror.Wrap(err, "query supplier product detail")
	}
	if product.Id == 0 || product.SupplierUserId != supplierUserID {
		return supplierProductModel.ProductItem{}, consts.ErrProductNotFound
	}
	var category struct{ Name string }
	if err := dao.Categories.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("name").Where(do.Categories{Id: product.CategoryId}).Scan(&category); err != nil {
		return supplierProductModel.ProductItem{}, gerror.Wrap(err, "query supplier product category")
	}
	var unit stockUnitRecord
	if err := dao.StockUnits.Ctx(ctx).Cache(gdb.CacheOption{}).
		Fields("id", "code", "name").Where("id", product.StockUnitId).Scan(&unit); err != nil {
		return supplierProductModel.ProductItem{}, gerror.Wrap(err, "query supplier product stock unit")
	}
	record := productListRecord{ID: product.Id, SupplierUserID: product.SupplierUserId, CategoryID: product.CategoryId, Name: product.Name,
		StockQuantity: product.StockQuantity, StockUnitID: product.StockUnitId, StockUnitCode: unit.Code, StockUnitName: unit.Name,
		Status: product.Status, CreatedAt: product.CreatedAt, CategoryName: category.Name}
	adminItem, err := service.AdminProducts().Detail(ctx, productID)
	if err != nil {
		return supplierProductModel.ProductItem{}, err
	}
	item := record.item()
	supplierProfile, err := service.UserProfile().Profile(ctx, product.SupplierUserId)
	if err != nil {
		return supplierProductModel.ProductItem{}, gerror.Wrap(err, "query supplier product contacts")
	}
	item.Contacts = supplierProfile.Contacts
	item.ImageCount = uint(len(adminItem.Images))
	item.Images = make([]supplierProductModel.ProductImageItem, 0, len(adminItem.Images))
	for _, image := range adminItem.Images {
		item.Images = append(item.Images, supplierProductModel.ProductImageItem{ID: image.ID, Path: image.Path, URL: image.URL, OriginalName: image.OriginalName, ContentType: image.ContentType, Size: image.Size, SortOrder: image.SortOrder, CreatedAt: image.CreatedAt})
	}
	if len(item.Images) > 0 {
		item.CoverURL = item.Images[0].URL
	}
	item.Attributes = make([]supplierProductModel.AttributeValueItem, 0, len(adminItem.Attributes))
	for _, attribute := range adminItem.Attributes {
		item.Attributes = append(item.Attributes, supplierProductModel.AttributeValueItem{AttributeID: attribute.AttributeID, Code: attribute.Code, Name: attribute.Name, Unit: attribute.Unit, Value: attribute.Value})
	}
	return item, nil
}

func (s *sSupplierProducts) Create(ctx context.Context, input supplierProductModel.CreateInput) (supplierProductModel.ProductItem, error) {
	if input.SupplierUserID == 0 {
		return supplierProductModel.ProductItem{}, consts.ErrSupplierRequired
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.CategoryID == 0 {
		return supplierProductModel.ProductItem{}, &consts.InvalidParameterError{Message: "category_id is required"}
	}
	if input.Name == "" || len([]rune(input.Name)) > 255 {
		return supplierProductModel.ProductItem{}, &consts.InvalidParameterError{Message: "name is required and must be no longer than 255 characters"}
	}
	if input.StockQuantity < 0 || math.IsNaN(input.StockQuantity) || math.IsInf(input.StockQuantity, 0) {
		return supplierProductModel.ProductItem{}, &consts.InvalidParameterError{Message: "stock_quantity must be greater than or equal to 0"}
	}
	if input.StockUnitID == 0 {
		return supplierProductModel.ProductItem{}, &consts.InvalidParameterError{Message: "stock_unit_id is required"}
	}
	if err := s.ensureLeafCategory(ctx, input.CategoryID); err != nil {
		return supplierProductModel.ProductItem{}, err
	}
	stockUnit, err := s.loadStockUnitForCategory(ctx, input.CategoryID, input.StockUnitID)
	if err != nil {
		return supplierProductModel.ProductItem{}, err
	}
	pendingImages, err := s.loadPendingImages(ctx, input.SupplierUserID, input.ImagePaths)
	if err != nil {
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
		if err := attributevalue.Validate(definition, value); err != nil {
			return supplierProductModel.ProductItem{}, err
		}
	}

	var product supplierProductModel.ProductItem
	err = dao.Products.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		id, err := tx.Model(dao.Products.Table()).Data(g.Map{
			"supplier_user_id": input.SupplierUserID,
			"category_id":      input.CategoryID,
			"name":             input.Name,
			"stock_unit_id":    input.StockUnitID,
			"stock_quantity":   input.StockQuantity,
			"stock_updated_at": time.Now(),
			"stock_updated_by": input.SupplierUserID,
			"status":           0,
		}).InsertAndGetId()
		if err != nil {
			return gerror.Wrap(err, "create product")
		}
		product = supplierProductModel.ProductItem{
			ID: uint64(id), SupplierUserID: input.SupplierUserID, CategoryID: input.CategoryID,
			Name: input.Name, StockQuantity: input.StockQuantity, StockUnitID: input.StockUnitID,
			StockUnitCode: stockUnit.Code, StockUnitName: stockUnit.Name, ImageCount: uint(len(pendingImages)), CreatedAt: time.Now(), Status: 0,
			Images: make([]supplierProductModel.ProductImageItem, 0, len(pendingImages)),
		}
		for index, image := range pendingImages {
			imageID, err := tx.Model(dao.ProductImages.Table()).Data(g.Map{
				"product_id": id, "user_file_id": image.ID, "sort_order": index,
			}).InsertAndGetId()
			if err != nil {
				return gerror.Wrap(err, "create product image")
			}
			product.Images = append(product.Images, supplierProductModel.ProductImageItem{
				ID: uint64(imageID), Endpoint: image.Endpoint, Bucket: image.Bucket, Path: image.ObjectKey,
				OriginalName: image.OriginalName, ContentType: image.ContentType, Size: image.FileSize,
				SortOrder: uint(index), CreatedAt: time.Now(),
			})
		}
		for _, value := range values {
			definition := definitions[value.AttributeID]
			if value.Value == "" && value.OptionID == 0 && len(value.OptionIDs) == 0 {
				continue
			}
			data, optionIDs, err := attributevalue.Data(definition, value)
			if err != nil {
				return err
			}
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

func (s *sSupplierProducts) UpdateStock(ctx context.Context, input supplierProductModel.UpdateStockInput) (supplierProductModel.ProductItem, error) {
	if input.SupplierUserID == 0 {
		return supplierProductModel.ProductItem{}, consts.ErrSupplierRequired
	}
	if input.ProductID == 0 {
		return supplierProductModel.ProductItem{}, &consts.InvalidParameterError{Message: "id is required"}
	}
	if input.StockQuantity < 0 || math.IsNaN(input.StockQuantity) || math.IsInf(input.StockQuantity, 0) {
		return supplierProductModel.ProductItem{}, &consts.InvalidParameterError{Message: "stock_quantity must be greater than or equal to 0"}
	}
	var current struct{ ID uint64 }
	if err := dao.Products.Ctx(ctx).Fields("id").Where(do.Products{Id: input.ProductID, SupplierUserId: input.SupplierUserID}).Scan(&current); err != nil {
		return supplierProductModel.ProductItem{}, gerror.Wrap(err, "query supplier product for stock update")
	}
	if current.ID == 0 {
		return supplierProductModel.ProductItem{}, consts.ErrProductNotFound
	}
	if _, err := dao.Products.Ctx(ctx).Cache(selectcache.Invalidate(dao.Products.Table(), input.ProductID)).Where(do.Products{Id: input.ProductID, SupplierUserId: input.SupplierUserID}).Data(g.Map{
		"stock_quantity":   input.StockQuantity,
		"stock_updated_at": time.Now(),
		"stock_updated_by": input.SupplierUserID,
	}).Update(); err != nil {
		return supplierProductModel.ProductItem{}, gerror.Wrap(err, "update supplier product stock")
	}
	return s.Detail(ctx, input.SupplierUserID, input.ProductID)
}

func (s *sSupplierProducts) UpdateStatus(ctx context.Context, supplierUserID, productID uint64, status uint8) (supplierProductModel.ProductItem, error) {
	if supplierUserID == 0 {
		return supplierProductModel.ProductItem{}, consts.ErrSupplierRequired
	}
	if productID == 0 {
		return supplierProductModel.ProductItem{}, &consts.InvalidParameterError{Message: "id is required"}
	}
	if status != 1 && status != 2 {
		return supplierProductModel.ProductItem{}, &consts.InvalidParameterError{Message: "supplier product status must be 1 or 2"}
	}
	if err := productstatus.Change(dao.Products.Ctx(ctx).
		Where(do.Products{Id: productID, SupplierUserId: supplierUserID}), status); err != nil {
		return supplierProductModel.ProductItem{}, err
	}
	return s.Detail(ctx, supplierUserID, productID)
}

const (
	productImagePurpose = "product_image"
	maxProductImageSize = 10 * 1024 * 1024
	maxProductImages    = 10
)

func (s *sSupplierProducts) UploadImage(ctx context.Context, input supplierProductModel.ProductImageUploadInput) (supplierProductModel.ProductImageUploadResult, error) {
	value := service.Bizctx().Get(ctx)
	if value == nil || value.User == nil || value.User.Status != 1 || value.User.Role != 2 {
		return supplierProductModel.ProductImageUploadResult{}, consts.ErrSupplierRequired
	}
	if input.Reader == nil {
		return supplierProductModel.ProductImageUploadResult{}, &consts.InvalidParameterError{Message: "product image is required"}
	}
	if input.Size <= 0 || input.Size > maxProductImageSize {
		return supplierProductModel.ProductImageUploadResult{}, &consts.InvalidParameterError{Message: "product image must be between 1 byte and 10 MB"}
	}
	contentType := strings.ToLower(strings.TrimSpace(input.ContentType))
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		return supplierProductModel.ProductImageUploadResult{}, &consts.InvalidParameterError{Message: "product image must be a JPG, PNG or WEBP image"}
	}
	filename := filepath.Base(strings.TrimSpace(input.Filename))
	if filename == "." || filename == "" {
		filename = "product-image" + extensionForImageContentType(contentType)
	}
	if filepath.Ext(filename) == "" {
		filename += extensionForImageContentType(contentType)
	}
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return supplierProductModel.ProductImageUploadResult{}, gerror.Wrap(err, "generate product image object key")
	}
	client, err := storage.OSSClient()
	if err != nil {
		return supplierProductModel.ProductImageUploadResult{}, gerror.Wrap(err, "initialize oss client")
	}
	objectKey := client.ObjectKey("product/images", fmt.Sprintf("%d-%d-%s%s", value.User.ID, time.Now().UnixNano(), hex.EncodeToString(nonce), extensionForImageContentType(contentType)))
	etag, err := client.Upload(ctx, objectKey, input.Reader, input.Size, contentType)
	if err != nil {
		_ = storage.DeleteWithTimeout(ctx, client, objectKey)
		return supplierProductModel.ProductImageUploadResult{}, gerror.Wrap(err, "upload product image to oss")
	}
	if _, err := dao.UserFiles.Ctx(ctx).Data(g.Map{
		"user_id": value.User.ID, "purpose": productImagePurpose, "endpoint": client.EndpointHost(),
		"bucket": client.Bucket(), "object_key": objectKey, "original_name": filename,
		"content_type": contentType, "file_size": input.Size, "etag": etag,
	}).Insert(); err != nil {
		_ = storage.DeleteWithTimeout(ctx, client, objectKey)
		return supplierProductModel.ProductImageUploadResult{}, gerror.Wrap(err, "save product image metadata")
	}
	return supplierProductModel.ProductImageUploadResult{
		Endpoint: client.EndpointHost(), Bucket: client.Bucket(), Path: objectKey,
		OriginalName: filename, ContentType: contentType, Size: input.Size,
	}, nil
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

type stockUnitRecord struct {
	ID        uint64
	Code      string
	Name      string
	SortOrder uint
	Status    uint8
}

type productListRecord struct {
	ID             uint64
	SupplierUserID uint64
	CategoryID     uint64
	Name           string
	StockQuantity  float64
	StockUnitID    uint64
	StockUnitCode  string
	StockUnitName  string
	Status         uint
	CreatedAt      time.Time
	CategoryName   string
}

func (s *sSupplierProducts) hydrateProductRecordNames(ctx context.Context, records []productListRecord) error {
	if len(records) == 0 {
		return nil
	}
	for index := range records {
		var category struct{ Name string }
		if err := dao.Categories.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("name").Where("id", records[index].CategoryID).Scan(&category); err != nil {
			return gerror.Wrap(err, "query supplier product category")
		}
		var unit struct {
			Code string
			Name string
		}
		if err := dao.StockUnits.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("code", "name").Where("id", records[index].StockUnitID).Scan(&unit); err != nil {
			return gerror.Wrap(err, "query supplier product stock unit")
		}
		records[index].CategoryName = category.Name
		records[index].StockUnitCode = unit.Code
		records[index].StockUnitName = unit.Name
	}
	return nil
}

type productImageRecord struct {
	ProductID  uint64 `orm:"product_id"`
	UserFileID uint64 `orm:"user_file_id"`
	Endpoint   string
	Bucket     string
	ObjectKey  string `orm:"object_key"`
}

type productMedia struct {
	Covers map[uint64]productImageRecord
	Counts map[uint64]uint
}

func (s *sSupplierProducts) loadProductMedia(ctx context.Context, records []productListRecord) (productMedia, error) {
	media := productMedia{Covers: make(map[uint64]productImageRecord), Counts: make(map[uint64]uint)}
	if len(records) == 0 {
		return media, nil
	}
	ids := make([]uint64, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.ID)
	}
	images, err := loadSupplierProductImages(ctx, ids)
	if err != nil {
		return media, gerror.Wrap(err, "query supplier product media")
	}
	for _, image := range images {
		media.Counts[image.ProductID]++
		if _, exists := media.Covers[image.ProductID]; !exists {
			media.Covers[image.ProductID] = image
		}
	}
	return media, nil
}

func loadSupplierProductImages(ctx context.Context, productIDs []uint64) ([]productImageRecord, error) {
	if len(productIDs) == 0 {
		return []productImageRecord{}, nil
	}
	var images []productImageRecord
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

func (r productListRecord) item() supplierProductModel.ProductItem {
	return supplierProductModel.ProductItem{
		ID: r.ID, SupplierUserID: r.SupplierUserID, CategoryID: r.CategoryID,
		Name: r.Name, StockQuantity: r.StockQuantity, StockUnitID: r.StockUnitID,
		StockUnitCode: r.StockUnitCode, StockUnitName: r.StockUnitName,
		Status: r.Status, CreatedAt: r.CreatedAt, CategoryName: r.CategoryName,
		Images: []supplierProductModel.ProductImageItem{}, Attributes: []supplierProductModel.AttributeValueItem{},
	}
}

func signProductImageURL(client *storage.Client, image productImageRecord) string {
	if client == nil || image.Endpoint == "" || image.Bucket == "" || image.ObjectKey == "" {
		return ""
	}
	url, err := client.SignedURLFor(image.Endpoint, image.Bucket, image.ObjectKey, 30*time.Minute)
	if err != nil {
		return ""
	}
	return url
}

func (s *sSupplierProducts) listStockUnits(ctx context.Context, categoryID *uint64) ([]supplierProductModel.StockUnitItem, error) {
	model := dao.StockUnits.Ctx(ctx).Where("status", 1)
	if categoryID != nil && *categoryID != 0 {
		bindingCount, err := dao.CategoryStockUnits.Ctx(ctx).Where("category_id", *categoryID).Count()
		if err != nil {
			return nil, gerror.Wrap(err, "count category stock units")
		}
		if bindingCount > 0 {
			var bindings []struct {
				StockUnitID uint64 `orm:"stock_unit_id"`
			}
			if err := dao.CategoryStockUnits.Ctx(ctx).
				Fields("stock_unit_id").Where("category_id", *categoryID).Scan(&bindings); err != nil {
				return nil, gerror.Wrap(err, "list category stock units")
			}
			unitIDs := make([]uint64, 0, len(bindings))
			for _, binding := range bindings {
				unitIDs = append(unitIDs, binding.StockUnitID)
			}
			if len(unitIDs) == 0 {
				return []supplierProductModel.StockUnitItem{}, nil
			}
			model = model.WhereIn("id", unitIDs)
		}
	}
	var records []stockUnitRecord
	if err := model.Fields("id", "code", "name", "sort_order", "status").
		OrderAsc("sort_order").OrderAsc("id").Scan(&records); err != nil {
		return nil, gerror.Wrap(err, "list stock units")
	}
	items := make([]supplierProductModel.StockUnitItem, 0, len(records))
	for _, record := range records {
		items = append(items, supplierProductModel.StockUnitItem{ID: record.ID, Code: record.Code, Name: record.Name, SortOrder: record.SortOrder})
	}
	return items, nil
}

func (s *sSupplierProducts) loadStockUnitForCategory(ctx context.Context, categoryID, stockUnitID uint64) (stockUnitRecord, error) {
	var record stockUnitRecord
	if err := dao.StockUnits.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("id", "code", "name", "sort_order", "status").
		Where(g.Map{"id": stockUnitID, "status": 1}).Scan(&record); err != nil {
		return stockUnitRecord{}, gerror.Wrap(err, "query stock unit")
	}
	if record.ID == 0 {
		return stockUnitRecord{}, &consts.InvalidParameterError{Message: "stock_unit_id is invalid or disabled"}
	}
	bindingCount, err := dao.CategoryStockUnits.Ctx(ctx).Where("category_id", categoryID).Count()
	if err != nil {
		return stockUnitRecord{}, gerror.Wrap(err, "count category stock units")
	}
	if bindingCount > 0 {
		allowed, err := dao.CategoryStockUnits.Ctx(ctx).Where(g.Map{"category_id": categoryID, "stock_unit_id": stockUnitID}).Count()
		if err != nil {
			return stockUnitRecord{}, gerror.Wrap(err, "check category stock unit")
		}
		if allowed == 0 {
			return stockUnitRecord{}, &consts.InvalidParameterError{Message: "stock_unit_id is not available for the selected category"}
		}
	}
	return record, nil
}

type pendingImageRecord struct {
	ID           uint64
	Endpoint     string
	Bucket       string
	ObjectKey    string `orm:"object_key"`
	OriginalName string `orm:"original_name"`
	ContentType  string `orm:"content_type"`
	FileSize     int64  `orm:"file_size"`
}

func (s *sSupplierProducts) loadPendingImages(ctx context.Context, userID uint64, paths []string) ([]pendingImageRecord, error) {
	if len(paths) > maxProductImages {
		return nil, &consts.InvalidParameterError{Message: "最多上传10张产品图片"}
	}
	images := make([]pendingImageRecord, 0, len(paths))
	seen := make(map[string]struct{}, len(paths))
	for _, rawPath := range paths {
		path := strings.TrimSpace(rawPath)
		if path == "" {
			return nil, &consts.InvalidParameterError{Message: "product image path is invalid"}
		}
		if _, exists := seen[path]; exists {
			return nil, &consts.InvalidParameterError{Message: "product image cannot be duplicated"}
		}
		seen[path] = struct{}{}
		var image pendingImageRecord
		if err := dao.UserFiles.Ctx(ctx).Cache(gdb.CacheOption{}).Fields("id", "endpoint", "bucket", "object_key", "original_name", "content_type", "file_size").
			Where(g.Map{"user_id": userID, "purpose": productImagePurpose, "object_key": path}).Scan(&image); err != nil {
			return nil, gerror.Wrap(err, "query product image metadata")
		}
		if image.ObjectKey == "" {
			return nil, &consts.InvalidParameterError{Message: "product image must be uploaded before submitting"}
		}
		images = append(images, image)
	}
	return images, nil
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
	if err := dao.Attributes.Ctx(ctx).Cache(gdb.CacheOption{}).Fields(cols.Id, cols.Code, cols.Name, cols.ValueType, cols.InputType, cols.Unit, cols.Placeholder, cols.HelpText, cols.Status).
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
