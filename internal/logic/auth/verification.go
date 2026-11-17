package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/url"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	"github.com/suxiangdong/indu-scope-backend/internal/dao"
	authmodel "github.com/suxiangdong/indu-scope-backend/internal/model/auth"
	"github.com/suxiangdong/indu-scope-backend/internal/model/do"
	"github.com/suxiangdong/indu-scope-backend/internal/model/entity"
	storePlatformModel "github.com/suxiangdong/indu-scope-backend/internal/model/storeplatform"
	verificationModel "github.com/suxiangdong/indu-scope-backend/internal/model/verification"
	"github.com/suxiangdong/indu-scope-backend/internal/ocr"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
	"github.com/suxiangdong/indu-scope-backend/internal/storage"
)

const (
	businessLicensePurpose      = "business_license"
	idCardFrontPurpose          = "id_card_front"
	idCardBackPurpose           = "id_card_back"
	maxBusinessLicenseSize      = 10 * 1024 * 1024
	businessLicenseOCRCooldown  = 30 * time.Second
	businessLicenseOCRKeyPrefix = "indu_scope:ocr:business_license:"
)

func (s *sAuth) GetVerification(ctx context.Context) (verificationModel.Summary, error) {
	user, err := authenticatedUser(ctx)
	if err != nil {
		return verificationModel.Summary{}, err
	}
	return s.loadVerificationSummary(ctx, *user)
}

func (s *sAuth) ListStorePlatforms(ctx context.Context) ([]storePlatformModel.Item, error) {
	cols := dao.StorePlatforms.Columns()
	var records []*entity.StorePlatforms
	if err := dao.StorePlatforms.Ctx(ctx).Fields(cols.Id, cols.Name, cols.SortOrder, cols.Status, cols.CreatedAt, cols.UpdatedAt).
		Where(do.StorePlatforms{Status: 1}).OrderAsc(cols.SortOrder).OrderAsc(cols.Id).Scan(&records); err != nil {
		return nil, gerror.Wrap(err, "list active store platforms")
	}
	items := make([]storePlatformModel.Item, 0, len(records))
	for _, record := range records {
		if record != nil {
			items = append(items, storePlatformModel.Item{ID: record.Id, Name: record.Name, SortOrder: record.SortOrder, Status: uint8(record.Status), CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt})
		}
	}
	return items, nil
}

func (s *sAuth) SubmitPersonalVerification(ctx context.Context, input verificationModel.PersonalInput) (verificationModel.Summary, error) {
	user, err := authenticatedUser(ctx)
	if err != nil {
		return verificationModel.Summary{}, err
	}
	role, err := resolveVerificationRole(user.Role, input.Role)
	if err != nil {
		return verificationModel.Summary{}, err
	}
	if role == 2 {
		return verificationModel.Summary{}, &consts.InvalidParameterError{Message: "supplier users must submit company verification"}
	}
	input.RealName = strings.TrimSpace(input.RealName)
	input.IDNumber = strings.ToUpper(strings.TrimSpace(input.IDNumber))
	input.StoreURL = strings.TrimSpace(input.StoreURL)
	input.IDFrontBucket = strings.TrimSpace(input.IDFrontBucket)
	input.IDFrontPath = strings.TrimSpace(input.IDFrontPath)
	input.IDFrontEndpoint = strings.TrimSpace(input.IDFrontEndpoint)
	input.IDBackBucket = strings.TrimSpace(input.IDBackBucket)
	input.IDBackPath = strings.TrimSpace(input.IDBackPath)
	input.IDBackEndpoint = strings.TrimSpace(input.IDBackEndpoint)
	if utf8.RuneCountInString(input.RealName) < 2 || utf8.RuneCountInString(input.RealName) > 50 {
		return verificationModel.Summary{}, &consts.InvalidParameterError{Message: "real_name must contain 2 to 50 characters"}
	}
	if !validIDNumber(input.IDNumber) {
		return verificationModel.Summary{}, &consts.InvalidParameterError{Message: "id_number must be a valid 15 or 18 character identity number"}
	}
	if input.StorePlatformID == 0 {
		return verificationModel.Summary{}, &consts.InvalidParameterError{Message: "store_platform_id is required"}
	}
	if err := validateStoreURL(input.StoreURL); err != nil {
		return verificationModel.Summary{}, err
	}
	var platform *entity.StorePlatforms
	if err := dao.StorePlatforms.Ctx(ctx).Where(do.StorePlatforms{Id: input.StorePlatformID, Status: 1}).Scan(&platform); err != nil {
		return verificationModel.Summary{}, gerror.Wrap(err, "query store platform")
	}
	if platform == nil {
		return verificationModel.Summary{}, consts.ErrStorePlatformNotFound
	}
	replaced, err := persistPersonalVerification(ctx, dao.Users.DB(), uint64(user.ID), role, input)
	if err != nil {
		return verificationModel.Summary{}, err
	}
	cleanupReplacedVerificationFiles(ctx, uint64(user.ID), replaced)
	user.Role, user.UserType, user.Status = role, 1, 0
	return s.loadVerificationSummary(ctx, *user)
}

