package adminverification

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	"github.com/suxiangdong/indu-scope-backend/internal/dao"
	adminverificationmodel "github.com/suxiangdong/indu-scope-backend/internal/model/adminverification"
	"github.com/suxiangdong/indu-scope-backend/internal/model/do"
	"github.com/suxiangdong/indu-scope-backend/internal/model/entity"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
	"github.com/suxiangdong/indu-scope-backend/internal/storage"
)

type sAdminVerification struct{}

var _ service.IAdminVerification = (*sAdminVerification)(nil)

func New() service.IAdminVerification {
	return &sAdminVerification{}
}

func init() {
	service.RegisterAdminVerification(New())
}

func (s *sAdminVerification) List(ctx context.Context, filter adminverificationmodel.ListFilter) (adminverificationmodel.ListResult, error) {
	if filter.Type != "" && filter.Type != adminverificationmodel.TypePersonal && filter.Type != adminverificationmodel.TypeCompany {
		return adminverificationmodel.ListResult{}, &consts.InvalidParameterError{Message: "type must be personal or company"}
	}
	if filter.ReviewStatus != nil && *filter.ReviewStatus > verificationModelReviewRejected {
		return adminverificationmodel.ListResult{}, &consts.InvalidParameterError{Message: "review_status must be 0, 1 or 2"}
	}
	items := make([]adminverificationmodel.ListItem, 0)
	if filter.Type == "" || filter.Type == adminverificationmodel.TypePersonal {
		personalItems, err := s.listPersonal(ctx, filter.ReviewStatus)
		if err != nil {
			return adminverificationmodel.ListResult{}, err
		}
		items = append(items, personalItems...)
	}
	if filter.Type == "" || filter.Type == adminverificationmodel.TypeCompany {
		companyItems, err := s.listCompany(ctx, filter.ReviewStatus)
		if err != nil {
			return adminverificationmodel.ListResult{}, err
		}
		items = append(items, companyItems...)
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return adminverificationmodel.ListResult{Items: items}, nil
}

func (s *sAdminVerification) listPersonal(ctx context.Context, reviewStatus *uint8) ([]adminverificationmodel.ListItem, error) {
	cols := dao.PersonalVerifications.Columns()
	query := dao.PersonalVerifications.Ctx(ctx).Fields(
		cols.Id, cols.UserId, cols.RealName, cols.ReviewStatus, cols.ReviewReason,
		cols.ReviewedAt, cols.CreatedAt, cols.UpdatedAt,
	).OrderDesc(cols.CreatedAt)
	if reviewStatus != nil {
		query = query.Where(do.PersonalVerifications{ReviewStatus: *reviewStatus})
	}
	var records []*entity.PersonalVerifications
	if err := query.Scan(&records); err != nil {
		return nil, gerror.Wrap(err, "list personal verifications")
	}
	users, err := loadUsers(ctx, personalUserIDs(records))
	if err != nil {
		return nil, err
	}
	items := make([]adminverificationmodel.ListItem, 0, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		user := users[record.UserId]
		items = append(items, adminverificationmodel.ListItem{
			ID: record.Id, UserID: record.UserId, Type: adminverificationmodel.TypePersonal,
			Phone: user.Phone, UserType: uint8(user.UserType), DisplayName: record.RealName,
			ReviewStatus: uint8(record.ReviewStatus), ReviewReason: record.ReviewReason,
			CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt, ReviewedAt: timePointer(record.ReviewedAt),
		})
	}
	return items, nil
}

func (s *sAdminVerification) listCompany(ctx context.Context, reviewStatus *uint8) ([]adminverificationmodel.ListItem, error) {
	cols := dao.CompanyVerifications.Columns()
	query := dao.CompanyVerifications.Ctx(ctx).Fields(
		cols.Id, cols.UserId, cols.CompanyName, cols.ReviewStatus, cols.ReviewReason,
		cols.ReviewedAt, cols.CreatedAt, cols.UpdatedAt,
	).OrderDesc(cols.CreatedAt)
	if reviewStatus != nil {
		query = query.Where(do.CompanyVerifications{ReviewStatus: *reviewStatus})
	}
	var records []*entity.CompanyVerifications
	if err := query.Scan(&records); err != nil {
		return nil, gerror.Wrap(err, "list company verifications")
	}
	users, err := loadUsers(ctx, companyUserIDs(records))
	if err != nil {
		return nil, err
	}
	items := make([]adminverificationmodel.ListItem, 0, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		user := users[record.UserId]
		items = append(items, adminverificationmodel.ListItem{
			ID: record.Id, UserID: record.UserId, Type: adminverificationmodel.TypeCompany,
			Phone: user.Phone, UserType: uint8(user.UserType), DisplayName: record.CompanyName,
			ReviewStatus: uint8(record.ReviewStatus), ReviewReason: record.ReviewReason,
			CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt, ReviewedAt: timePointer(record.ReviewedAt),
		})
	}
	return items, nil
}

func (s *sAdminVerification) Detail(ctx context.Context, verificationType string, id uint64) (adminverificationmodel.Detail, error) {
	if err := validateLookup(verificationType, id); err != nil {
		return adminverificationmodel.Detail{}, err
	}
	if verificationType == adminverificationmodel.TypePersonal {
		return s.personalDetail(ctx, id)
	}
	return s.companyDetail(ctx, id)
}

func (s *sAdminVerification) personalDetail(ctx context.Context, id uint64) (adminverificationmodel.Detail, error) {
	cols := dao.PersonalVerifications.Columns()
	var record *entity.PersonalVerifications
	if err := dao.PersonalVerifications.Ctx(ctx).Fields(
		cols.Id, cols.UserId, cols.RealName, cols.IdNumber, cols.IdNumberEncrypted,
		cols.StorePlatformId, cols.StoreUrl,
		cols.IdFrontEndpoint, cols.IdFrontBucket, cols.IdFrontPath,
		cols.IdBackEndpoint, cols.IdBackBucket, cols.IdBackPath,
		cols.VerifiedAt, cols.ReviewStatus, cols.ReviewReason, cols.ReviewedAt,
		cols.CreatedAt, cols.UpdatedAt,
	).Where(do.PersonalVerifications{Id: id}).Scan(&record); err != nil {
		return adminverificationmodel.Detail{}, gerror.Wrap(err, "query personal verification")
	}
	if record == nil {
		return adminverificationmodel.Detail{}, consts.ErrVerificationNotFound
	}
	user, err := loadUser(ctx, record.UserId)
	if err != nil {
		return adminverificationmodel.Detail{}, err
	}
	storePlatformName, err := loadStorePlatformName(ctx, record.StorePlatformId)
	if err != nil {
		return adminverificationmodel.Detail{}, err
	}
	idNumberMasked := maskIDNumber(record.IdNumber)
	if idNumberMasked == "" && len(record.IdNumberEncrypted) > 0 {
		idNumberMasked = "已提交"
	}
	item := listItemFromPersonal(record, user)
	return adminverificationmodel.Detail{
		ListItem: item, UserStatus: uint8(user.Status), Role: uint8(user.Role),
		Personal: &adminverificationmodel.PersonalDetail{
			RealName: record.RealName, IDNumberMasked: idNumberMasked,
			StorePlatformID: record.StorePlatformId, StorePlatformName: storePlatformName, StoreURL: record.StoreUrl,
			IDFrontEndpoint: record.IdFrontEndpoint, IDFrontBucket: record.IdFrontBucket, IDFrontPath: record.IdFrontPath,
			IDBackEndpoint: record.IdBackEndpoint, IDBackBucket: record.IdBackBucket, IDBackPath: record.IdBackPath,
			VerifiedAt: timePointer(record.VerifiedAt),
		},
	}, nil
}

func (s *sAdminVerification) companyDetail(ctx context.Context, id uint64) (adminverificationmodel.Detail, error) {
	cols := dao.CompanyVerifications.Columns()
	var record *entity.CompanyVerifications
	if err := dao.CompanyVerifications.Ctx(ctx).Fields(
		cols.Id, cols.UserId, cols.CompanyName, cols.UnifiedSocialCreditCode,
		cols.BusinessLicenseEndpoint, cols.BusinessLicenseBucket, cols.BusinessLicensePath,
		cols.RegisteredAddress, cols.VerifiedAt,
		cols.ReviewStatus, cols.ReviewReason, cols.ReviewedAt, cols.CreatedAt, cols.UpdatedAt,
	).Where(do.CompanyVerifications{Id: id}).Scan(&record); err != nil {
		return adminverificationmodel.Detail{}, gerror.Wrap(err, "query company verification")
	}
	if record == nil {
		return adminverificationmodel.Detail{}, consts.ErrVerificationNotFound
	}
	user, err := loadUser(ctx, record.UserId)
	if err != nil {
		return adminverificationmodel.Detail{}, err
	}
	profile, err := loadCompanyProfile(ctx, record.UserId)
	if err != nil {
		return adminverificationmodel.Detail{}, err
	}
	contacts, err := loadCompanyContacts(ctx, record.UserId)
	if err != nil {
		return adminverificationmodel.Detail{}, err
	}
	file, err := loadBusinessLicenseFile(ctx, record.UserId, record.BusinessLicenseEndpoint, record.BusinessLicenseBucket, record.BusinessLicensePath)
	if err != nil {
		return adminverificationmodel.Detail{}, err
	}
	offices, factories, introduction, storePlatformID, storeURL := decodeProfile(profile)
	storePlatformName, err := loadStorePlatformName(ctx, storePlatformID)
	if err != nil {
		return adminverificationmodel.Detail{}, err
	}
	company := &adminverificationmodel.CompanyDetail{
		CompanyName: record.CompanyName, UnifiedSocialCreditCode: record.UnifiedSocialCreditCode,
		BusinessLicenseEndpoint: record.BusinessLicenseEndpoint, BusinessLicenseBucket: record.BusinessLicenseBucket,
		BusinessLicensePath: record.BusinessLicensePath, RegisteredAddress: record.RegisteredAddress,
		Introduction: introduction, StorePlatformID: storePlatformID, StorePlatformName: storePlatformName, StoreURL: storeURL,
		Offices: offices, Factories: factories, Contacts: contacts,
		VerifiedAt: timePointer(record.VerifiedAt),
	}
	if file != nil {
		company.BusinessLicenseEndpoint = file.Endpoint
		company.BusinessLicenseBucket = file.Bucket
		company.BusinessLicensePath = file.ObjectKey
		company.BusinessLicenseFileName = file.OriginalName
		company.BusinessLicenseFileSize = int64(file.FileSize)
		company.BusinessLicenseContentType = file.ContentType
		if client, clientErr := storage.OSSClient(); clientErr == nil {
			company.BusinessLicenseURL, _ = client.SignedURLFor(file.Endpoint, file.Bucket, file.ObjectKey, 10*time.Minute)
		}
	}
	return adminverificationmodel.Detail{ListItem: listItemFromCompany(record, user), UserStatus: uint8(user.Status), Role: uint8(user.Role), Company: company}, nil
}

func (s *sAdminVerification) Review(ctx context.Context, adminID uint64, input adminverificationmodel.ReviewInput) (adminverificationmodel.Detail, error) {
	input.Type = strings.TrimSpace(input.Type)
	input.Action = strings.TrimSpace(input.Action)
	input.Reason = strings.TrimSpace(input.Reason)
	if adminID == 0 {
		return adminverificationmodel.Detail{}, gerror.New("admin authorization is required")
	}
	if err := validateLookup(input.Type, input.ID); err != nil {
		return adminverificationmodel.Detail{}, err
	}
	if input.Action != adminverificationmodel.ActionApprove && input.Action != adminverificationmodel.ActionReject {
		return adminverificationmodel.Detail{}, &consts.InvalidParameterError{Message: "action must be approve or reject"}
	}
	if input.Action == adminverificationmodel.ActionReject && (input.Reason == "" || len(input.Reason) > 500) {
		return adminverificationmodel.Detail{}, &consts.InvalidParameterError{Message: "reason is required and must contain at most 500 characters when rejecting"}
	}
	var userID uint64
	if input.Type == adminverificationmodel.TypePersonal {
		var record *entity.PersonalVerifications
		if err := dao.PersonalVerifications.Ctx(ctx).Where(do.PersonalVerifications{Id: input.ID}).Scan(&record); err != nil {
			return adminverificationModelError(err, "query personal verification")
		}
		if record == nil {
			return adminverificationModelError(consts.ErrVerificationNotFound, "")
		}
		userID = record.UserId
	} else {
		var record *entity.CompanyVerifications
		if err := dao.CompanyVerifications.Ctx(ctx).Where(do.CompanyVerifications{Id: input.ID}).Scan(&record); err != nil {
			return adminverificationModelError(err, "query company verification")
		}
		if record == nil {
			return adminverificationModelError(consts.ErrVerificationNotFound, "")
		}
		userID = record.UserId
	}
	if err := applyVerificationReview(ctx, dao.Users.DB(), adminID, userID, input); err != nil {
		return adminverificationModelError(err, "review verification")
	}
	return s.Detail(ctx, input.Type, input.ID)
}

func applyVerificationReview(ctx context.Context, db gdb.DB, adminID, userID uint64, input adminverificationmodel.ReviewInput) error {
	return db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// Match submission/role switching: always lock the user first, and
		// decide from the current row rather than the pre-transaction hint.
		var user *entity.Users
		if err := tx.Model("users").Fields("id", "role", "status").Where("id", userID).LockUpdate().Scan(&user); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return gerror.Wrap(err, "lock verification user for review")
		}
		if user == nil {
			return consts.ErrVerificationNotFound
		}
		if user.Status == 2 {
			return consts.ErrUserDisabled
		}
		if user.Role != 1 && user.Role != 2 || (input.Type == adminverificationmodel.TypePersonal && user.Role != 1) {
			return consts.ErrVerificationTypeConflict
		}
		table := "personal_verifications"
		var companyName string
		var currentStatus uint
		if input.Type == adminverificationmodel.TypePersonal {
			var record *entity.PersonalVerifications
			if err := tx.Model(table).Where("id", input.ID).Where("user_id", userID).Scan(&record); err != nil && !errors.Is(err, sql.ErrNoRows) {
				return gerror.Wrap(err, "read personal verification after user lock")
			}
			if record == nil {
				return consts.ErrVerificationNotFound
			}
			currentStatus = record.ReviewStatus
		} else {
			table = "company_verifications"
			var record *entity.CompanyVerifications
			if err := tx.Model(table).Where("id", input.ID).Where("user_id", userID).Scan(&record); err != nil && !errors.Is(err, sql.ErrNoRows) {
				return gerror.Wrap(err, "read company verification after user lock")
			}
			if record == nil {
				return consts.ErrVerificationNotFound
			}
			currentStatus, companyName = record.ReviewStatus, strings.TrimSpace(record.CompanyName)
		}
		if currentStatus != verificationModelReviewPending {
			return consts.ErrVerificationAlreadyReviewed
		}
		now := time.Now()
		reviewStatus, userStatus := uint(verificationModelReviewApproved), uint(1)
		if input.Action == adminverificationmodel.ActionReject {
			reviewStatus, userStatus = verificationModelReviewRejected, 0
		}
		result, err := tx.Model(table).Where("id", input.ID).Where("review_status", verificationModelReviewPending).Data(g.Map{
			"review_status": reviewStatus, "review_reason": nullableReason(input.Reason),
			"reviewed_by": adminID, "reviewed_at": now, "verified_at": verifiedAt(input.Action, now),
		}).Update()
		if err != nil {
			return gerror.Wrap(err, "review verification")
		}
		if affected, err := result.RowsAffected(); err != nil {
			return gerror.Wrap(err, "check verification review")
		} else if affected == 0 {
			return consts.ErrVerificationAlreadyReviewed
		}
		userData := g.Map{"status": userStatus}
		if input.Type == adminverificationmodel.TypeCompany && input.Action == adminverificationmodel.ActionApprove && companyName != "" {
			userData["company_name"] = companyName
		}
		if _, err := tx.Model("users").Where("id", userID).Data(userData).Update(); err != nil {
			return gerror.Wrap(err, "update user verification status")
		}
		return nil
	})
}

