package admincustomerusers

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	"github.com/suxiangdong/indu-scope-backend/internal/dao"
	adminCustomerUserModel "github.com/suxiangdong/indu-scope-backend/internal/model/admincustomeruser"
	"github.com/suxiangdong/indu-scope-backend/internal/model/do"
	"github.com/suxiangdong/indu-scope-backend/internal/model/entity"
	"github.com/suxiangdong/indu-scope-backend/internal/model/verification"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

type sAdminCustomerUsers struct{}

var _ service.IAdminCustomerUsers = (*sAdminCustomerUsers)(nil)

func New() service.IAdminCustomerUsers { return &sAdminCustomerUsers{} }

func init() { service.RegisterAdminCustomerUsers(New()) }

func (s *sAdminCustomerUsers) List(ctx context.Context, filter adminCustomerUserModel.ListFilter) (adminCustomerUserModel.ListResult, error) {
	filter.Phone = strings.TrimSpace(filter.Phone)
	if err := validateFilter(filter); err != nil {
		return adminCustomerUserModel.ListResult{}, err
	}
	cols := dao.Users.Columns()
	query := dao.Users.Ctx(ctx).Fields(cols.Id, cols.Phone, cols.Status, cols.Role, cols.UserType, cols.WechatOpenID, cols.CreatedAt, cols.UpdatedAt).OrderDesc(cols.CreatedAt)
	if filter.Phone != "" {
		query = query.WhereLike(cols.Phone, "%"+filter.Phone+"%")
	}
	if filter.Role != nil {
		query = query.Where(do.Users{Role: *filter.Role})
	}
	if filter.UserType != nil {
		query = query.Where(do.Users{UserType: *filter.UserType})
	}
	if filter.Status != nil {
		query = query.Where(do.Users{Status: *filter.Status})
	}
	var records []*entity.Users
	if err := query.Scan(&records); err != nil {
		return adminCustomerUserModel.ListResult{}, gerror.Wrap(err, "list customer users")
	}
	verificationMap, err := loadVerificationMap(ctx, records)
	if err != nil {
		return adminCustomerUserModel.ListResult{}, err
	}
	items := make([]adminCustomerUserModel.UserItem, 0, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		item := toItem(record)
		if value, ok := verificationMap[record.Id]; ok {
			item.VerificationType = value.Type
			item.VerificationStatus = value.Status
		} else if record.Status == 1 {
			item.VerificationStatus = "verified"
		} else {
			item.VerificationStatus = "unsubmitted"
		}
		items = append(items, item)
	}
	return adminCustomerUserModel.ListResult{Items: items}, nil
}

func (s *sAdminCustomerUsers) UpdateStatus(ctx context.Context, id uint64, status uint8) (adminCustomerUserModel.UserItem, error) {
	if id == 0 {
		return adminCustomerUserModel.UserItem{}, &consts.InvalidParameterError{Message: "id is required"}
	}
	if status > 2 {
		return adminCustomerUserModel.UserItem{}, &consts.InvalidParameterError{Message: "status must be 0, 1 or 2"}
	}
	if count, err := dao.Users.Ctx(ctx).Where(do.Users{Id: id}).Count(); err != nil {
		return adminCustomerUserModel.UserItem{}, gerror.Wrap(err, "query customer user")
	} else if count == 0 {
		return adminCustomerUserModel.UserItem{}, consts.ErrCustomerUserNotFound
	}
	if _, err := dao.Users.Ctx(ctx).Where(do.Users{Id: id}).Data(g.Map{"status": status}).Update(); err != nil {
		return adminCustomerUserModel.UserItem{}, gerror.Wrap(err, "update customer user status")
	}
	result, err := s.List(ctx, adminCustomerUserModel.ListFilter{})
	if err != nil {
		return adminCustomerUserModel.UserItem{}, err
	}
	for _, item := range result.Items {
		if uint64(item.ID) == id {
			return item, nil
		}
	}
	return adminCustomerUserModel.UserItem{}, consts.ErrCustomerUserNotFound
}

type verificationSummary struct {
	Type   string
	Status string
}

func loadVerificationMap(ctx context.Context, records []*entity.Users) (map[uint64]verificationSummary, error) {
	result := make(map[uint64]verificationSummary)
	ids := make([]uint64, 0, len(records))
	for _, record := range records {
		if record != nil {
			ids = append(ids, record.Id)
		}
	}
	if len(ids) == 0 {
		return result, nil
	}
	var personalRecords []*entity.PersonalVerifications
	personalCols := dao.PersonalVerifications.Columns()
	if err := dao.PersonalVerifications.Ctx(ctx).Fields(personalCols.UserId, personalCols.ReviewStatus, personalCols.VerifiedAt).
		WhereIn(personalCols.UserId, ids).Scan(&personalRecords); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, gerror.Wrap(err, "query personal verification status")
	}
	for _, record := range personalRecords {
		if record != nil {
			result[record.UserId] = verificationSummary{Type: "personal", Status: verificationStatus(uint8(record.ReviewStatus), !record.VerifiedAt.IsZero())}
		}
	}
	var companyRecords []*entity.CompanyVerifications
	companyCols := dao.CompanyVerifications.Columns()
	if err := dao.CompanyVerifications.Ctx(ctx).Fields(companyCols.UserId, companyCols.ReviewStatus, companyCols.VerifiedAt).
		WhereIn(companyCols.UserId, ids).Scan(&companyRecords); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, gerror.Wrap(err, "query company verification status")
	}
	for _, record := range companyRecords {
		if record != nil {
			result[record.UserId] = verificationSummary{Type: "company", Status: verificationStatus(uint8(record.ReviewStatus), !record.VerifiedAt.IsZero())}
		}
	}
	return result, nil
}

func verificationStatus(reviewStatus uint8, verifiedAt bool) string {
	if verifiedAt || reviewStatus == verification.ReviewStatusApproved {
		return "verified"
	}
	if reviewStatus == verification.ReviewStatusRejected {
		return "rejected"
	}
	return "pending"
}

func validateFilter(filter adminCustomerUserModel.ListFilter) error {
	if filter.Role != nil && *filter.Role > 2 {
		return &consts.InvalidParameterError{Message: "role must be 0, 1 or 2"}
	}
	if filter.UserType != nil && (*filter.UserType < 1 || *filter.UserType > 2) {
		return &consts.InvalidParameterError{Message: "user_type must be 1 or 2"}
	}
	if filter.Status != nil && *filter.Status > 2 {
		return &consts.InvalidParameterError{Message: "status must be 0, 1 or 2"}
	}
	return nil
}

func toItem(record *entity.Users) adminCustomerUserModel.UserItem {
	return adminCustomerUserModel.UserItem{
		ID: int64(record.Id), Phone: record.Phone, Role: uint8(record.Role), UserType: uint8(record.UserType), Status: uint8(record.Status),
		WechatBound: strings.TrimSpace(record.WechatOpenID) != "", CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
	}
}