func (s *sAuth) SubmitCompanyVerification(ctx context.Context, input verificationModel.CompanyInput) (verificationModel.Summary, error) {
	user, err := authenticatedUser(ctx)
	if err != nil {
		return verificationModel.Summary{}, err
	}
	role, err := resolveVerificationRole(user.Role, input.Role)
	if err != nil {
		return verificationModel.Summary{}, err
	}
	input.CompanyName = strings.TrimSpace(input.CompanyName)
	input.UnifiedSocialCreditCode = strings.ToUpper(strings.TrimSpace(input.UnifiedSocialCreditCode))
	input.BusinessLicenseBucket = strings.TrimSpace(input.BusinessLicenseBucket)
	input.BusinessLicensePath = strings.TrimSpace(input.BusinessLicensePath)
	input.BusinessLicenseEndpoint = strings.TrimSpace(input.BusinessLicenseEndpoint)
	input.RegisteredAddress = strings.TrimSpace(input.RegisteredAddress)
	input.Introduction = strings.TrimSpace(input.Introduction)
	input.StoreURL = strings.TrimSpace(input.StoreURL)
	input.Offices = normalizeCompanyLocations(input.Offices)
	input.Factories = normalizeCompanyLocations(input.Factories)
	for index := range input.Contacts {
		input.Contacts[index].Name = strings.TrimSpace(input.Contacts[index].Name)
		input.Contacts[index].Position = strings.TrimSpace(input.Contacts[index].Position)
		input.Contacts[index].Phone = strings.TrimSpace(input.Contacts[index].Phone)
	}
	if utf8.RuneCountInString(input.CompanyName) < 2 || utf8.RuneCountInString(input.CompanyName) > 255 {
		return verificationModel.Summary{}, &consts.InvalidParameterError{Message: "company_name must contain 2 to 255 characters"}
	}
	if len(input.UnifiedSocialCreditCode) != 18 || !validUpperAlphaNumeric(input.UnifiedSocialCreditCode) {
		return verificationModel.Summary{}, &consts.InvalidParameterError{Message: "unified_social_credit_code must contain 18 letters or digits"}
	}
	if utf8.RuneCountInString(input.RegisteredAddress) < 5 || utf8.RuneCountInString(input.RegisteredAddress) > 500 {
		return verificationModel.Summary{}, &consts.InvalidParameterError{Message: "registered_address must contain 5 to 500 characters"}
	}
	if utf8.RuneCountInString(input.Introduction) > 5000 {
		return verificationModel.Summary{}, &consts.InvalidParameterError{Message: "introduction must contain at most 5000 characters"}
	}
	if len(input.Offices) > 20 || len(input.Factories) > 20 {
		return verificationModel.Summary{}, &consts.InvalidParameterError{Message: "offices and factories can contain at most 20 locations each"}
	}
	if err := validateCompanyLocations(input.Offices, "offices"); err != nil {
		return verificationModel.Summary{}, err
	}
	if err := validateCompanyLocations(input.Factories, "factories"); err != nil {
		return verificationModel.Summary{}, err
	}
	if role == 1 {
		if len(input.Factories) > 0 {
			return verificationModel.Summary{}, &consts.InvalidParameterError{Message: "buyer verification must not include factory locations"}
		}
		if len(input.Offices) == 0 {
			return verificationModel.Summary{}, &consts.InvalidParameterError{Message: "at least one office location is required for buyer verification"}
		}
		if input.StorePlatformID == 0 {
			return verificationModel.Summary{}, &consts.InvalidParameterError{Message: "store_platform_id is required for buyer verification"}
		}
		if err := validateStoreURL(input.StoreURL); err != nil {
			return verificationModel.Summary{}, err
		}
		var platform *entity.StorePlatforms
		if err := dao.StorePlatforms.Ctx(ctx).Where(do.StorePlatforms{Id: input.StorePlatformID, Status: 1}).Scan(&platform); err != nil {
			return verificationModel.Summary{}, gerror.Wrap(err, "query store platform")
		}
		if platform == nil {
			return verificationModel.Summary{}, consts.ErrStorePlatformNotFound
		}
	} else {
		if len(input.Offices) == 0 || len(input.Factories) == 0 {
			return verificationModel.Summary{}, &consts.InvalidParameterError{Message: "supplier verification requires at least one office and one factory location"}
		}
		input.StorePlatformID = 0
		input.StoreURL = ""
	}
	if err := validateCompanyContacts(input.Contacts); err != nil {
		return verificationModel.Summary{}, err
	}
	replaced, err := persistCompanyVerification(ctx, dao.Users.DB(), uint64(user.ID), role, input)
	if err != nil {
		return verificationModel.Summary{}, err
	}
	cleanupReplacedVerificationFiles(ctx, uint64(user.ID), replaced)
	user.Role, user.UserType, user.Status = role, 2, 0
	return s.loadVerificationSummary(ctx, *user)
}