const (
	verificationModelReviewPending  = 0
	verificationModelReviewApproved = 1
	verificationModelReviewRejected = 2
)

func validateLookup(verificationType string, id uint64) error {
	if id == 0 {
		return &consts.InvalidParameterError{Message: "id is required"}
	}
	if verificationType != adminverificationmodel.TypePersonal && verificationType != adminverificationmodel.TypeCompany {
		return &consts.InvalidParameterError{Message: "type must be personal or company"}
	}
	return nil
}

func adminverificationModelError(err error, operation string) (adminverificationmodel.Detail, error) {
	if operation == "" {
		return adminverificationmodel.Detail{}, err
	}
	return adminverificationmodel.Detail{}, gerror.Wrap(err, operation)
}

func nullableReason(reason string) any {
	if reason == "" {
		return nil
	}
	return reason
}

func verifiedAt(action string, now time.Time) any {
	if action == adminverificationmodel.ActionApprove {
		return now
	}
	return nil
}

func loadUser(ctx context.Context, id uint64) (*entity.Users, error) {
	cols := dao.Users.Columns()
	var user *entity.Users
	if err := dao.Users.Ctx(ctx).Fields(cols.Id, cols.Phone, cols.CompanyName, cols.Status, cols.Role, cols.UserType).
		Where(do.Users{Id: id}).Scan(&user); err != nil {
		return nil, gerror.Wrap(err, "query verification user")
	}
	if user == nil {
		return nil, gerror.New("verification user not found")
	}
	return user, nil
}

