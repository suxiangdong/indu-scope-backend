// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

type StorePlatformsDao struct {
	table    string
	group    string
	columns  StorePlatformsColumns
	handlers []gdb.ModelHandler
}

type StorePlatformsColumns struct {
	Id        string
	Name      string
	SortOrder string
	Status    string
	CreatedAt string
	UpdatedAt string
}

var storePlatformsColumns = StorePlatformsColumns{
	Id: "id", Name: "name", SortOrder: "sort_order", Status: "status",
	CreatedAt: "created_at", UpdatedAt: "updated_at",
}

func NewStorePlatformsDao(handlers ...gdb.ModelHandler) *StorePlatformsDao {
	return &StorePlatformsDao{group: "default", table: "store_platforms", columns: storePlatformsColumns, handlers: handlers}
}

func (dao *StorePlatformsDao) DB() gdb.DB                     { return g.DB(dao.group) }
func (dao *StorePlatformsDao) Table() string                  { return dao.table }
func (dao *StorePlatformsDao) Columns() StorePlatformsColumns { return dao.columns }
func (dao *StorePlatformsDao) Group() string                  { return dao.group }
func (dao *StorePlatformsDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}
func (dao *StorePlatformsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) error {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
