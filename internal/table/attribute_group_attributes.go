package table

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
)

var AttributeGroupAttributes = map[string]*gdb.TableField{
	"group_id":     {Index: 0, Name: "group_id", Type: "bigint unsigned", Null: false, Key: "PRI"},
	"attribute_id": {Index: 1, Name: "attribute_id", Type: "bigint unsigned", Null: false, Key: "PRI"},
	"sort_order":   {Index: 2, Name: "sort_order", Type: "int unsigned", Null: false, Default: "0"},
	"created_at":   {Index: 3, Name: "created_at", Type: "datetime(3)", Null: false, Default: "CURRENT_TIMESTAMP(3)", Extra: "DEFAULT_GENERATED"},
	"updated_at":   {Index: 4, Name: "updated_at", Type: "datetime(3)", Null: false, Default: "CURRENT_TIMESTAMP(3)", Extra: "DEFAULT_GENERATED on update CURRENT_TIMESTAMP(3)"},
}

func SetAttributeGroupAttributesTableFields(ctx context.Context, db gdb.DB, schema ...string) error {
	return db.GetCore().SetTableFields(ctx, "attribute_group_attributes", AttributeGroupAttributes, schema...)
}
