package internal

import "github.com/gogf/gf/v2/database/gdb"

type HotCategoriesDao struct{ *TableDao }

func NewHotCategoriesDao(handlers ...gdb.ModelHandler) *HotCategoriesDao {
	return &HotCategoriesDao{TableDao: NewTableDao("hot_categories", handlers...)}
}
