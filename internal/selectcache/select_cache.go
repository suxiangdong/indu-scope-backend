// Package selectcache defines shared names for cached, complete table rows.
package selectcache

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

const DefaultPrefix = "indu-scope"

var prefix atomic.Value

func init() { prefix.Store(DefaultPrefix) }

type configReader interface {
	Get(context.Context, string, ...any) (*gvar.Var, error)
}

// Init loads cache.prefix at startup. A missing setting retains compatibility
// with older configuration files, while an explicit blank setting is invalid.
func Init(ctx context.Context) error { return initFromConfig(ctx, g.Cfg()) }

func initFromConfig(ctx context.Context, config configReader) error {
	value, err := config.Get(ctx, "cache.prefix", DefaultPrefix)
	if err != nil {
		return fmt.Errorf("load cache.prefix: %w", err)
	}
	configured := DefaultPrefix
	if value != nil && !value.IsNil() {
		configured = strings.TrimSpace(value.String())
	}
	if configured == "" {
		return errors.New("cache.prefix must not be blank")
	}
	prefix.Store(configured)
	return nil
}

// Name identifies a complete physical table row by ID. Partial-field queries
// or queries with additional permission filters must not share this name.
func Name(table string, id uint64) string {
	return prefix.Load().(string) + "#selectCache#" + table + "_" + strconv.FormatUint(id, 10)
}

// Query retains the existing non-expiring cache duration. Only complete rows
// queried by ID may share this key; writes must use the matching Invalidate.
func Query(table string, id uint64) gdb.CacheOption {
	return gdb.CacheOption{Name: Name(table, id)}
}

// Invalidate is applied to the corresponding model's update/delete operation
// so GoFrame removes the same named SELECT result after a successful write.
func Invalidate(table string, id uint64) gdb.CacheOption {
	return gdb.CacheOption{Name: Name(table, id), Duration: -1}
}
