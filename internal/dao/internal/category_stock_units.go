package internal

import "github.com/gogf/gf/v2/database/gdb"

type CategoryStockUnitsDao struct{ *TableDao }

func NewCategoryStockUnitsDao(handlers ...gdb.ModelHandler) *CategoryStockUnitsDao {
	return &CategoryStockUnitsDao{TableDao: NewTableDao("category_stock_units", handlers...)}
}
