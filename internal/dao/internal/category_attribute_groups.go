package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

type CategoryAttributeGroupsDao struct {
	table    string
	group    string
	columns  CategoryAttributeGroupsColumns
	handlers []gdb.ModelHandler
}

type CategoryAttributeGroupsColumns struct {
	Id         string
	CategoryId string
	Code       string
	Name       string
	SortOrder  string
	Status     string
	CreatedAt  string
	UpdatedAt  string
}

var categoryAttributeGroupsColumns = CategoryAttributeGroupsColumns{
	Id: "id", CategoryId: "category_id", Code: "code", Name: "name",
	SortOrder: "sort_order", Status: "status", CreatedAt: "created_at", UpdatedAt: "updated_at",
}

func NewCategoryAttributeGroupsDao(handlers ...gdb.ModelHandler) *CategoryAttributeGroupsDao {
	return &CategoryAttributeGroupsDao{
		group: "default", table: "category_attribute_groups",
		columns: categoryAttributeGroupsColumns, handlers: handlers,
	}
}

func (dao *CategoryAttributeGroupsDao) DB() gdb.DB                              { return g.DB(dao.group) }
func (dao *CategoryAttributeGroupsDao) Table() string                           { return dao.table }
func (dao *CategoryAttributeGroupsDao) Columns() CategoryAttributeGroupsColumns { return dao.columns }
func (dao *CategoryAttributeGroupsDao) Group() string                           { return dao.group }

func (dao *CategoryAttributeGroupsDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

func (dao *CategoryAttributeGroupsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) error {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
