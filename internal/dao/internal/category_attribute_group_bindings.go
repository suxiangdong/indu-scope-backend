package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

type CategoryAttributeGroupBindingsDao struct {
	table    string
	group    string
	columns  CategoryAttributeGroupBindingsColumns
	handlers []gdb.ModelHandler
}

type CategoryAttributeGroupBindingsColumns struct {
	Id         string
	CategoryId string
	GroupId    string
	SortOrder  string
	Status     string
	CreatedAt  string
	UpdatedAt  string
}

var categoryAttributeGroupBindingsColumns = CategoryAttributeGroupBindingsColumns{
	Id:         "id",
	CategoryId: "category_id", GroupId: "group_id", SortOrder: "sort_order", Status: "status",
	CreatedAt: "created_at", UpdatedAt: "updated_at",
}

func NewCategoryAttributeGroupBindingsDao(handlers ...gdb.ModelHandler) *CategoryAttributeGroupBindingsDao {
	return &CategoryAttributeGroupBindingsDao{
		group: "default", table: "category_attribute_group_bindings", columns: categoryAttributeGroupBindingsColumns, handlers: handlers,
	}
}

func (dao *CategoryAttributeGroupBindingsDao) DB() gdb.DB    { return g.DB(dao.group) }
func (dao *CategoryAttributeGroupBindingsDao) Table() string { return dao.table }
func (dao *CategoryAttributeGroupBindingsDao) Columns() CategoryAttributeGroupBindingsColumns {
	return dao.columns
}
func (dao *CategoryAttributeGroupBindingsDao) Group() string { return dao.group }

func (dao *CategoryAttributeGroupBindingsDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

func (dao *CategoryAttributeGroupBindingsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) error {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
