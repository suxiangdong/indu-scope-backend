package productstatus

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gogf/gf/v2/database/gdb"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
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
				mock.ExpectExec("UPDATE `products` SET `status`=\\?,`updated_at`=\\? WHERE (.+)`id`=\\?(.+)`status`=\\?").
					WithArgs(Offline, sqlmock.AnyArg(), uint64(10), PendingReview).
					WillReturnResult(sqlmock.NewResult(0, 1))
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
					mock.ExpectExec("UPDATE `products` SET `status`=\\?,`updated_at`=\\? WHERE (.+)`id`=\\?(.+)`supplier_user_id`=\\?(.+)`status`=\\?").
						WithArgs(next, sqlmock.AnyArg(), uint64(10), uint64(2), current).
						WillReturnResult(sqlmock.NewResult(0, 1))
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

func TestConcurrentStatusChangeCannotOverwriteNewReviewState(t *testing.T) {
	for _, review := range []bool{false, true} {
		t.Run(fmt.Sprint(review), func(t *testing.T) {
			model, mock := fixture(t, false)
			current, next := Offline, Online
			if review {
				current, next = PendingReview, Offline
			}
			expectStatus(mock, current)
			// Another operation changed the row after the read; the guarded
			// UPDATE must leave that state intact and report a conflict.
			mock.ExpectExec("UPDATE `products` SET `status`=\\?,`updated_at`=\\? WHERE (.+)`id`=\\?(.+)`status`=\\?").
				WithArgs(next, sqlmock.AnyArg(), uint64(10), current).WillReturnResult(sqlmock.NewResult(0, 0))
			var err error
			if review {
				err = Approve(model)
			} else {
				err = Change(model, next)
			}
			if !errors.Is(err, consts.ErrProductStatusChanged) {
				t.Fatalf("stale mutation must fail: %v", err)
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