func (s *sAuth) UploadBusinessLicense(ctx context.Context, input verificationModel.BusinessLicenseUploadInput) (verificationModel.BusinessLicenseInfo, error) {
	user, err := authenticatedUser(ctx)
	if err != nil {
		return verificationModel.BusinessLicenseInfo{}, err
	}
	if user.Status == 1 {
		return verificationModel.BusinessLicenseInfo{}, consts.ErrVerificationLocked
	}
	personal, err := loadPersonalVerification(ctx, uint64(user.ID))
	if err != nil {
		return verificationModel.BusinessLicenseInfo{}, err
	} else if personal != nil {
		return verificationModel.BusinessLicenseInfo{}, consts.ErrVerificationTypeConflict
	}
	if company, err := loadCompanyVerification(ctx, uint64(user.ID)); err != nil {
		return verificationModel.BusinessLicenseInfo{}, err
	} else if company != nil && (company.ReviewStatus != uint(verificationModel.ReviewStatusRejected) || !company.VerifiedAt.IsZero()) {
		return verificationModel.BusinessLicenseInfo{}, consts.ErrVerificationLocked
	}
	if input.Reader == nil {
		return verificationModel.BusinessLicenseInfo{}, &consts.InvalidParameterError{Message: "business license file is required"}
	}
	if input.Size <= 0 || input.Size > maxBusinessLicenseSize {
		return verificationModel.BusinessLicenseInfo{}, &consts.InvalidParameterError{Message: "business license file must be between 1 byte and 10 MB"}
	}
	contentType := strings.ToLower(strings.TrimSpace(input.ContentType))
	if !allowedBusinessLicenseContentType(contentType) {
		return verificationModel.BusinessLicenseInfo{}, &consts.InvalidParameterError{Message: "business license file must be a JPG, PNG or PDF"}
	}
	filename := filepath.Base(strings.TrimSpace(input.Filename))
	if filename == "." || filename == "" {
		filename = "business-license" + extensionForContentType(contentType)
	}
	if filepath.Ext(filename) == "" {
		filename += extensionForContentType(contentType)
	}
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return verificationModel.BusinessLicenseInfo{}, gerror.Wrap(err, "generate business license object key")
	}
	client, err := storage.OSSClient()
	if err != nil {
		return verificationModel.BusinessLicenseInfo{}, gerror.Wrap(err, "initialize oss client")
	}
	objectKey := client.ObjectKey(businessLicensePurpose, fmt.Sprintf("%d-%d-%s%s", user.ID, time.Now().UnixNano(), hex.EncodeToString(nonce), extensionForContentType(contentType)))
	etag, err := client.Upload(ctx, objectKey, input.Reader, input.Size, contentType)
	if err != nil {
		_ = storage.DeleteWithTimeout(ctx, client, objectKey)
		return verificationModel.BusinessLicenseInfo{}, gerror.Wrap(err, "upload business license to oss")
	}
	if _, err := dao.UserFiles.Ctx(ctx).Data(g.Map{
		"user_id":       user.ID,
		"purpose":       businessLicensePurpose,
		"endpoint":      client.EndpointHost(),
		"bucket":        client.Bucket(),
		"object_key":    objectKey,
		"original_name": filename,
		"content_type":  contentType,
		"file_size":     input.Size,
		"etag":          etag,
	}).Insert(); err != nil {
		_ = storage.DeleteWithTimeout(ctx, client, objectKey)
		return verificationModel.BusinessLicenseInfo{}, gerror.Wrap(err, "save business license metadata")
	}
	return verificationModel.BusinessLicenseInfo{
		Endpoint: client.EndpointHost(), Bucket: client.Bucket(), Path: objectKey,
		Filename: filename, ContentType: contentType, Size: input.Size,
	}, nil
}

