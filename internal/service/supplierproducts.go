package service

import (
	"context"

	supplierProductModel "github.com/suxiangdong/indu-scope-backend/internal/model/supplierproduct"
)

type ISupplierProducts interface {
	FormOptions(ctx context.Context, categoryID *uint64) (supplierProductModel.FormOptions, error)
	Create(ctx context.Context, input supplierProductModel.CreateInput) (supplierProductModel.ProductItem, error)
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
