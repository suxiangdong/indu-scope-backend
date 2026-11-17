package auth

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gogf/gf/v2/database/gdb"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	verificationModel "github.com/suxiangdong/indu-scope-backend/internal/model/verification"
	"github.com/suxiangdong/indu-scope-backend/internal/testutil/dbmock"
)

const verificationTestFields = `id user_id phone company_name status role user_type wechat_openid
real_name id_number store_platform_id store_url id_front_endpoint id_front_bucket id_front_path id_back_endpoint id_back_bucket id_back_path
business_license_endpoint business_license_bucket business_license_path unified_social_credit_code registered_address
review_status review_reason reviewed_by reviewed_at verified_at purpose endpoint bucket object_key introduction_html offices factories name position`

func verificationFixture(t *testing.T) (gdb.DB, sqlmock.Sqlmock) {
	return dbmock.New(t, verificationTestFields)
}

func expectSubmissionUser(mock sqlmock.Sqlmock, role, status int) {
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM `users`(.+)FOR UPDATE").WithArgs(uint64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "role", "status"}).AddRow(2, role, status))
}

func expectVerificationRows(mock sqlmock.Sqlmock, personal, company *sqlmock.Rows) {
	mock.ExpectQuery("SELECT (.+) FROM `personal_verifications` WHERE (.+)$").WithArgs(uint64(2)).WillReturnRows(personal)
	mock.ExpectQuery("SELECT (.+) FROM `company_verifications` WHERE (.+)$").WithArgs(uint64(2)).WillReturnRows(company)
}

func emptyPersonalVerification() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "user_id", "review_status"})
}

func emptyCompanyVerification() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "user_id", "review_status"})
}

