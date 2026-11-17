package auth

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/suxiangdong/indu-scope-backend/api/auth/v1"
	verificationModel "github.com/suxiangdong/indu-scope-backend/internal/model/verification"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) GetVerification(ctx context.Context, _ *v1.GetVerificationReq) (*v1.GetVerificationRes, error) {
	if err := requireAuthenticatedUser(ctx); err != nil {
		return nil, err
	}
	result, err := service.Auth().GetVerification(ctx)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.GetVerificationRes{VerificationInfo: toAPIVerification(result)}, nil
}

func (c *ControllerV1) ListStorePlatforms(ctx context.Context, _ *v1.ListStorePlatformsReq) (*v1.ListStorePlatformsRes, error) {
	result, err := service.Auth().ListStorePlatforms(ctx)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	items := make([]v1.StorePlatformInfo, 0, len(result))
	for _, item := range result {
		items = append(items, v1.StorePlatformInfo{ID: item.ID, Name: item.Name})
	}
	return &v1.ListStorePlatformsRes{Items: items}, nil
}

func (c *ControllerV1) UploadBusinessLicense(ctx context.Context, req *v1.UploadBusinessLicenseReq) (*v1.UploadBusinessLicenseRes, error) {
	if err := requireAuthenticatedUser(ctx); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "business license file is required")
	}
	upload := req.File
	if upload == nil {
		upload = g.RequestFromCtx(ctx).GetUploadFile("file")
	}
	if upload == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "business license file is required")
	}
	file, err := upload.Open()
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "business license file could not be opened")
	}
	defer file.Close()
	head := make([]byte, 512)
	read, readErr := io.ReadFull(file, head)
	if readErr != nil && readErr != io.ErrUnexpectedEOF && readErr != io.EOF {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "business license file could not be read")
	}
	if read == 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "business license file is empty")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "business license file could not be read")
	}
	contentType := strings.ToLower(http.DetectContentType(head[:read]))
	result, err := service.Auth().UploadBusinessLicense(ctx, verificationModel.BusinessLicenseUploadInput{
		Reader: file, Filename: upload.Filename, ContentType: contentType, Size: upload.Size,
	})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.UploadBusinessLicenseRes{BusinessLicenseUploadInfo: v1.BusinessLicenseUploadInfo{
		Endpoint: result.Endpoint, Bucket: result.Bucket, Path: result.Path,
		Filename: result.Filename, ContentType: result.ContentType, Size: result.Size,
	}}, nil
}

func (c *ControllerV1) RecognizeBusinessLicense(ctx context.Context, _ *v1.RecognizeBusinessLicenseReq) (*v1.RecognizeBusinessLicenseRes, error) {
	if err := requireAuthenticatedUser(ctx); err != nil {
		return nil, err
	}
	result, err := service.Auth().RecognizeBusinessLicense(ctx)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.RecognizeBusinessLicenseRes{
		CompanyName:             result.CompanyName,
		UnifiedSocialCreditCode: result.UnifiedSocialCreditCode,
		BusinessAddress:         result.BusinessAddress,
	}, nil
}

func (c *ControllerV1) UploadIDCard(ctx context.Context, req *v1.UploadIDCardReq) (*v1.UploadIDCardRes, error) {
	if err := requireAuthenticatedUser(ctx); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "identity document side and file are required")
	}
	upload := req.File
	if upload == nil {
		upload = g.RequestFromCtx(ctx).GetUploadFile("file")
	}
	if upload == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "identity document file is required")
	}
	file, err := upload.Open()
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "identity document file could not be opened")
	}
	defer file.Close()
	head := make([]byte, 512)
	read, readErr := io.ReadFull(file, head)
	if readErr != nil && readErr != io.ErrUnexpectedEOF && readErr != io.EOF {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "identity document file could not be read")
	}
	if read == 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "identity document file is empty")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "identity document file could not be read")
	}
	contentType := strings.ToLower(http.DetectContentType(head[:read]))
	result, err := service.Auth().UploadIDCard(ctx, verificationModel.IDCardUploadInput{
		Side: req.Side, Reader: file, Filename: upload.Filename, ContentType: contentType, Size: upload.Size,
	})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.UploadIDCardRes{IDCardUploadInfo: v1.IDCardUploadInfo{
		Side: result.Side, Endpoint: result.Endpoint, Bucket: result.Bucket, Path: result.Path,
		Filename: result.Filename, ContentType: result.ContentType, Size: result.Size,
	}}, nil
}

func (c *ControllerV1) SubmitPersonalVerification(ctx context.Context, req *v1.SubmitPersonalVerificationReq) (*v1.SubmitPersonalVerificationRes, error) {
	if err := requireAuthenticatedUser(ctx); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "personal verification request is required")
	}
	result, err := service.Auth().SubmitPersonalVerification(ctx, verificationModel.PersonalInput{
		Role: req.Role, RealName: req.RealName, IDNumber: req.IDNumber, StorePlatformID: req.StorePlatformID, StoreURL: req.StoreURL,
		IDFrontEndpoint: req.IDFrontEndpoint, IDFrontBucket: req.IDFrontBucket, IDFrontPath: req.IDFrontPath,
		IDBackEndpoint: req.IDBackEndpoint, IDBackBucket: req.IDBackBucket, IDBackPath: req.IDBackPath,
	})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.SubmitPersonalVerificationRes{VerificationInfo: toAPIVerification(result)}, nil
}

