package table

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
)

var AttributeGroups = map[string]*gdb.TableField{
	"id":           {Index: 0, Name: "id", Type: "bigint unsigned", Null: false, Key: "PRI", Extra: "auto_increment"},
	"code":         {Index: 1, Name: "code", Type: "varchar(64)", Null: false, Key: "UNI"},
	"name":         {Index: 2, Name: "name", Type: "varchar(100)", Null: false, Key: "UNI"},
	"display_name": {Index: 3, Name: "display_name", Type: "varchar(100)", Null: false},
	"sort_order":   {Index: 4, Name: "sort_order", Type: "int unsigned", Null: false, Default: "0"},
	"status":       {Index: 5, Name: "status", Type: "tinyint unsigned", Null: false, Default: "1"},
	"created_at":   {Index: 6, Name: "created_at", Type: "datetime(3)", Null: false, Default: "CURRENT_TIMESTAMP(3)", Extra: "DEFAULT_GENERATED"},
	"updated_at":   {Index: 7, Name: "updated_at", Type: "datetime(3)", Null: false, Default: "CURRENT_TIMESTAMP(3)", Extra: "DEFAULT_GENERATED on update CURRENT_TIMESTAMP(3)"},
}

func SetAttributeGroupsTableFields(ctx context.Context, db gdb.DB, schema ...string) error {
	return db.GetCore().SetTableFields(ctx, "attribute_groups", AttributeGroups, schema...)
}