func (s *sAuth) RecognizeBusinessLicense(ctx context.Context) (verificationModel.BusinessLicenseOCR, error) {
	user, err := authenticatedUser(ctx)
	if err != nil {
		return verificationModel.BusinessLicenseOCR{}, err
	}
	if user.Status == 1 {
		return verificationModel.BusinessLicenseOCR{}, consts.ErrVerificationLocked
	}
	var record *userFileRecord
	err = dao.UserFiles.Ctx(ctx).
		Where("user_id", user.ID).
		Where("purpose", businessLicensePurpose).
		OrderDesc("id").Limit(1).Scan(&record)
	if errors.Is(err, sql.ErrNoRows) {
		record = nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return verificationModel.BusinessLicenseOCR{}, gerror.Wrap(err, "query business license metadata")
	}
	if record == nil {
		return verificationModel.BusinessLicenseOCR{}, &consts.InvalidParameterError{Message: "business license file must be uploaded before recognition"}
	}
	if record.ContentType == "application/pdf" {
		return verificationModel.BusinessLicenseOCR{}, &consts.InvalidParameterError{Message: "business license OCR currently supports image files only"}
	}
	ossClient, err := storage.OSSClient()
	if err != nil {
		return verificationModel.BusinessLicenseOCR{}, gerror.Wrap(err, "initialize oss client")
	}
	imageURL, err := ossClient.SignedURLFor(record.Endpoint, record.Bucket, record.ObjectKey, 10*time.Minute)
	if err != nil {
		return verificationModel.BusinessLicenseOCR{}, gerror.Wrap(err, "create business license OCR URL")
	}
	if err := reserveBusinessLicenseOCR(ctx, uint64(user.ID)); err != nil {
		return verificationModel.BusinessLicenseOCR{}, err
	}
	result, err := ocr.RecognizeBusinessLicense(ctx, imageURL)
	if err != nil {
		return verificationModel.BusinessLicenseOCR{}, gerror.Wrap(err, "recognize business license")
	}
	return verificationModel.BusinessLicenseOCR{
		CompanyName:             result.CompanyName,
		UnifiedSocialCreditCode: result.UnifiedSocialCreditCode,
		BusinessAddress:         result.BusinessAddress,
	}, nil
}

func reserveBusinessLicenseOCR(ctx context.Context, userID uint64) error {
	if userID == 0 {
		return consts.ErrTokenInvalid
	}
	key := fmt.Sprintf("%s%d", businessLicenseOCRKeyPrefix, userID)
	result, err := g.Redis().Do(ctx, "SET", key, "1", "NX", "EX", int(businessLicenseOCRCooldown.Seconds()))
	if err != nil {
		return gerror.Wrap(err, "reserve business license OCR request")
	}
	if result == nil || !strings.EqualFold(result.String(), "OK") {
		return &consts.InvalidParameterError{Message: "营业执照识别请求过于频繁，请稍后再试"}
	}
	return nil
}

