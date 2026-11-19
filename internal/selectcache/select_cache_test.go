package selectcache

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/os/gcfg"
)

func preservePrefix(t *testing.T) {
	t.Helper()
	previous := prefix.Load().(string)
	t.Cleanup(func() { prefix.Store(previous) })
}

func contentConfig(t *testing.T, content string) *gcfg.Config {
	t.Helper()
	adapter, err := gcfg.NewAdapterContent(content)
	if err != nil {
		t.Fatal(err)
	}
	return gcfg.NewWithAdapter(adapter)
}

func TestPrefixIsLoadedAndValidatedAtStartup(t *testing.T) {
	for _, test := range []struct {
		name, content, wantPrefix string
		wantError                 bool
	}{
		{name: "missing", content: "{}", wantPrefix: DefaultPrefix},
		{name: "configured", content: "cache:\n  prefix: warehouse-production\n", wantPrefix: "warehouse-production"},
		{name: "trimmed", content: "cache:\n  prefix: ' warehouse '\n", wantPrefix: "warehouse"},
		{name: "empty", content: "cache:\n  prefix: ''\n", wantError: true},
		{name: "whitespace", content: "cache:\n  prefix: '   '\n", wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			preservePrefix(t)
			prefix.Store("previous")
			err := initFromConfig(context.Background(), contentConfig(t, test.content))
			if (err != nil) != test.wantError {
				t.Fatalf("startup error=%v, want error=%t", err, test.wantError)
			}
			if test.wantError {
				if got := Name("products", 10); got != "previous#selectCache#products_10" {
					t.Fatalf("invalid configuration changed the prefix: %q", got)
				}
				return
			}
			want := test.wantPrefix + "#selectCache#products_10"
			if got := Name("products", 10); got != want {
				t.Fatalf("cache name=%q, want=%q", got, want)
			}
		})
	}
}

type countedConfig struct {
	calls int
	err   error
}

func (c *countedConfig) Get(_ context.Context, key string, _ ...any) (*gvar.Var, error) {
	c.calls++
	if key != "cache.prefix" {
		return nil, errors.New("unexpected configuration key")
	}
	return gvar.New("warehouse"), c.err
}

func TestNamesAndOptionsDoNotReadConfigurationDuringQueries(t *testing.T) {
	preservePrefix(t)
	config := &countedConfig{}
	if err := initFromConfig(context.Background(), config); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		query, invalidate := Query("products", 10), Invalidate("products", 10)
		want := "warehouse#selectCache#products_10"
		if query.Name != want || invalidate.Name != want || query.Duration != 0 || invalidate.Duration >= 0 {
			t.Fatalf("inconsistent read/write cache options: query=%+v invalidate=%+v", query, invalidate)
		}
	}
	if config.calls != 1 {
		t.Fatalf("configuration was read %d times, want 1 at initialization", config.calls)
	}
	if Name("products", 10) == Name("products", 11) || Name("products", 10) == Name("users", 10) {
		t.Fatal("different rows or tables share a cache name")
	}
}

func TestConfigurationReadFailurePreservesPreviousPrefix(t *testing.T) {
	preservePrefix(t)
	prefix.Store("previous")
	cause := errors.New("configuration unavailable")
	if err := initFromConfig(context.Background(), &countedConfig{err: cause}); !errors.Is(err, cause) {
		t.Fatalf("configuration error was lost: %v", err)
	}
	if got := Name("products", 10); got != "previous#selectCache#products_10" {
		t.Fatalf("failed initialization changed the prefix: %q", got)
	}
}

func TestExampleConfigurationsHaveSeparateCachePrefixes(t *testing.T) {
	for _, test := range []struct{ file, want string }{
		{file: "config.example.yaml", want: "indu-scope#selectCache#products_10"},
		{file: "config.production.example.yaml", want: "indu-scope-production#selectCache#products_10"},
	} {
		t.Run(test.file, func(t *testing.T) {
			preservePrefix(t)
			// Only public example files are opened; the real ignored config.yaml
			// and database/network services are never used by these tests.
			content, err := os.ReadFile(filepath.Join("..", "..", "manifest", "config", test.file))
			if err != nil {
				t.Fatal(err)
			}
			if err := initFromConfig(context.Background(), contentConfig(t, string(content))); err != nil {
				t.Fatal(err)
			}
			if got := Name("products", 10); got != test.want {
				t.Fatalf("example cache name=%q, want=%q", got, test.want)
			}
		})
	}
}
