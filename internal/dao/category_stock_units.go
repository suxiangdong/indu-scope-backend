package dao

import "github.com/suxiangdong/indu-scope-backend/internal/dao/internal"

type categoryStockUnitsDao struct {
	*internal.CategoryStockUnitsDao
}

var CategoryStockUnits = categoryStockUnitsDao{internal.NewCategoryStockUnitsDao()}