func loadUsers(ctx context.Context, ids []uint64) (map[uint64]*entity.Users, error) {
	result := make(map[uint64]*entity.Users, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	cols := dao.Users.Columns()
	var users []*entity.Users
	if err := dao.Users.Ctx(ctx).Fields(cols.Id, cols.Phone, cols.CompanyName, cols.Status, cols.Role, cols.UserType).
		WhereIn(cols.Id, ids).Scan(&users); err != nil {
		return nil, gerror.Wrap(err, "query verification users")
	}
	for _, user := range users {
		if user != nil {
			result[user.Id] = user
		}
	}
	return result, nil
}

func personalUserIDs(records []*entity.PersonalVerifications) []uint64 {
	ids := make([]uint64, 0, len(records))
	for _, record := range records {
		if record != nil {
			ids = append(ids, record.UserId)
		}
	}
	return ids
}

func companyUserIDs(records []*entity.CompanyVerifications) []uint64 {
	ids := make([]uint64, 0, len(records))
	for _, record := range records {
		if record != nil {
			ids = append(ids, record.UserId)
		}
	}
	return ids
}

func listItemFromPersonal(record *entity.PersonalVerifications, user *entity.Users) adminverificationmodel.ListItem {
	return adminverificationmodel.ListItem{ID: record.Id, UserID: record.UserId, Type: adminverificationmodel.TypePersonal,
		Phone: user.Phone, UserType: uint8(user.UserType), DisplayName: record.RealName,
		ReviewStatus: uint8(record.ReviewStatus), ReviewReason: record.ReviewReason,
		CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt, ReviewedAt: timePointer(record.ReviewedAt)}
}

func listItemFromCompany(record *entity.CompanyVerifications, user *entity.Users) adminverificationmodel.ListItem {
	return adminverificationmodel.ListItem{ID: record.Id, UserID: record.UserId, Type: adminverificationmodel.TypeCompany,
		Phone: user.Phone, UserType: uint8(user.UserType), DisplayName: record.CompanyName,
		ReviewStatus: uint8(record.ReviewStatus), ReviewReason: record.ReviewReason,
		CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt, ReviewedAt: timePointer(record.ReviewedAt)}
}

type fileRecord struct {
	Endpoint     string `orm:"endpoint"`
	Bucket       string `orm:"bucket"`
	ObjectKey    string `orm:"object_key"`
	OriginalName string `orm:"original_name"`
	ContentType  string `orm:"content_type"`
	FileSize     uint64 `orm:"file_size"`
}

func loadBusinessLicenseFile(ctx context.Context, userID uint64, endpoint, bucket, path string) (*fileRecord, error) {
	model := dao.UserFiles.Ctx(ctx).Where("user_id", userID).Where("purpose", "business_license")
	if endpoint != "" && bucket != "" && path != "" {
		model = model.Where("endpoint", endpoint).Where("bucket", bucket).Where("object_key", path)
	} else if bucket != "" && path != "" {
		model = model.Where("bucket", bucket).Where("object_key", path)
	} else {
		return nil, nil
	}
	var record *fileRecord
	if err := model.Scan(&record); errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, gerror.Wrap(err, "query business license metadata")
	}
	return record, nil
}

