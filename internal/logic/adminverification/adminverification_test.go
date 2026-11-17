package adminverification

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	model "github.com/suxiangdong/indu-scope-backend/internal/model/adminverification"
	"github.com/suxiangdong/indu-scope-backend/internal/testutil/dbmock"
)

const reviewFields = "id user_id role status company_name review_status review_reason reviewed_by reviewed_at verified_at"

func expectReviewUser(mock sqlmock.Sqlmock, role, status int) {
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM `users`(.+)FOR UPDATE").WithArgs(uint64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "role", "status"}).AddRow(2, role, status))
}

func TestVerificationReviewUsesLockedCompanyAndUserRows(t *testing.T) {
	db, mock := dbmock.New(t, reviewFields)
	expectReviewUser(mock, 2, 0)
	mock.ExpectQuery("SELECT (.+) FROM `company_verifications` WHERE (.+)$").WithArgs(uint64(10), uint64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "review_status", "company_name"}).AddRow(10, 2, 0, "最新提交的公司名称"))
	mock.ExpectExec("UPDATE `company_verifications`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("^UPDATE `users` SET `status`=\\?,`company_name`=\\? WHERE `id`=\\?$").
		WithArgs(uint(1), "最新提交的公司名称", uint64(2)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := applyVerificationReview(context.Background(), db, 1, 2, model.ReviewInput{Type: model.TypeCompany, ID: 10, Action: model.ActionApprove}); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestVerificationReviewCannotEnableConcurrentlyDisabledAccount(t *testing.T) {
	db, mock := dbmock.New(t, reviewFields)
	expectReviewUser(mock, 2, 2)
	mock.ExpectRollback()
	err := applyVerificationReview(context.Background(), db, 1, 2, model.ReviewInput{Type: model.TypeCompany, ID: 10, Action: model.ActionApprove})
	if !errors.Is(err, consts.ErrUserDisabled) {
		t.Fatalf("disabled account reviewed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestVerificationReviewRejectsConcurrentRoleChange(t *testing.T) {
	db, mock := dbmock.New(t, reviewFields)
	expectReviewUser(mock, 2, 0)
	mock.ExpectRollback()
	err := applyVerificationReview(context.Background(), db, 1, 2, model.ReviewInput{Type: model.TypePersonal, ID: 10, Action: model.ActionApprove})
	if !errors.Is(err, consts.ErrVerificationTypeConflict) {
		t.Fatalf("personal verification approved for supplier: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestVerificationReviewCannotOverwriteAnotherReview(t *testing.T) {
	db, mock := dbmock.New(t, reviewFields)
	expectReviewUser(mock, 1, 1)
	mock.ExpectQuery("SELECT (.+) FROM `personal_verifications` WHERE (.+)$").WithArgs(uint64(10), uint64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "review_status"}).AddRow(10, 2, 1))
	mock.ExpectRollback()
	err := applyVerificationReview(context.Background(), db, 1, 2, model.ReviewInput{Type: model.TypePersonal, ID: 10, Action: model.ActionReject, Reason: "另一审核意见"})
	if !errors.Is(err, consts.ErrVerificationAlreadyReviewed) {
		t.Fatalf("existing review overwritten: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestVerificationReviewRollsBackWhenUserUpdateFails(t *testing.T) {
	db, mock := dbmock.New(t, reviewFields)
	expectReviewUser(mock, 1, 0)
	mock.ExpectQuery("SELECT (.+) FROM `personal_verifications` WHERE (.+)$").WithArgs(uint64(10), uint64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "review_status"}).AddRow(10, 2, 0))
	mock.ExpectExec("UPDATE `personal_verifications`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE `users`").WillReturnError(errors.New("database unavailable"))
	mock.ExpectRollback()
	err := applyVerificationReview(context.Background(), db, 1, 2, model.ReviewInput{Type: model.TypePersonal, ID: 10, Action: model.ActionApprove})
	if err == nil {
		t.Fatal("partial review succeeded")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