func (s *sAuth) UploadIDCard(ctx context.Context, input verificationModel.IDCardUploadInput) (verificationModel.IDCardUploadInfo, error) {
	user, err := authenticatedUser(ctx)
	if err != nil {
		return verificationModel.IDCardUploadInfo{}, err
	}
	if user.Role == 2 {
		return verificationModel.IDCardUploadInfo{}, &consts.InvalidParameterError{Message: "supplier users cannot submit personal verification"}
	}
	if user.Status == 1 {
		return verificationModel.IDCardUploadInfo{}, consts.ErrVerificationLocked
	}
	input.Side = strings.ToLower(strings.TrimSpace(input.Side))
	purpose := ""
	switch input.Side {
	case "front":
		purpose = idCardFrontPurpose
	case "back":
		purpose = idCardBackPurpose
	default:
		return verificationModel.IDCardUploadInfo{}, &consts.InvalidParameterError{Message: "side must be front or back"}
	}
	if company, err := loadCompanyVerification(ctx, uint64(user.ID)); err != nil {
		return verificationModel.IDCardUploadInfo{}, err
	} else if company != nil {
		return verificationModel.IDCardUploadInfo{}, consts.ErrVerificationTypeConflict
	}
	if personal, err := loadPersonalVerification(ctx, uint64(user.ID)); err != nil {
		return verificationModel.IDCardUploadInfo{}, err
	} else if personal != nil && (personal.ReviewStatus != uint(verificationModel.ReviewStatusRejected) || !personal.VerifiedAt.IsZero()) {
		return verificationModel.IDCardUploadInfo{}, consts.ErrVerificationLocked
	}
	if input.Reader == nil {
		return verificationModel.IDCardUploadInfo{}, &consts.InvalidParameterError{Message: "identity document file is required"}
	}
	if input.Size <= 0 || input.Size > maxBusinessLicenseSize {
		return verificationModel.IDCardUploadInfo{}, &consts.InvalidParameterError{Message: "identity document file must be between 1 byte and 10 MB"}
	}
	contentType := strings.ToLower(strings.TrimSpace(input.ContentType))
	if contentType != "image/jpeg" && contentType != "image/png" {
		return verificationModel.IDCardUploadInfo{}, &consts.InvalidParameterError{Message: "identity document file must be a JPG or PNG image"}
	}
	filename := filepath.Base(strings.TrimSpace(input.Filename))
	if filename == "." || filename == "" {
		filename = "id-card-" + input.Side + extensionForContentType(contentType)
	}
	if filepath.Ext(filename) == "" {
		filename += extensionForContentType(contentType)
	}
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return verificationModel.IDCardUploadInfo{}, gerror.Wrap(err, "generate identity document object key")
	}
	client, err := storage.OSSClient()
	if err != nil {
		return verificationModel.IDCardUploadInfo{}, gerror.Wrap(err, "initialize oss client")
	}
	objectKey := client.ObjectKey("id_card/"+input.Side, fmt.Sprintf("%d-%d-%s%s", user.ID, time.Now().UnixNano(), hex.EncodeToString(nonce), extensionForContentType(contentType)))
	etag, err := client.Upload(ctx, objectKey, input.Reader, input.Size, contentType)
	if err != nil {
		_ = storage.DeleteWithTimeout(ctx, client, objectKey)
		return verificationModel.IDCardUploadInfo{}, gerror.Wrap(err, "upload identity document to oss")
	}
	if _, err := dao.UserFiles.Ctx(ctx).Data(g.Map{
		"user_id": user.ID, "purpose": purpose, "endpoint": client.EndpointHost(), "bucket": client.Bucket(),
		"object_key": objectKey, "original_name": filename, "content_type": contentType,
		"file_size": input.Size, "etag": etag,
	}).Insert(); err != nil {
		_ = storage.DeleteWithTimeout(ctx, client, objectKey)
		return verificationModel.IDCardUploadInfo{}, gerror.Wrap(err, "save identity document metadata")
	}
	return verificationModel.IDCardUploadInfo{
		Side: input.Side, Endpoint: client.EndpointHost(), Bucket: client.Bucket(), Path: objectKey,
		Filename: filename, ContentType: contentType, Size: input.Size,
	}, nil
}

func authenticatedUser(ctx context.Context) (*authmodel.User, error) {
	value := service.Bizctx().Get(ctx)
	if value == nil || value.User == nil {
		return nil, gerror.New("authorization is required")
	}
	return value.User, nil
}

func resolveVerificationRole(currentRole, requestedRole uint8) (uint8, error) {
	if currentRole == 0 {
		if requestedRole != 1 && requestedRole != 2 {
			return 0, &consts.InvalidParameterError{Message: "role must be selected as 1 (buyer) or 2 (supplier)"}
		}
		return requestedRole, nil
	}
	if currentRole != 1 && currentRole != 2 {
		return 0, &consts.InvalidParameterError{Message: "current role is invalid"}
	}
	if requestedRole != 0 && requestedRole != currentRole {
		return 0, &consts.InvalidParameterError{Message: "role cannot be changed after selection"}
	}
	return currentRole, nil
}

