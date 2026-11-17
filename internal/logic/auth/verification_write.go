package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	"github.com/suxiangdong/indu-scope-backend/internal/dao"
	"github.com/suxiangdong/indu-scope-backend/internal/model/entity"
	verificationModel "github.com/suxiangdong/indu-scope-backend/internal/model/verification"
	"github.com/suxiangdong/indu-scope-backend/internal/storage"
)

type verificationFile struct {
	Purpose, Endpoint, Bucket, ObjectKey string
}

func (f verificationFile) complete() bool {
	return f.Purpose != "" && f.Endpoint != "" && f.Bucket != "" && f.ObjectKey != ""
}

// Submission, review, role switching and file retirement all lock the user
// before touching verification rows. Re-read status under that lock rather
// than trusting the identity loaded by middleware at the start of a request.
func lockVerificationSubmission(tx gdb.TX, userID uint64, role uint8, personal bool) (*entity.PersonalVerifications, *entity.CompanyVerifications, error) {
	var user *entity.Users
	if err := tx.Model("users").Fields("id", "role", "status").Where("id", userID).LockUpdate().Scan(&user); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, gerror.Wrap(err, "lock verification user")
	}
	if user == nil {
		return nil, nil, consts.ErrTokenInvalid
	}
	if user.Status == 2 {
		return nil, nil, consts.ErrUserDisabled
	}
	if user.Status == 1 {
		return nil, nil, consts.ErrVerificationLocked
	}
	if _, err := resolveVerificationRole(uint8(user.Role), role); err != nil {
		return nil, nil, err
	}
	if personal && role != 1 {
		return nil, nil, consts.ErrVerificationTypeConflict
	}
	var personalRecord *entity.PersonalVerifications
	// The user lock serializes all writers. Do not lock absent child rows:
	// their index gap can be shared by unrelated users and deadlock inserts.
	// This is the first consistent read, so its snapshot follows the user lock.
	if err := tx.Model("personal_verifications").Where("user_id", userID).Scan(&personalRecord); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, gerror.Wrap(err, "read personal verification after user lock")
	}
	var companyRecord *entity.CompanyVerifications
	if err := tx.Model("company_verifications").Where("user_id", userID).Scan(&companyRecord); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, gerror.Wrap(err, "read company verification after user lock")
	}
	if personal {
		if companyRecord != nil {
			return nil, nil, consts.ErrVerificationTypeConflict
		}
		if personalRecord != nil && (personalRecord.ReviewStatus != uint(verificationModel.ReviewStatusRejected) || !personalRecord.VerifiedAt.IsZero()) {
			return nil, nil, consts.ErrVerificationLocked
		}
	} else {
		if personalRecord != nil {
			return nil, nil, consts.ErrVerificationTypeConflict
		}
		if companyRecord != nil && (companyRecord.ReviewStatus != uint(verificationModel.ReviewStatusRejected) || !companyRecord.VerifiedAt.IsZero()) {
			return nil, nil, consts.ErrVerificationLocked
		}
	}
	return personalRecord, companyRecord, nil
}

func validateVerificationFileTx(tx gdb.TX, userID uint64, file verificationFile) error {
	if !file.complete() {
		return &consts.InvalidParameterError{Message: "verification endpoint, bucket and path are required"}
	}
	var record *userFileRecord
	err := tx.Model("user_files").Where("user_id", userID).Where("purpose", file.Purpose).
		Where("endpoint", file.Endpoint).Where("bucket", file.Bucket).Where("object_key", file.ObjectKey).LockUpdate().Scan(&record)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return gerror.Wrap(err, "lock verification file")
	}
	if record == nil {
		return &consts.InvalidParameterError{Message: "verification files must be uploaded before submission"}
	}
	return nil
}