func (c *ControllerV1) SubmitCompanyVerification(ctx context.Context, req *v1.SubmitCompanyVerificationReq) (*v1.SubmitCompanyVerificationRes, error) {
	if err := requireAuthenticatedUser(ctx); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "company verification request is required")
	}
	result, err := service.Auth().SubmitCompanyVerification(ctx, verificationModel.CompanyInput{
		Role: req.Role, CompanyName: req.CompanyName, UnifiedSocialCreditCode: req.UnifiedSocialCreditCode,
		BusinessLicenseEndpoint: req.BusinessLicenseEndpoint,
		BusinessLicenseBucket:   req.BusinessLicenseBucket, BusinessLicensePath: req.BusinessLicensePath,
		RegisteredAddress: req.RegisteredAddress,
		Introduction:      req.Introduction, StorePlatformID: req.StorePlatformID, StoreURL: req.StoreURL,
		Offices: req.Offices, Factories: req.Factories,
		Contacts: toContactInputs(req.Contacts),
	})
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.SubmitCompanyVerificationRes{VerificationInfo: toAPIVerification(result)}, nil
}

func requireAuthenticatedUser(ctx context.Context) error {
	if value := service.Bizctx().Get(ctx); value != nil && value.User != nil {
		return nil
	}
	return gerror.NewCode(gcode.CodeNotAuthorized, "authorization is required")
}

func toAPIVerification(result verificationModel.Summary) v1.VerificationInfo {
	info := v1.VerificationInfo{Role: result.Role, UserType: result.UserType, Status: result.Status, ReviewReason: result.ReviewReason}
	if result.Personal != nil {
		info.Personal = &v1.PersonalVerificationInfo{
			RealName: result.Personal.RealName, IDNumberMasked: result.Personal.IDNumberMasked,
			StorePlatformID: result.Personal.StorePlatformID, StorePlatformName: result.Personal.StorePlatformName, StoreURL: result.Personal.StoreURL,
			IDFrontEndpoint: result.Personal.IDFrontEndpoint,
			IDFrontBucket:   result.Personal.IDFrontBucket, IDFrontPath: result.Personal.IDFrontPath,
			IDBackEndpoint: result.Personal.IDBackEndpoint,
			IDBackBucket:   result.Personal.IDBackBucket, IDBackPath: result.Personal.IDBackPath,
			VerifiedAt: result.Personal.VerifiedAt, ReviewStatus: result.Personal.ReviewStatus,
			ReviewReason: result.Personal.ReviewReason, ReviewedAt: result.Personal.ReviewedAt,
		}
	}
	if result.Company != nil {
		info.Company = &v1.CompanyVerificationInfo{
			CompanyName:                result.Company.CompanyName,
			UnifiedSocialCreditCode:    result.Company.UnifiedSocialCreditCode,
			BusinessLicenseEndpoint:    result.Company.BusinessLicenseEndpoint,
			BusinessLicenseBucket:      result.Company.BusinessLicenseBucket,
			BusinessLicensePath:        result.Company.BusinessLicensePath,
			BusinessLicenseFileName:    result.Company.BusinessLicenseFileName,
			BusinessLicenseFileSize:    result.Company.BusinessLicenseFileSize,
			BusinessLicenseContentType: result.Company.BusinessLicenseContentType,
			RegisteredAddress:          result.Company.RegisteredAddress,
			Introduction:               result.Company.Introduction,
			StorePlatformID:            result.Company.StorePlatformID,
			StorePlatformName:          result.Company.StorePlatformName,
			StoreURL:                   result.Company.StoreURL,
			Offices:                    result.Company.Offices,
			Factories:                  result.Company.Factories,
			Contacts:                   toAPIContacts(result.Company.Contacts),
			VerifiedAt:                 result.Company.VerifiedAt, ReviewStatus: result.Company.ReviewStatus,
			ReviewReason: result.Company.ReviewReason, ReviewedAt: result.Company.ReviewedAt,
		}
	}
	return info
}

func toContactInputs(items []v1.CompanyContactInfo) []verificationModel.ContactInput {
	result := make([]verificationModel.ContactInput, 0, len(items))
	for _, item := range items {
		result = append(result, verificationModel.ContactInput{Name: item.Name, Position: item.Position, Phone: item.Phone})
	}
	return result
}

func toAPIContacts(items []verificationModel.ContactInfo) []v1.CompanyContactInfo {
	result := make([]v1.CompanyContactInfo, 0, len(items))
	for _, item := range items {
		result = append(result, v1.CompanyContactInfo{Name: item.Name, Position: item.Position, Phone: item.Phone})
	}
	return result
}
