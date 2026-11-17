package internal

import "github.com/gogf/gf/v2/database/gdb"

type ProductImagesDao struct{ *TableDao }

func NewProductImagesDao(handlers ...gdb.ModelHandler) *ProductImagesDao {
	return &ProductImagesDao{TableDao: NewTableDao("product_images", handlers...)}
}
