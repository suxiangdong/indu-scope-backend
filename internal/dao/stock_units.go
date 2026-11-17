package dao

import "github.com/suxiangdong/indu-scope-backend/internal/dao/internal"

type stockUnitsDao struct{ *internal.StockUnitsDao }

var StockUnits = stockUnitsDao{internal.NewStockUnitsDao()}
