package service

import (
	"context"

	productCatalogModel "github.com/suxiangdong/indu-scope-backend/internal/model/productcatalog"
)

type IProductCatalog interface {
	List(ctx context.Context, buyerUserID uint64, filter productCatalogModel.ListFilter) (productCatalogModel.ListResult, error)
	Detail(ctx context.Context, buyerUserID, productID uint64) (productCatalogModel.ProductDetail, error)
}

var localProductCatalog IProductCatalog

func ProductCatalog() IProductCatalog {
	if localProductCatalog == nil {
		panic("implement not found for interface IProductCatalog, forgot register?")
	}
	return localProductCatalog
}

func RegisterProductCatalog(i IProductCatalog) {
	localProductCatalog = i
}
