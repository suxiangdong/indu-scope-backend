package adminverification

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"

	"github.com/suxiangdong/indu-scope-backend/api/adminverification/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	adminverificationmodel "github.com/suxiangdong/indu-scope-backend/internal/model/adminverification"
)

func toAPIListItem(item adminverificationmodel.ListItem) v1.VerificationListItem {
	return v1.VerificationListItem{ID: item.ID, UserID: item.UserID, Type: item.Type, Phone: item.Phone,
		UserType: item.UserType, DisplayName: item.DisplayName, ReviewStatus: item.ReviewStatus,
		ReviewReason: item.ReviewReason, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt, ReviewedAt: item.ReviewedAt}
}

func toAPIDetail(detail adminverificationmodel.Detail) v1.DetailInfo {
	result := v1.DetailInfo{VerificationListItem: toAPIListItem(detail.ListItem), UserStatus: detail.UserStatus, Role: detail.Role}
	if detail.Personal != nil {
		result.Personal = &v1.PersonalDetail{RealName: detail.Personal.RealName, IDNumberMasked: detail.Personal.IDNumberMasked,
			StorePlatformID: detail.Personal.StorePlatformID, StorePlatformName: detail.Personal.StorePlatformName, StoreURL: detail.Personal.StoreURL,
			IDFrontEndpoint: detail.Personal.IDFrontEndpoint, IDFrontBucket: detail.Personal.IDFrontBucket, IDFrontPath: detail.Personal.IDFrontPath,
			IDBackEndpoint: detail.Personal.IDBackEndpoint, IDBackBucket: detail.Personal.IDBackBucket, IDBackPath: detail.Personal.IDBackPath,
			VerifiedAt: detail.Personal.VerifiedAt}
	}
	if detail.Company != nil {
		contacts := make([]v1.ContactInfo, 0, len(detail.Company.Contacts))
		for _, contact := range detail.Company.Contacts {
			contacts = append(contacts, v1.ContactInfo{Name: contact.Name, Position: contact.Position, Phone: contact.Phone})
		}
		result.Company = &v1.CompanyDetail{CompanyName: detail.Company.CompanyName, UnifiedSocialCreditCode: detail.Company.UnifiedSocialCreditCode,
			BusinessLicenseEndpoint: detail.Company.BusinessLicenseEndpoint, BusinessLicenseBucket: detail.Company.BusinessLicenseBucket,
			BusinessLicensePath: detail.Company.BusinessLicensePath, BusinessLicenseURL: detail.Company.BusinessLicenseURL,
			BusinessLicenseFileName: detail.Company.BusinessLicenseFileName, BusinessLicenseFileSize: detail.Company.BusinessLicenseFileSize,
			BusinessLicenseContentType: detail.Company.BusinessLicenseContentType, RegisteredAddress: detail.Company.RegisteredAddress,
			Introduction: detail.Company.Introduction, StorePlatformID: detail.Company.StorePlatformID,
			StorePlatformName: detail.Company.StorePlatformName, StoreURL: detail.Company.StoreURL,
			Offices: detail.Company.Offices, Factories: detail.Company.Factories,
			Contacts: contacts, VerifiedAt: detail.Company.VerifiedAt}
	}
	return result
}

func toAPIError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, consts.ErrVerificationNotFound), errors.Is(err, consts.ErrVerificationAlreadyReviewed),
		errors.Is(err, consts.ErrUserDisabled), errors.Is(err, consts.ErrVerificationTypeConflict):
		return gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
	}
	var parameterError *consts.InvalidParameterError
	if errors.As(err, &parameterError) {
		return gerror.NewCode(gcode.CodeInvalidParameter, parameterError.Error())
	}
	glog.Errorf(ctx, "admin verification internal error: %+v", err)
	return gerror.NewCode(gcode.CodeInternalError, "internal server error")
}
