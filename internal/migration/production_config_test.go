package migration

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	mysqlDriver "github.com/go-sql-driver/mysql"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/util/gconv"
)

func TestProductionDatabaseTemplateWorksThroughGoFrameAndMigrations(t *testing.T) {
	path, err := filepath.Abs(filepath.Join("..", "..", "manifest", "config", "config.production.example.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	adapter, err := gcfg.NewAdapterFile(path)
	if err != nil {
		t.Fatal(err)
	}
	config, err := gcfg.NewWithAdapter(adapter).Get(context.Background(), "database.default")
	if err != nil {
		t.Fatal(err)
	}
	var node gdb.ConfigNode
	if err := gconv.Struct(config.Map(), &node); err != nil {
		t.Fatal(err)
	}
	// Exercise the same normalization as application startup with dummy values.
	// sql.Open parses the DSN but never dials a server; this test never calls Ping.
	node.Link = strings.NewReplacer(
		"__MYSQL_USER__", "config_test",
		"__MYSQL_PASSWORD__", "not-a-real-password",
		"__MYSQL_HOST__", "127.0.0.1",
		"__MYSQL_DATABASE__", "config_test",
	).Replace(node.Link)
	db, err := gdb.New(node)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close(context.Background())
	raw, err := db.Open(db.GetConfig())
	if err != nil {
		t.Fatalf("GoFrame could not open the template DSN: %v", err)
	}
	if err := raw.Close(); err != nil {
		t.Fatal(err)
	}

	// Migrations use the underlying SQL driver directly. Both code paths must
	// accept the same configuration and interpret DATETIME in the same zone.
	dsn, err := normalizeMySQLDSN(node.Link)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := mysqlDriver.ParseDSN(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Loc.String() != "Asia/Shanghai" || !parsed.ParseTime || !parsed.MultiStatements {
		t.Fatalf("unexpected migration connection settings: location=%s parseTime=%t multiStatements=%t", parsed.Loc, parsed.ParseTime, parsed.MultiStatements)
	}
}
