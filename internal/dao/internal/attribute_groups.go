package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

type AttributeGroupsDao struct {
	table    string
	group    string
	columns  AttributeGroupsColumns
	handlers []gdb.ModelHandler
}

type AttributeGroupsColumns struct {
	Id          string
	Code        string
	Name        string
	DisplayName string
	SortOrder   string
	Status      string
	CreatedAt   string
	UpdatedAt   string
}

var attributeGroupsColumns = AttributeGroupsColumns{
	Id: "id", Code: "code", Name: "name", DisplayName: "display_name", SortOrder: "sort_order", Status: "status",
	CreatedAt: "created_at", UpdatedAt: "updated_at",
}

func NewAttributeGroupsDao(handlers ...gdb.ModelHandler) *AttributeGroupsDao {
	return &AttributeGroupsDao{
		group: "default", table: "attribute_groups", columns: attributeGroupsColumns, handlers: handlers,
	}
}

func (dao *AttributeGroupsDao) DB() gdb.DB                      { return g.DB(dao.group) }
func (dao *AttributeGroupsDao) Table() string                   { return dao.table }
func (dao *AttributeGroupsDao) Columns() AttributeGroupsColumns { return dao.columns }
func (dao *AttributeGroupsDao) Group() string                   { return dao.group }

func (dao *AttributeGroupsDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

func (dao *AttributeGroupsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) error {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
