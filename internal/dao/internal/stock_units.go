package internal

import "github.com/gogf/gf/v2/database/gdb"

type StockUnitsDao struct{ *TableDao }

func NewStockUnitsDao(handlers ...gdb.ModelHandler) *StockUnitsDao {
	return &StockUnitsDao{TableDao: NewTableDao("stock_units", handlers...)}
}
