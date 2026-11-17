// Package dbmock provides an isolated GoFrame database backed only by SQLMock.
// It never reads the application's database configuration or opens a network.
package dbmock

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	mysqldriver "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/database/gdb"
)

var sequence atomic.Uint64

type driver struct {
	*mysqldriver.Driver
	db     *sql.DB
	fields []string
}

func (d *driver) New(core *gdb.Core, _ *gdb.ConfigNode) (gdb.DB, error) {
	return &driver{Driver: &mysqldriver.Driver{Core: core}, db: d.db, fields: d.fields}, nil
}

func (d *driver) Open(_ *gdb.ConfigNode) (*sql.DB, error) { return d.db, nil }

func (d *driver) TableFields(_ context.Context, _ string, _ ...string) (map[string]*gdb.TableField, error) {
	fields := make(map[string]*gdb.TableField, len(d.fields))
	for i, name := range d.fields {
		kind := "varchar(1024)"
		if name == "id" || strings.HasSuffix(name, "_id") {
			kind = "bigint unsigned"
		} else if strings.HasSuffix(name, "_at") {
			kind = "datetime(3)"
		}
		fields[name] = &gdb.TableField{Index: i, Name: name, Type: kind}
	}
	return fields, nil
}

func New(t *testing.T, fields string) (gdb.DB, sqlmock.Sqlmock) {
	t.Helper()
	raw, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	name := fmt.Sprintf("isolated_sqlmock_%d", sequence.Add(1))
	if err := gdb.Register(name, &driver{Driver: &mysqldriver.Driver{}, db: raw, fields: strings.Fields(fields)}); err != nil {
		t.Fatal(err)
	}
	db, err := gdb.New(gdb.ConfigNode{Type: name, Name: name})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = raw.Close() })
	return db, mock
}
