package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	mysqlDriver "github.com/go-sql-driver/mysql"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
)

func expectWechatBindingLock(mock sqlmock.Sqlmock, phone, openID string, status int) {
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM `users`(.+)FOR UPDATE").WithArgs(uint64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "phone", "wechat_openid", "status", "role", "user_type"}).AddRow(2, phone, openID, status, 1, 1))
}

func TestConcurrentWechatBindingCannotOverwriteFirstIdentity(t *testing.T) {
	db, mock := verificationFixture(t)
	expectWechatBindingLock(mock, "13800138000", "", 0)
	mock.ExpectExec("UPDATE `users` SET `wechat_openid`=\\? WHERE `id`=\\?").WithArgs("openid-first", uint64(2)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	user, err := bindWechatUser(context.Background(), db, 2, "13800138000", "openid-first")
	if err != nil || user.WechatOpenID != "openid-first" {
		t.Fatalf("first binding: user=%+v error=%v", user, err)
	}
	// A competing request initially saw an empty OpenID, but now reads the
	// first commit while holding the row lock; it must not rebind or log in.
	expectWechatBindingLock(mock, "13800138000", "openid-first", 0)
	mock.ExpectRollback()
	user, err = bindWechatUser(context.Background(), db, 2, "13800138000", "openid-second")
	if !errors.Is(err, consts.ErrWechatPhoneConflict) || user != nil {
		t.Fatalf("competing identity accepted: user=%+v error=%v", user, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWechatBindingRechecksPhoneAndDisabledStatus(t *testing.T) {
	for _, test := range []struct {
		name   string
		phone  string
		status int
		want   error
	}{
		{"phone changed", "13900139000", 0, consts.ErrWechatPhoneConflict},
		{"user disabled", "13800138000", 2, consts.ErrUserDisabled},
	} {
		t.Run(test.name, func(t *testing.T) {
			db, mock := verificationFixture(t)
			expectWechatBindingLock(mock, test.phone, "", test.status)
			mock.ExpectRollback()
			_, err := bindWechatUser(context.Background(), db, 2, "13800138000", "openid-test")
			if !errors.Is(err, test.want) {
				t.Fatalf("error=%v want=%v", err, test.want)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestWechatBindingIsIdempotentForTheSameIdentity(t *testing.T) {
	db, mock := verificationFixture(t)
	expectWechatBindingLock(mock, "13800138000", "openid-test", 1)
	mock.ExpectCommit()
	user, err := bindWechatUser(context.Background(), db, 2, "13800138000", "openid-test")
	if err != nil || user.Status != 1 || user.WechatOpenID != "openid-test" {
		t.Fatalf("user=%+v error=%v", user, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWechatBindingMapsUniqueIdentityConflict(t *testing.T) {
	db, mock := verificationFixture(t)
	expectWechatBindingLock(mock, "13800138000", "", 0)
	mock.ExpectExec("UPDATE `users`").WillReturnError(&mysqlDriver.MySQLError{Number: 1062, Message: "duplicate OpenID"})
	mock.ExpectRollback()
	_, err := bindWechatUser(context.Background(), db, 2, "13800138000", "openid-test")
	if !errors.Is(err, consts.ErrWechatPhoneConflict) {
		t.Fatalf("error=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