func persistPersonalVerification(ctx context.Context, db gdb.DB, userID uint64, role uint8, input verificationModel.PersonalInput) ([]verificationFile, error) {
	var replaced []verificationFile
	err := db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		record, _, err := lockVerificationSubmission(tx, userID, role, true)
		if err != nil {
			return err
		}
		front := verificationFile{idCardFrontPurpose, input.IDFrontEndpoint, input.IDFrontBucket, input.IDFrontPath}
		back := verificationFile{idCardBackPurpose, input.IDBackEndpoint, input.IDBackBucket, input.IDBackPath}
		for _, file := range []verificationFile{front, back} {
			if err := validateVerificationFileTx(tx, userID, file); err != nil {
				return err
			}
		}
		data := g.Map{
			"user_id": userID, "real_name": input.RealName, "id_number": input.IDNumber,
			"store_platform_id": input.StorePlatformID, "store_url": input.StoreURL,
			"id_front_endpoint": front.Endpoint, "id_front_bucket": front.Bucket, "id_front_path": front.ObjectKey,
			"id_back_endpoint": back.Endpoint, "id_back_bucket": back.Bucket, "id_back_path": back.ObjectKey,
			"verified_at": nil, "review_status": verificationModel.ReviewStatusPending,
			"review_reason": nil, "reviewed_by": nil, "reviewed_at": nil,
		}
		if record == nil {
			if _, err := tx.Model("personal_verifications").Data(data).Insert(); err != nil {
				return gerror.Wrap(err, "create personal verification")
			}
		} else {
			if _, err := tx.Model("personal_verifications").Where("id", record.Id).Data(data).Update(); err != nil {
				return gerror.Wrap(err, "resubmit personal verification")
			}
			replaced = replacedVerificationFiles(
				[]verificationFile{{idCardFrontPurpose, record.IdFrontEndpoint, record.IdFrontBucket, record.IdFrontPath}, {idCardBackPurpose, record.IdBackEndpoint, record.IdBackBucket, record.IdBackPath}},
				[]verificationFile{front, back},
			)
		}
		_, err = tx.Model("users").Where("id", userID).Data(g.Map{"role": role, "user_type": 1, "status": 0}).Update()
		return gerror.Wrap(err, "update personal verification status")
	})
	if err != nil {
		return nil, err
	}
	return replaced, nil
}

func persistCompanyVerification(ctx context.Context, db gdb.DB, userID uint64, role uint8, input verificationModel.CompanyInput) ([]verificationFile, error) {
	var replaced []verificationFile
	err := db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_, record, err := lockVerificationSubmission(tx, userID, role, false)
		if err != nil {
			return err
		}
		file := verificationFile{businessLicensePurpose, input.BusinessLicenseEndpoint, input.BusinessLicenseBucket, input.BusinessLicensePath}
		if err := validateVerificationFileTx(tx, userID, file); err != nil {
			return err
		}
		data := g.Map{
			"user_id": userID, "company_name": input.CompanyName, "unified_social_credit_code": input.UnifiedSocialCreditCode,
			"business_license_endpoint": file.Endpoint, "business_license_bucket": file.Bucket, "business_license_path": file.ObjectKey,
			"registered_address": input.RegisteredAddress, "verified_at": nil,
			"review_status": verificationModel.ReviewStatusPending, "review_reason": nil, "reviewed_by": nil, "reviewed_at": nil,
		}
		if record == nil {
			if _, err := tx.Model("company_verifications").Data(data).Insert(); err != nil {
				return gerror.Wrap(err, "create company verification")
			}
		} else {
			if _, err := tx.Model("company_verifications").Where("id", record.Id).Data(data).Update(); err != nil {
				return gerror.Wrap(err, "resubmit company verification")
			}
			replaced = replacedVerificationFiles([]verificationFile{{businessLicensePurpose, record.BusinessLicenseEndpoint, record.BusinessLicenseBucket, record.BusinessLicensePath}}, []verificationFile{file})
		}
		if err := saveCompanyProfileTx(ctx, tx, userID, input); err != nil {
			return err
		}
		if err := replaceCompanyContactsTx(tx, userID, input.Contacts); err != nil {
			return err
		}
		_, err = tx.Model("users").Where("id", userID).Data(g.Map{"role": role, "user_type": 2, "status": 0}).Update()
		return gerror.Wrap(err, "update company verification status")
	})
	if err != nil {
		return nil, err
	}
	return replaced, nil
}

