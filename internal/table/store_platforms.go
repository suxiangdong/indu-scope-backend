// =================================================================================
// Code generated and maintained by GoFrame CLI tool. You may modify it as needed.
// =================================================================================

package table

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
)

var StorePlatforms = map[string]*gdb.TableField{
	"id":         {Index: 0, Name: "id", Type: "bigint unsigned", Null: false, Key: "PRI", Extra: "auto_increment"},
	"name":       {Index: 1, Name: "name", Type: "varchar(100)", Null: false, Key: "UNI"},
	"sort_order": {Index: 2, Name: "sort_order", Type: "int unsigned", Null: false, Default: "0"},
	"status":     {Index: 3, Name: "status", Type: "tinyint unsigned", Null: false, Default: "1"},
	"created_at": {Index: 4, Name: "created_at", Type: "datetime(3)", Null: false, Default: "CURRENT_TIMESTAMP(3)", Extra: "DEFAULT_GENERATED"},
	"updated_at": {Index: 5, Name: "updated_at", Type: "datetime(3)", Null: false, Default: "CURRENT_TIMESTAMP(3)", Extra: "DEFAULT_GENERATED on update CURRENT_TIMESTAMP(3)"},
}

func SetStorePlatformsTableFields(ctx context.Context, db gdb.DB, schema ...string) error {
	return db.GetCore().SetTableFields(ctx, "store_platforms", StorePlatforms, schema...)
}