func loadCompanyProfile(ctx context.Context, userID uint64) (*entity.CompanyProfiles, error) {
	var record *entity.CompanyProfiles
	if err := dao.CompanyProfiles.Ctx(ctx).Where(do.CompanyProfiles{UserId: userID}).Scan(&record); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, gerror.Wrap(err, "query company profile")
	}
	return record, nil
}

func loadCompanyContacts(ctx context.Context, userID uint64) ([]adminverificationmodel.Contact, error) {
	var records []*entity.CompanyContacts
	if err := dao.CompanyContacts.Ctx(ctx).Where(do.CompanyContacts{UserId: userID}).OrderAsc("id").Scan(&records); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, gerror.Wrap(err, "query company contacts")
	}
	contacts := make([]adminverificationmodel.Contact, 0, len(records))
	for _, record := range records {
		if record != nil {
			contacts = append(contacts, adminverificationmodel.Contact{Name: record.Name, Position: record.Position, Phone: record.Phone})
		}
	}
	return contacts, nil
}

func decodeProfile(profile *entity.CompanyProfiles) (offices, factories []string, introduction string, storePlatformID uint64, storeURL string) {
	offices, factories = []string{}, []string{}
	if profile == nil {
		return offices, factories, "", 0, ""
	}
	introduction = profile.IntroductionHtml
	storePlatformID = profile.StorePlatformId
	storeURL = profile.StoreUrl
	_ = json.Unmarshal([]byte(profile.Offices), &offices)
	_ = json.Unmarshal([]byte(profile.Factories), &factories)
	return offices, factories, introduction, storePlatformID, storeURL
}

func loadStorePlatformName(ctx context.Context, id uint64) (string, error) {
	if id == 0 {
		return "", nil
	}
	var record *entity.StorePlatforms
	if err := dao.StorePlatforms.Ctx(ctx).Fields(dao.StorePlatforms.Columns().Name).
		Where(do.StorePlatforms{Id: id}).Scan(&record); err != nil {
		return "", gerror.Wrap(err, "query store platform")
	}
	if record == nil {
		return "", nil
	}
	return record.Name, nil
}

func maskIDNumber(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 6 {
		return strings.Repeat("*", len(value))
	}
	return value[:3] + strings.Repeat("*", len(value)-7) + value[len(value)-4:]
}

func timePointer(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	copy := value
	return &copy
}
