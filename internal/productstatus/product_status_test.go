package productstatus

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gogf/gf/v2/database/gdb"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	"github.com/suxiangdong/indu-scope-backend/internal/model/entity"
	"github.com/suxiangdong/indu-scope-backend/internal/selectcache"
	"github.com/suxiangdong/indu-scope-backend/internal/testutil/dbmock"
)

func fixture(t *testing.T, supplier bool) (*gdb.Model, sqlmock.Sqlmock) {
	t.Helper()
	db, mock := dbmock.New(t, "id supplier_user_id status updated_at")
	model := db.Model("products").Ctx(context.Background()).Where("id", uint64(10))
	if supplier {
		model = model.Where("supplier_user_id", uint64(2))
	}
	return model, mock
}

func expectStatus(mock sqlmock.Sqlmock, status uint8) {
	mock.ExpectQuery("SELECT `id`,`status` FROM `products` WHERE (.+)").
		WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow(10, status))
}

func expectScopedStatusUpdate(mock sqlmock.Sqlmock, supplier bool, status uint8) *sqlmock.ExpectedExec {
	// Match the whole statement: the WHERE clause must preserve the caller's
	// id/ownership scope, without appending a previous-status predicate.
	statement := "UPDATE `products` SET `status`=?,`updated_at`=? WHERE `id`=?"
	if supplier {
		statement = "UPDATE `products` SET `status`=?,`updated_at`=? WHERE (`id`=?) AND (`supplier_user_id`=?)"
		return mock.ExpectExec("^"+regexp.QuoteMeta(statement)+"$").
			WithArgs(status, sqlmock.AnyArg(), uint64(10), uint64(2))
	}
	return mock.ExpectExec("^"+regexp.QuoteMeta(statement)+"$").
		WithArgs(status, sqlmock.AnyArg(), uint64(10))
}

