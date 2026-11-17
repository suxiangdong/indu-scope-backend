package internal

import "github.com/gogf/gf/v2/database/gdb"

type ProductFavoritesDao struct{ *TableDao }

func NewProductFavoritesDao(handlers ...gdb.ModelHandler) *ProductFavoritesDao {
	return &ProductFavoritesDao{TableDao: NewTableDao("product_favorites", handlers...)}
}
