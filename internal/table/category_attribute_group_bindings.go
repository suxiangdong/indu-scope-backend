package table

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
)

var CategoryAttributeGroupBindings = map[string]*gdb.TableField{
	"category_id": {Index: 0, Name: "category_id", Type: "bigint unsigned", Null: false, Key: "PRI"},
	"group_id":    {Index: 1, Name: "group_id", Type: "bigint unsigned", Null: false, Key: "PRI"},
	"sort_order":  {Index: 2, Name: "sort_order", Type: "int unsigned", Null: false, Default: "0"},
	"status":      {Index: 3, Name: "status", Type: "tinyint unsigned", Null: false, Default: "1"},
	"created_at":  {Index: 4, Name: "created_at", Type: "datetime(3)", Null: false, Default: "CURRENT_TIMESTAMP(3)", Extra: "DEFAULT_GENERATED"},
	"updated_at":  {Index: 5, Name: "updated_at", Type: "datetime(3)", Null: false, Default: "CURRENT_TIMESTAMP(3)", Extra: "DEFAULT_GENERATED on update CURRENT_TIMESTAMP(3)"},
}

func SetCategoryAttributeGroupBindingsTableFields(ctx context.Context, db gdb.DB, schema ...string) error {
	return db.GetCore().SetTableFields(ctx, "category_attribute_group_bindings", CategoryAttributeGroupBindings, schema...)
}
