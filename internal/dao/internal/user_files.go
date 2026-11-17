package internal

import "github.com/gogf/gf/v2/database/gdb"

type UserFilesDao struct{ *TableDao }

func NewUserFilesDao(handlers ...gdb.ModelHandler) *UserFilesDao {
	return &UserFilesDao{TableDao: NewTableDao("user_files", handlers...)}
}
