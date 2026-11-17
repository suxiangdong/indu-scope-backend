package internal

import "github.com/gogf/gf/v2/database/gdb"

type AppPageContentsDao struct{ *TableDao }

func NewAppPageContentsDao(handlers ...gdb.ModelHandler) *AppPageContentsDao {
	return &AppPageContentsDao{TableDao: NewTableDao("app_page_contents", handlers...)}
}