func (s *sAuth) loadVerificationSummary(ctx context.Context, user authmodel.User) (verificationModel.Summary, error) {
	personal, err := loadPersonalVerification(ctx, uint64(user.ID))
	if err != nil {
		return verificationModel.Summary{}, err
	}
	company, err := loadCompanyVerification(ctx, uint64(user.ID))
	if err != nil {
		return verificationModel.Summary{}, err
	}
	profile, err := loadCompanyProfile(ctx, uint64(user.ID))
	if err != nil {
		return verificationModel.Summary{}, err
	}
	contacts, err := loadCompanyContacts(ctx, uint64(user.ID))
	if err != nil {
		return verificationModel.Summary{}, err
	}
	status := "unsubmitted"
	reviewReason := ""
	if user.Status == 1 || (personal != nil && (personal.ReviewStatus == uint(verificationModel.ReviewStatusApproved) || !personal.VerifiedAt.IsZero())) || (company != nil && (company.ReviewStatus == uint(verificationModel.ReviewStatusApproved) || !company.VerifiedAt.IsZero())) {
		status = "verified"
	} else if personal != nil && personal.ReviewStatus == uint(verificationModel.ReviewStatusRejected) {
		status = "rejected"
		reviewReason = personal.ReviewReason
	} else if company != nil && company.ReviewStatus == uint(verificationModel.ReviewStatusRejected) {
		status = "rejected"
		reviewReason = company.ReviewReason
	} else if personal != nil || company != nil {
		status = "pending"
	}
	result := verificationModel.Summary{Role: uint8(user.Role), UserType: uint8(user.UserType), Status: status, ReviewReason: reviewReason}
	if personal != nil {
		storePlatformName, err := loadStorePlatformName(ctx, personal.StorePlatformId)
		if err != nil {
			return verificationModel.Summary{}, err
		}
		idNumberMasked := maskIDNumber(personal.IdNumber)
		if idNumberMasked == "" && len(personal.IdNumberEncrypted) > 0 {
			idNumberMasked = "已提交"
		}
		result.Personal = &verificationModel.PersonalInfo{
			RealName: personal.RealName, IDNumberMasked: idNumberMasked,
			StorePlatformID: personal.StorePlatformId, StorePlatformName: storePlatformName, StoreURL: personal.StoreUrl,
			IDFrontEndpoint: personal.IdFrontEndpoint,
			IDFrontBucket:   personal.IdFrontBucket, IDFrontPath: personal.IdFrontPath,
			IDBackEndpoint: personal.IdBackEndpoint,
			IDBackBucket:   personal.IdBackBucket, IDBackPath: personal.IdBackPath,
			VerifiedAt:   formatVerificationTime(personal.VerifiedAt),
			ReviewStatus: uint8(personal.ReviewStatus), ReviewReason: personal.ReviewReason,
			ReviewedAt: formatVerificationTime(personal.ReviewedAt),
		}
	}
	if company != nil {
		offices, factories := []string{}, []string{}
		introduction := ""
		storePlatformID := uint64(0)
		storeURL := ""
		if profile != nil {
			introduction = profile.IntroductionHtml
			storePlatformID = profile.StorePlatformId
			storeURL = profile.StoreUrl
			if err := json.Unmarshal([]byte(profile.Offices), &offices); err != nil {
				return verificationModel.Summary{}, gerror.Wrap(err, "decode company offices")
			}
			if err := json.Unmarshal([]byte(profile.Factories), &factories); err != nil {
				return verificationModel.Summary{}, gerror.Wrap(err, "decode company factories")
			}
		}
		contactItems := make([]verificationModel.ContactInfo, 0, len(contacts))
		for _, contact := range contacts {
			contactItems = append(contactItems, verificationModel.ContactInfo{Name: contact.Name, Position: contact.Position, Phone: contact.Phone})
		}
		businessLicenseFile, err := loadBusinessLicenseFile(ctx, uint64(user.ID), company.BusinessLicenseEndpoint, company.BusinessLicenseBucket, company.BusinessLicensePath)
		if err != nil {
			return verificationModel.Summary{}, err
		}
		storePlatformName, err := loadStorePlatformName(ctx, storePlatformID)
		if err != nil {
			return verificationModel.Summary{}, err
		}
		companyInfo := &verificationModel.CompanyInfo{
			CompanyName: company.CompanyName, UnifiedSocialCreditCode: company.UnifiedSocialCreditCode,
			BusinessLicenseEndpoint: company.BusinessLicenseEndpoint,
			BusinessLicenseBucket:   company.BusinessLicenseBucket, BusinessLicensePath: company.BusinessLicensePath,
			RegisteredAddress: company.RegisteredAddress,
			Introduction:      introduction, StorePlatformID: storePlatformID, StorePlatformName: storePlatformName, StoreURL: storeURL,
			Offices: offices, Factories: factories, Contacts: contactItems,
			VerifiedAt:   formatVerificationTime(company.VerifiedAt),
			ReviewStatus: uint8(company.ReviewStatus), ReviewReason: company.ReviewReason,
			ReviewedAt: formatVerificationTime(company.ReviewedAt),
		}
		if businessLicenseFile != nil {
			companyInfo.BusinessLicenseEndpoint = businessLicenseFile.Endpoint
			companyInfo.BusinessLicenseBucket = businessLicenseFile.Bucket
			companyInfo.BusinessLicensePath = businessLicenseFile.ObjectKey
			companyInfo.BusinessLicenseFileName = businessLicenseFile.OriginalName
			companyInfo.BusinessLicenseFileSize = int64(businessLicenseFile.FileSize)
			companyInfo.BusinessLicenseContentType = businessLicenseFile.ContentType
		}
		result.Company = companyInfo
	}
	return result, nil
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

type userFileRecord struct {
	ID           uint64 `orm:"id"`
	UserID       uint64 `orm:"user_id"`
	Purpose      string `orm:"purpose"`
	Endpoint     string `orm:"endpoint"`
	Bucket       string `orm:"bucket"`
	ObjectKey    string `orm:"object_key"`
	OriginalName string `orm:"original_name"`
	ContentType  string `orm:"content_type"`
	FileSize     uint64 `orm:"file_size"`
	ETag         string `orm:"etag"`
}

func loadBusinessLicenseFile(ctx context.Context, userID uint64, endpoint, bucket, objectPath string) (*userFileRecord, error) {
	endpoint = strings.TrimSpace(endpoint)
	bucket = strings.TrimSpace(bucket)
	objectPath = strings.TrimSpace(objectPath)
	var record *userFileRecord
	model := dao.UserFiles.Ctx(ctx).
		Where("user_id", userID).
		Where("purpose", businessLicensePurpose)
	if endpoint != "" && bucket != "" && objectPath != "" {
		model = model.Where("endpoint", endpoint).Where("bucket", bucket).Where("object_key", objectPath)
	} else if bucket != "" && objectPath != "" {
		model = model.Where("bucket", bucket).Where("object_key", objectPath)
	} else {
		return nil, nil
	}
	err := model.Scan(&record)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, gerror.Wrap(err, "query business license metadata")
	}
	return record, nil
}