func TestPublicationChangesRequireSuccessfulReview(t *testing.T) {
	for _, supplier := range []bool{false, true} {
		for _, target := range []uint8{Online, Offline} {
			t.Run(fmt.Sprintf("supplier=%v/status=%d", supplier, target), func(t *testing.T) {
				model, mock := fixture(t, supplier)
				expectStatus(mock, PendingReview)
				if err := Change(model, target); !errors.Is(err, consts.ErrProductPendingReview) {
					t.Fatalf("unreviewed product must not change publication status: %v", err)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestReviewApprovesWithoutPublishingAndCannotRepeat(t *testing.T) {
	for _, current := range []uint8{PendingReview, Online, Offline} {
		t.Run(fmt.Sprint(current), func(t *testing.T) {
			model, mock := fixture(t, false)
			expectStatus(mock, current)
			if current == PendingReview {
				expectScopedStatusUpdate(mock, false, Offline).WillReturnResult(sqlmock.NewResult(0, 1))
			}
			err := Approve(model)
			if current == PendingReview && err != nil {
				t.Fatalf("approve: %v", err)
			}
			if current != PendingReview && !errors.Is(err, consts.ErrProductAlreadyReviewed) {
				t.Fatalf("repeat review: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestReviewedProductsCanChangePublicationStatusOrReturnToReview(t *testing.T) {
	for _, current := range []uint8{Online, Offline} {
		for _, next := range []uint8{PendingReview, Online, Offline} {
			t.Run(fmt.Sprintf("%d->%d", current, next), func(t *testing.T) {
				model, mock := fixture(t, true)
				expectStatus(mock, current)
				if next != current {
					expectScopedStatusUpdate(mock, true, next).WillReturnResult(sqlmock.NewResult(0, 1))
				}
				if err := Change(model, next); err != nil {
					t.Fatal(err)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestStatusUpdateWithNoMatchedRowReturnsError(t *testing.T) {
	for _, test := range []struct {
		name     string
		supplier bool
		review   bool
	}{
		{name: "admin status"},
		{name: "supplier status", supplier: true},
		{name: "admin review", review: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			model, mock := fixture(t, test.supplier)
			current, next := Offline, Online
			if test.review {
				current, next = PendingReview, Offline
			}
			expectStatus(mock, current)
			// A product removed or no longer matching the owner scope after
			// the read must not be reported as successfully updated.
			expectScopedStatusUpdate(mock, test.supplier, next).WillReturnResult(sqlmock.NewResult(0, 0))
			var err error
			if test.review {
				err = Approve(model)
			} else {
				err = Change(model, next)
			}
			if !errors.Is(err, consts.ErrProductStatusChanged) {
				t.Fatalf("an unmatched mutation must fail: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestMissingOrUnownedProductsAreNotFound(t *testing.T) {
	for _, noRowsError := range []bool{false, true} {
		t.Run(fmt.Sprint(noRowsError), func(t *testing.T) {
			model, mock := fixture(t, true)
			query := mock.ExpectQuery("SELECT `id`,`status` FROM `products` WHERE (.+)`id`=\\?(.+)`supplier_user_id`=\\?").
				WithArgs(uint64(10), uint64(2))
			if noRowsError {
				query.WillReturnError(sql.ErrNoRows)
			} else {
				query.WillReturnRows(sqlmock.NewRows([]string{"id", "status"}))
			}
			if err := Change(model, Online); !errors.Is(err, consts.ErrProductNotFound) {
				t.Fatalf("missing scoped product: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSuccessfulMutationInvalidatesSharedProductRowCache(t *testing.T) {
	for _, test := range []struct {
		name     string
		supplier bool
		review   bool
	}{
		{name: "admin publication"},
		{name: "supplier publication", supplier: true},
		{name: "admin review", review: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			model, mock := fixture(t, test.supplier)
			query := model.Clone().Cache(selectcache.Query("products", 10))
			var product entity.Products
			// All detail consumers use this full-row key. Keep it deliberately
			// stale so a mutation must check the scoped database row instead.
			mock.ExpectQuery("SELECT (.+) FROM `products` WHERE").
				WillReturnRows(sqlmock.NewRows([]string{"id", "supplier_user_id", "status"}).AddRow(10, 2, Online))
			if err := query.Clone().Scan(&product); err != nil {
				t.Fatal(err)
			}
			current, next := Offline, Online
			if test.review {
				current, next = PendingReview, Offline
			}
			expectStatus(mock, current)
			expectScopedStatusUpdate(mock, test.supplier, next).WillReturnResult(sqlmock.NewResult(0, 1))
			var err error
			if test.review {
				err = Approve(model)
			} else {
				err = Change(model, next)
			}
			if err != nil {
				t.Fatal(err)
			}
			mock.ExpectQuery("SELECT (.+) FROM `products` WHERE").
				WillReturnRows(sqlmock.NewRows([]string{"id", "supplier_user_id", "status"}).AddRow(10, 2, next))
			if err := query.Clone().Scan(&product); err != nil {
				t.Fatal(err)
			}
			if product.Status != uint(next) || product.SupplierUserId != 2 {
				t.Fatalf("detail must reload the complete updated product: %+v", product)
			}
			// The reloaded row is cached again; no additional SQL is expected.
			if err := query.Clone().Scan(&product); err != nil {
				t.Fatal(err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCachedProductDoesNotBypassScopedMutationCheck(t *testing.T) {
	model, mock := fixture(t, true)
	query := model.Clone().Cache(selectcache.Query("products", 10))
	var product entity.Products
	mock.ExpectQuery("SELECT (.+) FROM `products` WHERE").
		WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow(10, Offline))
	if err := query.Clone().Scan(&product); err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery("SELECT `id`,`status` FROM `products` WHERE (.+)`supplier_user_id`=\\?").
		WithArgs(uint64(10), uint64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status"}))
	if err := Change(model, Online); !errors.Is(err, consts.ErrProductNotFound) {
		t.Fatalf("a cached row must not authorize an unmatched owner: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFailedMutationDoesNotEvictProductRowCache(t *testing.T) {
	model, mock := fixture(t, false)
	query := model.Clone().Cache(selectcache.Query("products", 10))
	var product entity.Products
	mock.ExpectQuery("SELECT (.+) FROM `products` WHERE").
		WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow(10, Offline))
	if err := query.Clone().Scan(&product); err != nil {
		t.Fatal(err)
	}
	expectStatus(mock, Offline)
	failure := errors.New("write failed")
	expectScopedStatusUpdate(mock, false, Online).WillReturnError(failure)
	if err := Change(model, Online); !errors.Is(err, failure) {
		t.Fatalf("mutation must return the database failure: %v", err)
	}
	// A failed database write must leave the previously cached value intact.
	if err := query.Clone().Scan(&product); err != nil || product.Status != uint(Offline) {
		t.Fatalf("failed write changed cached product: %+v, %v", product, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
