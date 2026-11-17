package service

import (
	"context"

	stockUnitModel "github.com/suxiangdong/indu-scope-backend/internal/model/stockunit"
)

type IAdminStockUnit interface {
	List(ctx context.Context) (stockUnitModel.ListResult, error)
	Create(ctx context.Context, input stockUnitModel.Input) (stockUnitModel.Item, error)
	Update(ctx context.Context, id uint64, input stockUnitModel.Input) (stockUnitModel.Item, error)
	Delete(ctx context.Context, id uint64) error
	ListCategories(ctx context.Context, stockUnitID uint64) (stockUnitModel.CategoryListResult, error)
	BindCategory(ctx context.Context, stockUnitID, categoryID uint64, sortOrder uint) (stockUnitModel.CategoryItem, error)
	UnbindCategory(ctx context.Context, stockUnitID, categoryID uint64) error
}

var localAdminStockUnit IAdminStockUnit

func AdminStockUnit() IAdminStockUnit {
	if localAdminStockUnit == nil {
		panic("implement not found for interface IAdminStockUnit, forgot register?")
	}
	return localAdminStockUnit
}

func RegisterAdminStockUnit(i IAdminStockUnit) {
	localAdminStockUnit = i
}