func replacedVerificationFiles(previous, current []verificationFile) []verificationFile {
	var replaced []verificationFile
	for _, old := range previous {
		if !old.complete() {
			continue
		}
		retained := false
		for _, file := range current {
			// An old reference without an endpoint is ambiguous; retain the
			// matching bucket/key rather than deleting a potentially live file.
			if (old.Endpoint == file.Endpoint || file.Endpoint == "") && old.Bucket == file.Bucket && old.ObjectKey == file.ObjectKey {
				retained = true
				break
			}
		}
		if !retained {
			replaced = append(replaced, old)
		}
	}
	return replaced
}

// Retire metadata under the same user lock as submission, then perform OSS I/O
// after committing. A later submission must validate metadata under this lock,
// so it cannot attach a file between the reference check and object deletion.
func retireVerificationFiles(ctx context.Context, db gdb.DB, userID uint64, candidates []verificationFile) ([]userFileRecord, error) {
	if len(candidates) == 0 {
		return nil, nil
	}
	var retired []userFileRecord
	err := db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var user *entity.Users
		if err := tx.Model("users").Fields("id").Where("id", userID).LockUpdate().Scan(&user); err != nil {
			return err
		}
		if user == nil {
			return consts.ErrTokenInvalid
		}
		var personal *entity.PersonalVerifications
		if err := tx.Model("personal_verifications").Where("user_id", userID).Scan(&personal); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		var company *entity.CompanyVerifications
		if err := tx.Model("company_verifications").Where("user_id", userID).Scan(&company); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		var current []verificationFile
		if personal != nil {
			current = append(current, verificationFile{idCardFrontPurpose, personal.IdFrontEndpoint, personal.IdFrontBucket, personal.IdFrontPath}, verificationFile{idCardBackPurpose, personal.IdBackEndpoint, personal.IdBackBucket, personal.IdBackPath})
		}
		if company != nil {
			current = append(current, verificationFile{businessLicensePurpose, company.BusinessLicenseEndpoint, company.BusinessLicenseBucket, company.BusinessLicensePath})
		}
		for _, file := range replacedVerificationFiles(candidates, current) {
			var record *userFileRecord
			err := tx.Model("user_files").Where("user_id", userID).Where("purpose", file.Purpose).
				Where("endpoint", file.Endpoint).Where("bucket", file.Bucket).Where("object_key", file.ObjectKey).LockUpdate().Scan(&record)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return err
			}
			if record == nil {
				continue
			}
			if _, err := tx.Model("user_files").Where("id", record.ID).Delete(); err != nil {
				return err
			}
			retired = append(retired, *record)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return retired, nil
}

func cleanupReplacedVerificationFiles(ctx context.Context, userID uint64, candidates []verificationFile) {
	if len(candidates) == 0 {
		return
	}
	client, err := storage.OSSClient()
	if err != nil {
		g.Log().Warningf(ctx, "initialize OSS for retired verification files for user %d: %v", userID, err)
		return
	}
	files, err := retireVerificationFiles(ctx, dao.Users.DB(), userID, candidates)
	if err != nil {
		g.Log().Warningf(ctx, "retire replaced verification files for user %d: %v", userID, err)
		return
	}
	for _, file := range files {
		deleteCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		err := client.DeleteFor(deleteCtx, file.Endpoint, file.Bucket, file.ObjectKey)
		cancel()
		if err != nil {
			g.Log().Warningf(ctx, "delete retired verification file %d: %v", file.ID, err)
		}
	}
}
