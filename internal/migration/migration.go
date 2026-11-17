package migration

import (
	"database/sql"
	"embed"
	"fmt"
	"net/url"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	mysqlmigrate "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// migrationsFS keeps migration files in the application binary so migrations
// can be executed from a built image without mounting the source repository.
//
//go:embed migrations/*.sql
var migrationsFS embed.FS

// New creates a migration runner for a GoFrame-style MySQL DSN, for example:
// mysql:user:password@tcp(host:3306)/database?charset=utf8mb4&parseTime=true
func New(dsn string) (*migrate.Migrate, error) {
	rawDSN, err := normalizeMySQLDSN(dsn)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("mysql", rawDSN)
	if err != nil {
		return nil, fmt.Errorf("open mysql connection: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	database, err := mysqlmigrate.WithInstance(db, &mysqlmigrate.Config{})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create mysql migration driver: %w", err)
	}

	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create migration source: %w", err)
	}

	runner, err := migrate.NewWithInstance("iofs", source, "indu_scope", database)
	if err != nil {
		db.Close()
		return nil, err
	}
	return runner, nil
}

// normalizeMySQLDSN converts a GoFrame-style DSN to a database/sql MySQL DSN
// and enables multi-statement migration files.
func normalizeMySQLDSN(dsn string) (string, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return "", fmt.Errorf("mysql DSN is empty")
	}
	if strings.HasPrefix(dsn, "mysql:") && !strings.HasPrefix(dsn, "mysql://") {
		dsn = strings.TrimPrefix(dsn, "mysql:")
	} else if strings.HasPrefix(dsn, "mysql://") {
		parsed, err := url.Parse(dsn)
		if err != nil {
			return "", fmt.Errorf("parse mysql DSN: %w", err)
		}
		if parsed.Host == "" || strings.Trim(parsed.Path, "/") == "" {
			return "", fmt.Errorf("mysql URL must include host and database")
		}
		userinfo := ""
		if parsed.User != nil {
			userinfo = parsed.User.String()
		}
		dsn = userinfo + "@tcp(" + parsed.Host + ")/" + strings.TrimPrefix(parsed.Path, "/") + "?" + parsed.RawQuery
	}
	if dsn == "" {
		return "", fmt.Errorf("mysql DSN is empty")
	}

	rawQuery := ""
	if strings.Contains(dsn, "?") {
		rawQuery = strings.SplitN(dsn, "?", 2)[1]
	}
	parsed, err := url.Parse("mysql://placeholder/placeholder?" + rawQuery)
	if err != nil {
		return "", fmt.Errorf("parse mysql DSN parameters: %w", err)
	}
	query := parsed.Query()
	query.Set("multiStatements", "true")
	if strings.Contains(dsn, "?") {
		return strings.SplitN(dsn, "?", 2)[0] + "?" + query.Encode(), nil
	}
	return dsn + "?" + query.Encode(), nil
}

// Up applies all pending migrations. ErrNoChange is treated as success.
func Up(dsn string) error {
	runner, err := New(dsn)
	if err != nil {
		return err
	}
	defer runner.Close()

	if err := runner.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}

// Down rolls back the latest migration.
func Down(dsn string) error {
	runner, err := New(dsn)
	if err != nil {
		return err
	}
	defer runner.Close()

	if err := runner.Steps(-1); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("rollback migration: %w", err)
	}
	return nil
}

// Force marks the migration metadata at version without running SQL. It is
// intended only to clear a dirty migration after the operator has verified
// whether its SQL was applied partially.
func Force(dsn string, version int) error {
	runner, err := New(dsn)
	if err != nil {
		return err
	}
	defer runner.Close()
	if err := runner.Force(version); err != nil {
		return fmt.Errorf("force migration version: %w", err)
	}
	return nil
}
