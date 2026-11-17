package procurementrequirements

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	procurementRequirementModel "github.com/suxiangdong/indu-scope-backend/internal/model/procurementrequirement"
	"github.com/suxiangdong/indu-scope-backend/internal/testutil/dbmock"
)

func TestResponseLocksParentBeforeInsertingChild(t *testing.T) {
	db, mock := dbmock.New(t, "id requirement_id supplier_user_id message available_quantity lead_time_days status review_status expires_at deleted_at response_count")
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM `procurement_requirements`(.+)FOR UPDATE").WithArgs(uint64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "review_status", "expires_at"}).AddRow(10, 1, 1, time.Now().Add(time.Hour)))
	mock.ExpectExec("INSERT INTO `procurement_requirement_responses`").WillReturnResult(sqlmock.NewResult(7, 1))
	mock.ExpectExec("UPDATE `procurement_requirements`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	id, err := createResponse(context.Background(), db, procurementRequirementModel.ResponseInput{RequirementID: 10, SupplierUserID: 2, Message: "available"})
	if err != nil || id != 7 {
		t.Fatalf("create response: id=%d err=%v", id, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestResponseRejectsClosedOrExpiredLockedRequirement(t *testing.T) {
	for _, test := range []struct {
		name   string
		status uint8
		end    time.Time
	}{
		{"closed", 2, time.Now().Add(time.Hour)},
		{"expired", 1, time.Now().Add(-time.Hour)},
	} {
		t.Run(test.name, func(t *testing.T) {
			db, mock := dbmock.New(t, "id requirement_id supplier_user_id status review_status expires_at deleted_at response_count")
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT (.+) FROM `procurement_requirements`(.+)FOR UPDATE").
				WillReturnRows(sqlmock.NewRows([]string{"id", "status", "review_status", "expires_at"}).AddRow(10, test.status, 1, test.end))
			mock.ExpectRollback()
			_, err := createResponse(context.Background(), db, procurementRequirementModel.ResponseInput{RequirementID: 10, SupplierUserID: 2})
			if !errors.Is(err, consts.ErrRequirementClosed) {
				t.Fatalf("closed requirement: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestResponseInsertRollsBackWhenExpiryCheckFails(t *testing.T) {
	db, mock := dbmock.New(t, "id requirement_id supplier_user_id message available_quantity lead_time_days status review_status expires_at deleted_at response_count")
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM `procurement_requirements`(.+)FOR UPDATE").
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "review_status", "expires_at"}).AddRow(10, 1, 1, time.Now().Add(time.Hour)))
	mock.ExpectExec("INSERT INTO `procurement_requirement_responses`").WillReturnResult(sqlmock.NewResult(7, 1))
	// The requirement can expire while the request is being processed.
	mock.ExpectExec("UPDATE `procurement_requirements`").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()
	_, err := createResponse(context.Background(), db, procurementRequirementModel.ResponseInput{RequirementID: 10, SupplierUserID: 2})
	if !errors.Is(err, consts.ErrRequirementClosed) {
		t.Fatalf("expiry during create: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