func allowedBusinessLicenseContentType(contentType string) bool {
	switch contentType {
	case "image/jpeg", "image/png", "application/pdf":
		return true
	default:
		return false
	}
}

func extensionForContentType(contentType string) string {
	if extensions, _ := mime.ExtensionsByType(contentType); len(extensions) > 0 {
		return extensions[0]
	}
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "application/pdf":
		return ".pdf"
	default:
		return ""
	}
}

func loadPersonalVerification(ctx context.Context, userID uint64) (*entity.PersonalVerifications, error) {
	var record *entity.PersonalVerifications
	err := dao.PersonalVerifications.Ctx(ctx).Where(do.PersonalVerifications{UserId: userID}).Scan(&record)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, gerror.Wrap(err, "query personal verification")
	}
	return record, nil
}

func loadCompanyVerification(ctx context.Context, userID uint64) (*entity.CompanyVerifications, error) {
	var record *entity.CompanyVerifications
	err := dao.CompanyVerifications.Ctx(ctx).Where(do.CompanyVerifications{UserId: userID}).Scan(&record)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, gerror.Wrap(err, "query company verification")
	}
	return record, nil
}

func loadCompanyProfile(ctx context.Context, userID uint64) (*entity.CompanyProfiles, error) {
	var record *entity.CompanyProfiles
	err := dao.CompanyProfiles.Ctx(ctx).Where(do.CompanyProfiles{UserId: userID}).Scan(&record)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, gerror.Wrap(err, "query company profile")
	}
	return record, nil
}

func loadCompanyContacts(ctx context.Context, userID uint64) ([]*entity.CompanyContacts, error) {
	var records []*entity.CompanyContacts
	err := dao.CompanyContacts.Ctx(ctx).Where(do.CompanyContacts{UserId: userID}).OrderAsc("id").Scan(&records)
	if errors.Is(err, sql.ErrNoRows) {
		return []*entity.CompanyContacts{}, nil
	}
	if err != nil {
		return nil, gerror.Wrap(err, "query company contacts")
	}
	return records, nil
}

