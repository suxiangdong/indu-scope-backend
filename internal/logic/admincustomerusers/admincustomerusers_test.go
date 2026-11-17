package admincustomerusers

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	"github.com/suxiangdong/indu-scope-backend/internal/testutil/dbmock"
)

func TestRoleSwitchLocksUserFirstAndNeverWritesStatus(t *testing.T) {
	db, mock := dbmock.New(t, "id user_id role status user_type")
	mock.ExpectBegin()
	// A concurrent disable may have committed before this row lock. The
	// role switch must preserve that disabled status rather than restoring 1.
	mock.ExpectQuery("SELECT (.+) FROM `users`(.+)FOR UPDATE").WithArgs(uint64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "role", "status"}).AddRow(2, 1, 2))
	mock.ExpectExec("DELETE FROM `personal_verifications` WHERE `user_id`=\\?").WithArgs(uint64(2)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM `buyer_sales_channels` WHERE `user_id`=\\?").WithArgs(uint64(2)).WillReturnResult(sqlmock.NewResult(0, 1))
	// Anchoring the full UPDATE ensures neither status nor any other stale
	// user fields are included in the write.
	mock.ExpectExec("^UPDATE `users` SET `role`=\\?,`user_type`=\\? WHERE `id`=\\?$").WithArgs(uint8(2), uint8(2), uint64(2)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := switchCustomerRole(context.Background(), db, 2, 2); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRoleSwitchUsesTheLockedCurrentRole(t *testing.T) {
	db, mock := dbmock.New(t, "id user_id role status user_type")
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM `users`(.+)FOR UPDATE").WithArgs(uint64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "role"}).AddRow(2, 2))
	mock.ExpectCommit()
	if err := switchCustomerRole(context.Background(), db, 2, 2); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRoleSwitchRollsBackVerificationDeletionOnFailure(t *testing.T) {
	db, mock := dbmock.New(t, "id user_id role status user_type")
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM `users`(.+)FOR UPDATE").WithArgs(uint64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "role"}).AddRow(2, 2))
	mock.ExpectExec("DELETE FROM `company_contacts`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM `company_profiles`").WillReturnError(errors.New("database unavailable"))
	mock.ExpectRollback()
	if err := switchCustomerRole(context.Background(), db, 2, 1); err == nil {
		t.Fatal("partial role switch succeeded")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRoleSwitchRejectsDeletedUserUnderLock(t *testing.T) {
	db, mock := dbmock.New(t, "id role user_type")
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM `users`(.+)FOR UPDATE").WithArgs(uint64(2)).WillReturnRows(sqlmock.NewRows([]string{"id", "role"}))
	mock.ExpectRollback()
	if err := switchCustomerRole(context.Background(), db, 2, 1); !errors.Is(err, consts.ErrCustomerUserNotFound) {
		t.Fatalf("error=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