func TestVerificationSubmissionRechecksUserUnderLock(t *testing.T) {
	for _, test := range []struct {
		name   string
		role   int
		status int
		want   error
	}{
		{"disabled while waiting", 1, 2, consts.ErrUserDisabled},
		{"approved while waiting", 1, 1, consts.ErrVerificationLocked},
		{"role switched while waiting", 2, 0, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			db, mock := verificationFixture(t)
			expectSubmissionUser(mock, test.role, test.status)
			mock.ExpectRollback()
			files, err := persistPersonalVerification(context.Background(), db, 2, 1, verificationModel.PersonalInput{})
			if err == nil || (test.want != nil && !errors.Is(err, test.want)) || len(files) != 0 {
				t.Fatalf("files=%v error=%v", files, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestVerificationSubmissionCannotOverwriteConcurrentPendingOrApproval(t *testing.T) {
	for _, personal := range []bool{true, false} {
		for _, status := range []int{0, 1} {
			db, mock := verificationFixture(t)
			expectSubmissionUser(mock, 1, 0)
			p, c := emptyPersonalVerification(), emptyCompanyVerification()
			if personal {
				p.AddRow(10, 2, status)
			} else {
				c.AddRow(11, 2, status)
			}
			expectVerificationRows(mock, p, c)
			mock.ExpectRollback()
			var err error
			if personal {
				_, err = persistPersonalVerification(context.Background(), db, 2, 1, verificationModel.PersonalInput{})
			} else {
				_, err = persistCompanyVerification(context.Background(), db, 2, 1, verificationModel.CompanyInput{})
			}
			if !errors.Is(err, consts.ErrVerificationLocked) {
				t.Fatalf("personal=%t status=%d error=%v", personal, status, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func personalSubmissionInput() verificationModel.PersonalInput {
	return verificationModel.PersonalInput{
		RealName: "测试用户", IDNumber: "11010519491231002X", StorePlatformID: 1, StoreURL: "https://example.com/store",
		IDFrontEndpoint: "oss.example.com", IDFrontBucket: "documents", IDFrontPath: "new-front.png",
		IDBackEndpoint: "oss.example.com", IDBackBucket: "documents", IDBackPath: "same-back.png",
	}
}

func expectRejectedPersonal(mock sqlmock.Sqlmock) {
	expectSubmissionUser(mock, 1, 0)
	p := sqlmock.NewRows([]string{"id", "user_id", "review_status", "id_front_endpoint", "id_front_bucket", "id_front_path", "id_back_endpoint", "id_back_bucket", "id_back_path"}).
		AddRow(10, 2, 2, "oss.example.com", "documents", "old-front.png", "oss.example.com", "documents", "same-back.png")
	expectVerificationRows(mock, p, emptyCompanyVerification())
}

func expectSubmissionFile(mock sqlmock.Sqlmock, purpose, key string, exists bool) {
	rows := sqlmock.NewRows([]string{"id", "user_id", "purpose", "endpoint", "bucket", "object_key"})
	if exists {
		rows.AddRow(30, 2, purpose, "oss.example.com", "documents", key)
	}
	mock.ExpectQuery("SELECT (.+) FROM `user_files`(.+)FOR UPDATE").
		WithArgs(uint64(2), purpose, "oss.example.com", "documents", key).WillReturnRows(rows)
}

func TestPersonalResubmissionOnlyRetiresActuallyReplacedMaterials(t *testing.T) {
	db, mock := verificationFixture(t)
	expectRejectedPersonal(mock)
	expectSubmissionFile(mock, idCardFrontPurpose, "new-front.png", true)
	expectSubmissionFile(mock, idCardBackPurpose, "same-back.png", true)
	mock.ExpectExec("UPDATE `personal_verifications`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE `users`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	files, err := persistPersonalVerification(context.Background(), db, 2, 1, personalSubmissionInput())
	want := []verificationFile{{idCardFrontPurpose, "oss.example.com", "documents", "old-front.png"}}
	if err != nil || !reflect.DeepEqual(files, want) {
		t.Fatalf("files=%v want=%v error=%v", files, want, err)
	}
	// The same in-flight submission, executing after the first commit, sees
	// pending status and cannot replace it or enqueue any more file deletions.
	expectSubmissionUser(mock, 1, 0)
	expectVerificationRows(mock, emptyPersonalVerification().AddRow(10, 2, 0), emptyCompanyVerification())
	mock.ExpectRollback()
	files, err = persistPersonalVerification(context.Background(), db, 2, 1, personalSubmissionInput())
	if !errors.Is(err, consts.ErrVerificationLocked) || len(files) != 0 {
		t.Fatalf("duplicate submission files=%v error=%v", files, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSubmissionRejectsFileRetiredAfterEarlierValidation(t *testing.T) {
	db, mock := verificationFixture(t)
	expectRejectedPersonal(mock)
	expectSubmissionFile(mock, idCardFrontPurpose, "new-front.png", false)
	mock.ExpectRollback()
	files, err := persistPersonalVerification(context.Background(), db, 2, 1, personalSubmissionInput())
	var parameter *consts.InvalidParameterError
	if !errors.As(err, &parameter) || len(files) != 0 {
		t.Fatalf("missing file was accepted: files=%v error=%v", files, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSubmissionDoesNotCleanUpAfterRollback(t *testing.T) {
	db, mock := verificationFixture(t)
	expectRejectedPersonal(mock)
	expectSubmissionFile(mock, idCardFrontPurpose, "new-front.png", true)
	expectSubmissionFile(mock, idCardBackPurpose, "same-back.png", true)
	mock.ExpectExec("UPDATE `personal_verifications`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE `users`").WillReturnError(errors.New("database unavailable"))
	mock.ExpectRollback()
	files, err := persistPersonalVerification(context.Background(), db, 2, 1, personalSubmissionInput())
	if err == nil || len(files) != 0 {
		t.Fatalf("rolled-back submission returned cleanup files=%v error=%v", files, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCompanyResubmissionKeepsProfileContactsAndStatusAtomic(t *testing.T) {
	for _, failProfile := range []bool{false, true} {
		db, mock := verificationFixture(t)
		expectSubmissionUser(mock, 2, 0)
		expectVerificationRows(mock, emptyPersonalVerification(),
			sqlmock.NewRows([]string{"id", "user_id", "review_status", "business_license_endpoint", "business_license_bucket", "business_license_path"}).
				AddRow(11, 2, 2, "oss.example.com", "documents", "old-license.png"))
		expectSubmissionFile(mock, businessLicensePurpose, "new-license.png", true)
		mock.ExpectExec("UPDATE `company_verifications`").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery("SELECT (.+) FROM `company_profiles`").WithArgs(uint64(2)).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}).AddRow(12, 2))
		if failProfile {
			mock.ExpectExec("UPDATE `company_profiles`").WillReturnError(errors.New("profile write failed"))
			mock.ExpectRollback()
		} else {
			mock.ExpectExec("UPDATE `company_profiles`").WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectExec("DELETE FROM `company_contacts`").WithArgs(uint64(2)).WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectExec("INSERT INTO `company_contacts`").WillReturnResult(sqlmock.NewResult(40, 1))
			mock.ExpectExec("UPDATE `users`").WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()
		}
		input := verificationModel.CompanyInput{CompanyName: "测试公司", BusinessLicenseEndpoint: "oss.example.com", BusinessLicenseBucket: "documents", BusinessLicensePath: "new-license.png",
			Contacts: []verificationModel.ContactInput{{Name: "联系人", Position: "经理", Phone: "13800138000"}}}
		files, err := persistCompanyVerification(context.Background(), db, 2, 2, input)
		if failProfile {
			if err == nil || len(files) != 0 {
				t.Fatalf("failed company update exposed cleanup files=%v error=%v", files, err)
			}
		} else if err != nil || len(files) != 1 || files[0].ObjectKey != "old-license.png" {
			t.Fatalf("company update files=%v error=%v", files, err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRetirementKeepsAFileReferencedByANewerSubmission(t *testing.T) {
	db, mock := verificationFixture(t)
	candidate := verificationFile{idCardFrontPurpose, "oss.example.com", "documents", "reused.png"}
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM `users`(.+)FOR UPDATE").WithArgs(uint64(2)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	expectVerificationRows(mock,
		sqlmock.NewRows([]string{"id", "id_front_endpoint", "id_front_bucket", "id_front_path"}).AddRow(10, candidate.Endpoint, candidate.Bucket, candidate.ObjectKey), emptyCompanyVerification())
	mock.ExpectCommit()
	files, err := retireVerificationFiles(context.Background(), db, 2, []verificationFile{candidate})
	if err != nil || len(files) != 0 {
		t.Fatalf("current material retired: files=%v error=%v", files, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRetirementRemovesOnlyTheSpecifiedUnreferencedFile(t *testing.T) {
	db, mock := verificationFixture(t)
	candidate := verificationFile{businessLicensePurpose, "oss.example.com", "documents", "old-license.png"}
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM `users`(.+)FOR UPDATE").WithArgs(uint64(2)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	expectVerificationRows(mock, emptyPersonalVerification(),
		sqlmock.NewRows([]string{"id", "business_license_endpoint", "business_license_bucket", "business_license_path"}).AddRow(11, candidate.Endpoint, candidate.Bucket, "current-license.png"))
	mock.ExpectQuery("SELECT (.+) FROM `user_files`(.+)FOR UPDATE").
		WithArgs(uint64(2), candidate.Purpose, candidate.Endpoint, candidate.Bucket, candidate.ObjectKey).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "purpose", "endpoint", "bucket", "object_key"}).AddRow(30, 2, candidate.Purpose, candidate.Endpoint, candidate.Bucket, candidate.ObjectKey))
	mock.ExpectExec("DELETE FROM `user_files` WHERE `id`=\\?").WithArgs(uint64(30)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	files, err := retireVerificationFiles(context.Background(), db, 2, []verificationFile{candidate})
	if err != nil || len(files) != 1 || files[0].ObjectKey != candidate.ObjectKey {
		t.Fatalf("files=%v error=%v", files, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestIncompleteLegacyMaterialsAreNeverDeleted(t *testing.T) {
	previous := []verificationFile{{businessLicensePurpose, "", "legacy-bucket", "legacy.png"}}
	if got := replacedVerificationFiles(previous, nil); len(got) != 0 {
		t.Fatalf("ambiguous legacy location selected for deletion: %v", got)
	}
	complete := []verificationFile{{businessLicensePurpose, "oss.example.com", "legacy-bucket", "legacy.png"}}
	if got := replacedVerificationFiles(complete, previous); len(got) != 0 {
		t.Fatalf("file with ambiguous current reference selected for deletion: %v", got)
	}
}