func saveCompanyProfileTx(ctx context.Context, tx gdb.TX, userID uint64, input verificationModel.CompanyInput) error {
	offices, err := json.Marshal(input.Offices)
	if err != nil {
		return gerror.Wrap(err, "encode company offices")
	}
	factories, err := json.Marshal(input.Factories)
	if err != nil {
		return gerror.Wrap(err, "encode company factories")
	}
	var record *entity.CompanyProfiles
	if err := tx.Model("company_profiles").Where("user_id", userID).Scan(&record); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return gerror.Wrap(err, "query company profile")
	}
	var storePlatformID any
	if input.StorePlatformID > 0 {
		storePlatformID = input.StorePlatformID
	}
	data := g.Map{
		"user_id": userID, "introduction_html": input.Introduction,
		"store_platform_id": storePlatformID, "store_url": input.StoreURL,
		"offices": string(offices), "factories": string(factories),
	}
	if record == nil {
		if _, err := tx.Model("company_profiles").Data(data).Insert(); err != nil {
			return gerror.Wrap(err, "create company profile")
		}
	} else if _, err := tx.Model("company_profiles").Where("user_id", userID).Data(data).Update(); err != nil {
		return gerror.Wrap(err, "update company profile")
	}
	return nil
}

func replaceCompanyContactsTx(tx gdb.TX, userID uint64, contacts []verificationModel.ContactInput) error {
	if _, err := tx.Model("company_contacts").Where("user_id", userID).Delete(); err != nil {
		return gerror.Wrap(err, "clear company contacts")
	}
	for _, contact := range contacts {
		if _, err := tx.Model("company_contacts").Data(g.Map{
			"user_id": userID, "name": contact.Name, "position": contact.Position, "phone": contact.Phone,
		}).Insert(); err != nil {
			return gerror.Wrap(err, "create company contact")
		}
	}
	return nil
}

func normalizeCompanyLocations(items []string) []string {
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			if _, exists := seen[item]; exists {
				continue
			}
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func validateCompanyLocations(locations []string, field string) error {
	for _, location := range locations {
		if utf8.RuneCountInString(location) < 5 || utf8.RuneCountInString(location) > 500 {
			return &consts.InvalidParameterError{Message: field + " entries must contain 5 to 500 characters"}
		}
	}
	return nil
}

func validateCompanyContacts(contacts []verificationModel.ContactInput) error {
	if len(contacts) == 0 || len(contacts) > 10 {
		return &consts.InvalidParameterError{Message: "contacts must contain between 1 and 10 items"}
	}
	seenPhones := make(map[string]struct{}, len(contacts))
	for _, contact := range contacts {
		if utf8.RuneCountInString(contact.Name) < 2 || utf8.RuneCountInString(contact.Name) > 100 {
			return &consts.InvalidParameterError{Message: "contact name must contain 2 to 100 characters"}
		}
		if utf8.RuneCountInString(contact.Position) > 100 {
			return &consts.InvalidParameterError{Message: "contact position must contain at most 100 characters"}
		}
		if len(contact.Phone) < 6 || len(contact.Phone) > 20 {
			return &consts.InvalidParameterError{Message: "contact phone must contain 6 to 20 digits"}
		}
		for _, char := range contact.Phone {
			if char < '0' || char > '9' {
				return &consts.InvalidParameterError{Message: "contact phone must contain digits only"}
			}
		}
		if _, exists := seenPhones[contact.Phone]; exists {
			return &consts.InvalidParameterError{Message: "contact phone cannot be duplicated"}
		}
		seenPhones[contact.Phone] = struct{}{}
	}
	return nil
}

func validIDNumber(value string) bool {
	if len(value) != 15 && len(value) != 18 {
		return false
	}
	for index := 0; index < len(value); index++ {
		if value[index] < '0' || value[index] > '9' {
			if !(len(value) == 18 && index == 17 && value[index] == 'X') {
				return false
			}
		}
	}
	date := value[6:14]
	if len(value) == 15 {
		date = "19" + value[6:12]
	}
	if _, err := time.Parse("20060102", date); err != nil {
		return false
	}
	if len(value) == 15 {
		return true
	}
	weights := [...]int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	checks := "10X98765432"
	sum := 0
	for index, weight := range weights {
		sum += int(value[index]-'0') * weight
	}
	return value[17] == checks[sum%11]
}

func validateStoreURL(value string) error {
	if value == "" || utf8.RuneCountInString(value) > 500 {
		return &consts.InvalidParameterError{Message: "store_url must contain 1 to 500 characters"}
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return &consts.InvalidParameterError{Message: "store_url must be a valid http or https URL"}
	}
	return nil
}

func validUpperAlphaNumeric(value string) bool {
	for _, char := range value {
		if (char < '0' || char > '9') && (char < 'A' || char > 'Z') {
			return false
		}
	}
	return true
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

func formatVerificationTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339)
}
