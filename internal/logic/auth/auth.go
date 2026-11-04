package auth

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
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

// sAuth implements the generated service.IAuth interface for frontend users.
type sAuth struct{}

var _ service.IAuth = (*sAuth)(nil)

// New creates the frontend user authentication service.
func New() service.IAuth {
	return &sAuth{}
}

func init() {
	service.RegisterAuth(New())
}

func (s *sAuth) Register(ctx context.Context, phone, password string, role, userType uint8) (authmodel.AuthResult, error) {
	phone = strings.TrimSpace(phone)
	if err := validateCredentials(phone, password); err != nil {
		return authmodel.AuthResult{}, err
	}
	if role != 1 && role != 2 {
		return authmodel.AuthResult{}, &consts.InvalidParameterError{Message: "role must be 1 or 2"}
	}
	if userType != 1 && userType != 2 {
		return authmodel.AuthResult{}, &consts.InvalidParameterError{Message: "user_type must be 1 or 2"}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return authmodel.AuthResult{}, gerror.Wrap(err, "hash password")
	}
	id, err := dao.Users.Ctx(ctx).Data(do.Users{
		Phone:        phone,
		PasswordHash: string(hash),
		Status:       0,
		Role:         role,
		UserType:     userType,
	}).InsertAndGetId()
	if err != nil {
		if isDuplicateKey(err) {
			return authmodel.AuthResult{}, consts.ErrPhoneAlreadyExists
		}
		return authmodel.AuthResult{}, gerror.Wrap(err, "create user")
	}
	result, err := authutil.IssueUserToken(ctx, authmodel.User{
		ID: id, Phone: phone, Role: role, UserType: userType, Status: 0,
	})
	if err != nil {
		return authmodel.AuthResult{}, err
	}
	if err := s.setUserSession(ctx, result.AccessToken, result.User); err != nil {
		return authmodel.AuthResult{}, gerror.Wrap(err, "create user session")
	}
	return result, nil
}

func (s *sAuth) Login(ctx context.Context, phone, password string) (authmodel.AuthResult, error) {
	phone = strings.TrimSpace(phone)
	if err := validateCredentials(phone, password); err != nil {
		return authmodel.AuthResult{}, consts.ErrInvalidCredentials
	}

	var (
		record *entity.Users
		cols   = dao.Users.Columns()
	)
	err := dao.Users.Ctx(ctx).
		Fields(cols.Id, cols.Phone, cols.PasswordHash, cols.Status, cols.Role, cols.UserType).
		Where(do.Users{Phone: phone}).
		Scan(&record)
	if errors.Is(err, sql.ErrNoRows) || record == nil {
		return authmodel.AuthResult{}, consts.ErrInvalidCredentials
	}
	if err != nil {
		return authmodel.AuthResult{}, gerror.Wrap(err, "query user")
	}
	if bcrypt.CompareHashAndPassword([]byte(record.PasswordHash), []byte(password)) != nil {
		return authmodel.AuthResult{}, consts.ErrInvalidCredentials
	}
	if record.Status == 2 {
		return authmodel.AuthResult{}, consts.ErrUserDisabled
	}
	result, err := authutil.IssueUserToken(ctx, authmodel.User{
		ID: int64(record.Id), Phone: record.Phone, Role: uint8(record.Role),
		UserType: uint8(record.UserType), Status: uint8(record.Status),
	})
	if err != nil {
		return authmodel.AuthResult{}, err
	}
	if err := s.setUserSession(ctx, result.AccessToken, result.User); err != nil {
		return authmodel.AuthResult{}, gerror.Wrap(err, "create user session")
	}
	return result, nil
}

// Me returns the active user represented by a user access token and its Redis session.
func (s *sAuth) Me(ctx context.Context, accessToken string) (authmodel.User, error) {
	id, err := s.loadUserSession(ctx, accessToken)
	if err != nil {
		return authmodel.User{}, err
	}
	return s.loadUserByID(ctx, id)
}

// Logout removes the Redis session represented by a user access token.
func (s *sAuth) Logout(ctx context.Context, accessToken string) error {
	if _, err := s.loadUserSession(ctx, accessToken); err != nil {
		return err
	}
	if err := service.Session().RemoveUser(ctx); err != nil {
		return gerror.Wrap(err, "remove user session")
	}
	return nil
}

func (s *sAuth) setUserSession(ctx context.Context, accessToken string, user authmodel.User) error {
	if err := authutil.BindSession(ctx, accessToken); err != nil {
		return err
	}
	if err := service.Session().SetUser(ctx, user); err != nil {
		return gerror.Wrap(err, "store user session")
	}
	return nil
}

func (s *sAuth) loadUserSession(ctx context.Context, accessToken string) (int64, error) {
	claims, err := authutil.ParseToken(ctx, accessToken, consts.UserTokenIssuer)
	if err != nil {
		return 0, consts.ErrTokenInvalid
	}
	claimsID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || claimsID <= 0 {
		return 0, consts.ErrTokenInvalid
	}
	storedUser, err := service.Session().GetUser(ctx)
	if err != nil {
		if errors.Is(err, consts.ErrTokenInvalid) {
			return 0, consts.ErrTokenInvalid
		}
		return 0, gerror.Wrap(err, "load user session")
	}
	if storedUser.ID != claimsID {
		return 0, consts.ErrTokenInvalid
	}
	return storedUser.ID, nil
}

func (s *sAuth) loadUserByID(ctx context.Context, id int64) (authmodel.User, error) {
	var (
		record *entity.Users
		cols   = dao.Users.Columns()
	)
	err := dao.Users.Ctx(ctx).
		Fields(cols.Id, cols.Phone, cols.Status, cols.Role, cols.UserType).
		Where(do.Users{Id: id}).
		Scan(&record)
	if errors.Is(err, sql.ErrNoRows) || record == nil {
		return authmodel.User{}, consts.ErrTokenInvalid
	}
	if err != nil {
		return authmodel.User{}, gerror.Wrap(err, "query current user")
	}
	if record.Status == 2 {
		return authmodel.User{}, consts.ErrUserDisabled
	}
	return authmodel.User{
		ID: int64(record.Id), Phone: record.Phone, Role: uint8(record.Role),
		UserType: uint8(record.UserType), Status: uint8(record.Status),
	}, nil
}

func validateCredentials(phone, password string) error {
	if len(phone) < 6 || len(phone) > 20 {
		return &consts.InvalidParameterError{Message: "phone must contain 6 to 20 digits"}
	}
	for _, r := range phone {
		if r < '0' || r > '9' {
			return &consts.InvalidParameterError{Message: "phone must contain digits only"}
		}
	}
	if len(password) < 8 || len(password) > 72 {
		return &consts.InvalidParameterError{Message: "password must contain 8 to 72 characters"}
	}
	return nil
}

func isDuplicateKey(err error) bool {
	var mysqlErr *mysqlDriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
