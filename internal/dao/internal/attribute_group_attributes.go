package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

type AttributeGroupAttributesDao struct {
	table    string
	group    string
	columns  AttributeGroupAttributesColumns
	handlers []gdb.ModelHandler
}

type AttributeGroupAttributesColumns struct {
	Id          string
	GroupId     string
	AttributeId string
	SortOrder   string
	CreatedAt   string
	UpdatedAt   string
}

var attributeGroupAttributesColumns = AttributeGroupAttributesColumns{
	Id:      "id",
	GroupId: "group_id", AttributeId: "attribute_id", SortOrder: "sort_order",
	CreatedAt: "created_at", UpdatedAt: "updated_at",
}

func NewAttributeGroupAttributesDao(handlers ...gdb.ModelHandler) *AttributeGroupAttributesDao {
	return &AttributeGroupAttributesDao{
		group: "default", table: "attribute_group_attributes", columns: attributeGroupAttributesColumns, handlers: handlers,
	}
}

func (dao *AttributeGroupAttributesDao) DB() gdb.DB                               { return g.DB(dao.group) }
func (dao *AttributeGroupAttributesDao) Table() string                            { return dao.table }
func (dao *AttributeGroupAttributesDao) Columns() AttributeGroupAttributesColumns { return dao.columns }
func (dao *AttributeGroupAttributesDao) Group() string                            { return dao.group }

func (dao *AttributeGroupAttributesDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

func (dao *AttributeGroupAttributesDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) error {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
