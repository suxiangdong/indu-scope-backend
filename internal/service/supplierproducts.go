package service

import (
	"context"

	supplierProductModel "github.com/suxiangdong/indu-scope-backend/internal/model/supplierproduct"
)

type ISupplierProducts interface {
	FormOptions(ctx context.Context, categoryID *uint64) (supplierProductModel.FormOptions, error)
	List(ctx context.Context, supplierUserID uint64, page, pageSize uint) (supplierProductModel.ListResult, error)
	Detail(ctx context.Context, supplierUserID, productID uint64) (supplierProductModel.ProductItem, error)
	Create(ctx context.Context, input supplierProductModel.CreateInput) (supplierProductModel.ProductItem, error)
	UpdateStock(ctx context.Context, input supplierProductModel.UpdateStockInput) (supplierProductModel.ProductItem, error)
	UpdateStatus(ctx context.Context, supplierUserID, productID uint64, status uint8) (supplierProductModel.ProductItem, error)
	UploadImage(ctx context.Context, input supplierProductModel.ProductImageUploadInput) (supplierProductModel.ProductImageUploadResult, error)
}

var localSupplierProducts ISupplierProducts

func SupplierProducts() ISupplierProducts {
	if localSupplierProducts == nil {
		panic("implement not found for interface ISupplierProducts, forgot register?")
	}
	return localSupplierProducts
}

func RegisterSupplierProducts(i ISupplierProducts) {
	localSupplierProducts = i
}
