package cmd

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"

	"github.com/suxiangdong/indu-scope-backend/internal/migration"
)

// Migrate is the database migration command group.
var Migrate = gcmd.Command{
	Name:  "migrate",
	Usage: "main migrate COMMAND",
	Brief: "manage database migrations",
}

var migrateUp = gcmd.Command{
	Name:  "up",
	Usage: "main migrate up",
	Brief: "apply all pending database migrations",
	Func: func(ctx context.Context, parser *gcmd.Parser) error {
		dsn, err := databaseDSN(ctx)
		if err != nil {
			return err
		}
		return migration.Up(dsn)
	},
}

var migrateDown = gcmd.Command{
	Name:  "down",
	Usage: "main migrate down",
	Brief: "rollback the latest database migration",
	Func: func(ctx context.Context, parser *gcmd.Parser) error {
		dsn, err := databaseDSN(ctx)
		if err != nil {
			return err
		}
		return migration.Down(dsn)
	},
}

func init() {
	if err := Migrate.AddCommand(&migrateUp, &migrateDown); err != nil {
		panic(err)
	}
	if err := Main.AddCommand(&Migrate); err != nil {
		panic(err)
	}
}

func databaseDSN(ctx context.Context) (string, error) {
	value, err := g.Cfg().Get(ctx, "database.default.link")
	if err != nil {
		return "", gerror.Wrap(err, "read database.default.link from config")
	}
	if value == nil || strings.TrimSpace(value.String()) == "" {
		return "", gerror.New("database.default.link is required in config.yaml")
	}
	return value.String(), nil
}
