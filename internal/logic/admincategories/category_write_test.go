package admincategories

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	"github.com/suxiangdong/indu-scope-backend/internal/testutil/dbmock"
)

func TestCategoryMoveChecksAncestryAfterAcquiringTreeLock(t *testing.T) {
	db, mock := dbmock.New(t, "lock_key id parent_id name sort_order status")
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM `backend_write_locks`(.+)FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"lock_key"}).AddRow("category_tree"))
	mock.ExpectQuery("SELECT (.+) FROM `categories`").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectQuery("SELECT COUNT(.+) FROM `categories`").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	// The preceding transaction moved A under B while this transaction waited
	// for the sentinel. Moving B under A must now fail without any UPDATE.
	mock.ExpectQuery("SELECT (.+) FROM `categories`").WillReturnRows(sqlmock.NewRows([]string{"id", "parent_id"}).AddRow(1, 2))
	mock.ExpectRollback()
	parent := uint64(1)
	err := db.Transaction(context.Background(), func(ctx context.Context, tx gdb.TX) error { return updateCategoryTx(ctx, tx, 2, &parent, "B", 0, 1) })
	if !errors.Is(err, consts.ErrCategoryParentCycle) {
		t.Fatalf("expected cycle rejection, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCategoryCreationCannotUseDeletedParent(t *testing.T) {
	db, mock := dbmock.New(t, "lock_key id parent_id name sort_order status")
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM `backend_write_locks`(.+)FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"lock_key"}).AddRow("category_tree"))
	mock.ExpectQuery("SELECT COUNT(.+) FROM `categories`").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectRollback()
	parent := uint64(1)
	err := db.Transaction(context.Background(), func(ctx context.Context, tx gdb.TX) error {
		_, err := createCategoryTx(ctx, tx, &parent, "B", 0, 1)
		return err
	})
	if !errors.Is(err, consts.ErrCategoryParentNotFound) {
		t.Fatalf("expected missing parent rejection, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
