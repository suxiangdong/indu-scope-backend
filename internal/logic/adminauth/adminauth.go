package adminauth

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"golang.org/x/crypto/bcrypt"

	"github.com/suxiangdong/indu-scope-backend/internal/authutil"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	"github.com/suxiangdong/indu-scope-backend/internal/dao"
	authmodel "github.com/suxiangdong/indu-scope-backend/internal/model/auth"
	"github.com/suxiangdong/indu-scope-backend/internal/model/do"
	"github.com/suxiangdong/indu-scope-backend/internal/model/entity"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

// sAdminAuth implements the generated service.IAdminAuth interface for
// management-console users.
type sAdminAuth struct{}

var _ service.IAdminAuth = (*sAdminAuth)(nil)

// New creates the administrator authentication service.
func New() service.IAdminAuth {
	return &sAdminAuth{}
}

func init() {
	service.RegisterAdminAuth(New())
}

func (s *sAdminAuth) Login(ctx context.Context, identifier, password string) (authmodel.AdminAuthResult, error) {
	identifier = strings.TrimSpace(identifier)
	if err := validateCredentials(identifier, password); err != nil {
		return authmodel.AdminAuthResult{}, err
	}

	var (
		record *entity.AdminUsers
		cols   = dao.AdminUsers.Columns()
	)
	err := dao.AdminUsers.Ctx(ctx).
		Fields(cols.Id, cols.Username, cols.Name, cols.Phone, cols.PasswordHash, cols.Status, cols.IsSuperAdmin).
		Where(do.AdminUsers{Username: identifier}).
		WhereOr(do.AdminUsers{Phone: identifier}).
		Scan(&record)
	if errors.Is(err, sql.ErrNoRows) || record == nil {
		return authmodel.AdminAuthResult{}, consts.ErrAdminInvalidCredentials
	}
	if err != nil {
		return authmodel.AdminAuthResult{}, gerror.Wrap(err, "query admin user")
	}
	if bcrypt.CompareHashAndPassword([]byte(record.PasswordHash), []byte(password)) != nil {
		return authmodel.AdminAuthResult{}, consts.ErrAdminInvalidCredentials
	}
	if record.Status == 0 {
		return authmodel.AdminAuthResult{}, consts.ErrAdminDisabled
	}
	if _, err := dao.AdminUsers.Ctx(ctx).
		Where(do.AdminUsers{Id: record.Id}).
		Data(do.AdminUsers{LastLoginAt: gdb.Raw("CURRENT_TIMESTAMP(3)")}).
		Update(); err != nil {
		return authmodel.AdminAuthResult{}, gerror.Wrap(err, "update admin last login time")
	}
	result, err := authutil.IssueAdminToken(ctx, authmodel.AdminUser{
		ID: int64(record.Id), Username: record.Username, Name: record.Name,
		Phone: record.Phone, Status: uint8(record.Status), IsSuperAdmin: record.IsSuperAdmin == 1,
	})
	if err != nil {
		return authmodel.AdminAuthResult{}, err
	}
	if err := s.setSession(ctx, result.AccessToken, result.User); err != nil {
		return authmodel.AdminAuthResult{}, gerror.Wrap(err, "create admin session")
	}
	return result, nil
}

// Me returns the active administrator represented by an admin access token.
func (s *sAdminAuth) Me(ctx context.Context, accessToken string) (authmodel.AdminUser, error) {
	id, err := s.loadSession(ctx, accessToken)
	if err != nil {
		return authmodel.AdminUser{}, err
	}
	return s.loadByID(ctx, id)
}

// Logout removes the Redis session represented by an admin access token.
func (s *sAdminAuth) Logout(ctx context.Context, accessToken string) error {
	if _, err := s.loadSession(ctx, accessToken); err != nil {
		return err
	}
	if err := service.Session().RemoveAdminUser(ctx); err != nil {
		return gerror.Wrap(err, "remove admin session")
	}
	return nil
}

func (s *sAdminAuth) setSession(ctx context.Context, accessToken string, user authmodel.AdminUser) error {
	if err := authutil.BindSession(ctx, accessToken); err != nil {
		return err
	}
	if err := service.Session().SetAdminUser(ctx, user); err != nil {
		return gerror.Wrap(err, "store admin session")
	}
	return nil
}

func (s *sAdminAuth) loadSession(ctx context.Context, accessToken string) (int64, error) {
	claims, err := authutil.ParseToken(ctx, accessToken, consts.AdminTokenIssuer)
	if err != nil {
		return 0, consts.ErrAdminTokenInvalid
	}
	claimsID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || claimsID <= 0 {
		return 0, consts.ErrAdminTokenInvalid
	}
	storedUser, err := service.Session().GetAdminUser(ctx)
	if err != nil {
		if errors.Is(err, consts.ErrTokenInvalid) {
			return 0, consts.ErrAdminTokenInvalid
		}
		return 0, gerror.Wrap(err, "load admin session")
	}
	if storedUser.ID != claimsID {
		return 0, consts.ErrAdminTokenInvalid
	}
	return storedUser.ID, nil
}

func (s *sAdminAuth) loadByID(ctx context.Context, id int64) (authmodel.AdminUser, error) {
	var (
		record *entity.AdminUsers
		cols   = dao.AdminUsers.Columns()
	)
	err := dao.AdminUsers.Ctx(ctx).
		Fields(cols.Id, cols.Username, cols.Name, cols.Phone, cols.Status, cols.IsSuperAdmin).
		Where(do.AdminUsers{Id: id}).
		Scan(&record)
	if errors.Is(err, sql.ErrNoRows) || record == nil {
		return authmodel.AdminUser{}, consts.ErrAdminTokenInvalid
	}
	if err != nil {
		return authmodel.AdminUser{}, gerror.Wrap(err, "query current admin user")
	}
	if record.Status == 0 {
		return authmodel.AdminUser{}, consts.ErrAdminDisabled
	}
	return authmodel.AdminUser{
		ID: int64(record.Id), Username: record.Username, Name: record.Name,
		Phone: record.Phone, Status: uint8(record.Status), IsSuperAdmin: record.IsSuperAdmin == 1,
	}, nil
}

func validateCredentials(identifier, password string) error {
	if len(identifier) < 1 || len(identifier) > 64 {
		return &consts.InvalidParameterError{Message: "username must contain 1 to 64 characters"}
	}
	if len(password) < 6 || len(password) > 72 {
		return &consts.InvalidParameterError{Message: "password must contain 6 to 72 characters"}
	}
	return nil
}
