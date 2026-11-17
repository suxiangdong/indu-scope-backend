package userprofile

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	"github.com/suxiangdong/indu-scope-backend/internal/dao"
	"github.com/suxiangdong/indu-scope-backend/internal/model/entity"
	userprofileModel "github.com/suxiangdong/indu-scope-backend/internal/model/userprofile"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

type sUserProfile struct{}

var _ service.IUserProfile = (*sUserProfile)(nil)

func New() service.IUserProfile { return &sUserProfile{} }

func init() { service.RegisterUserProfile(New()) }

func (s *sUserProfile) Profile(ctx context.Context, userID uint64) (userprofileModel.Profile, error) {
	if userID == 0 {
		return userprofileModel.Profile{}, &consts.InvalidParameterError{Message: "id is required"}
	}

	var user entity.Users
	if err := dao.Users.Ctx(ctx).Cache(gdb.CacheOption{}).
		Fields("id", "company_name", "status", "role").Where("id", userID).Scan(&user); err != nil {
		return userprofileModel.Profile{}, gerror.Wrap(err, "query user profile user")
	}
	if user.Id == 0 || user.Status != 1 || (user.Role != 1 && user.Role != 2) {
		return userprofileModel.Profile{}, consts.ErrCustomerUserNotFound
	}

	var verification entity.CompanyVerifications
	if err := dao.CompanyVerifications.Ctx(ctx).Cache(gdb.CacheOption{}).
		Fields("user_id", "company_name", "registered_address").Where("user_id", userID).Scan(&verification); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return userprofileModel.Profile{}, gerror.Wrap(err, "query user profile verification")
	}

	var companyProfile entity.CompanyProfiles
	if err := dao.CompanyProfiles.Ctx(ctx).Cache(gdb.CacheOption{}).
		Fields("user_id", "introduction_html", "offices", "factories").Where("user_id", userID).Scan(&companyProfile); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return userprofileModel.Profile{}, gerror.Wrap(err, "query user profile company")
	}

	var contacts []*entity.CompanyContacts
	if err := dao.CompanyContacts.Ctx(ctx).Cache(gdb.CacheOption{}).
		Where("user_id", userID).OrderAsc("id").Scan(&contacts); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return userprofileModel.Profile{}, gerror.Wrap(err, "query user profile contacts")
	}

	companyName := user.CompanyName
	if companyName == "" {
		companyName = verification.CompanyName
	}
	profile := userprofileModel.Profile{
		ID:                user.Id,
		CompanyName:       companyName,
		Role:              uint8(user.Role),
		Introduction:      companyProfile.IntroductionHtml,
		RegisteredAddress: verification.RegisteredAddress,
		Offices:           decodeLocations(companyProfile.Offices),
		Factories:         decodeLocations(companyProfile.Factories),
		Contacts:          make([]userprofileModel.Contact, 0, len(contacts)),
	}
	for _, contact := range contacts {
		if contact == nil {
			continue
		}
		profile.Contacts = append(profile.Contacts, userprofileModel.Contact{Name: contact.Name, Position: contact.Position, Phone: contact.Phone})
	}
	return profile, nil
}

func decodeLocations(raw string) []string {
	locations := make([]string, 0)
	if raw == "" {
		return locations
	}
	if err := json.Unmarshal([]byte(raw), &locations); err != nil {
		return []string{}
	}
	return locations
}
