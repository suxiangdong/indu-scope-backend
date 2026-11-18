package service

import (
	"context"

	adminProductModel "github.com/suxiangdong/indu-scope-backend/internal/model/adminproduct"
)

type IAdminProducts interface {
	List(ctx context.Context, filter adminProductModel.ListFilter) (adminProductModel.ListResult, error)
	Detail(ctx context.Context, id uint64) (adminProductModel.ProductItem, error)
	UpdateStatus(ctx context.Context, id uint64, status uint8) (adminProductModel.ProductItem, error)
	Review(ctx context.Context, id uint64) (adminProductModel.ProductItem, error)
}

var localAdminProducts IAdminProducts

func AdminProducts() IAdminProducts {
	if localAdminProducts == nil {
		panic("implement not found for interface IAdminProducts, forgot register?")
	}
	return localAdminProducts
}

func RegisterAdminProducts(i IAdminProducts) {
	localAdminProducts = i
}
