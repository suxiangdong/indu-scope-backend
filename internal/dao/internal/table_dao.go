// ============================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ============================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TableDao contains the common implementation shared by the generated DAO
// objects for tables introduced after the original DAO generation.
type TableDao struct {
	table    string
	group    string
	handlers []gdb.ModelHandler
}

func NewTableDao(table string, handlers ...gdb.ModelHandler) *TableDao {
	return &TableDao{table: table, group: "default", handlers: handlers}
}

func (dao *TableDao) DB() gdb.DB { return g.DB(dao.group) }

func (dao *TableDao) Table() string { return dao.table }

func (dao *TableDao) Group() string { return dao.group }

func (dao *TableDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

func (dao *TableDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) error {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
