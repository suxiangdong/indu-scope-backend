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
	"github.com/gogf/gf/v2/frame/g"
	"golang.org/x/crypto/bcrypt"

	"github.com/suxiangdong/indu-scope-backend/internal/authutil"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	"github.com/suxiangdong/indu-scope-backend/internal/dao"
	authmodel "github.com/suxiangdong/indu-scope-backend/internal/model/auth"
	"github.com/suxiangdong/indu-scope-backend/internal/model/do"
	"github.com/suxiangdong/indu-scope-backend/internal/model/entity"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
	"github.com/suxiangdong/indu-scope-backend/internal/wechat"
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

func (s *sAuth) Register(ctx context.Context, phone, password string, userType uint8) (authmodel.AuthResult, error) {
	phone = strings.TrimSpace(phone)
	if err := validateCredentials(phone, password); err != nil {
		return authmodel.AuthResult{}, err
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
		Role:         0,
		UserType:     userType,
	}).InsertAndGetId()
	if err != nil {
		if isDuplicateKey(err) {
			return authmodel.AuthResult{}, consts.ErrPhoneAlreadyExists
		}
		return authmodel.AuthResult{}, gerror.Wrap(err, "create user")
	}
	result, err := authutil.IssueUserToken(ctx, authmodel.User{
		ID: id, Phone: phone, Role: 0, UserType: userType, Status: 0,
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

func (s *sAuth) WechatPhoneLogin(ctx context.Context, loginCode, phoneCode string) (authmodel.AuthResult, error) {
	loginCode, phoneCode = strings.TrimSpace(loginCode), strings.TrimSpace(phoneCode)
	if loginCode == "" || phoneCode == "" {
		return authmodel.AuthResult{}, &consts.InvalidParameterError{Message: "wechat login code and phone code are required"}
	}
	loginSession, err := wechat.ResolveLoginCode(ctx, loginCode)
	if err != nil {
		return authmodel.AuthResult{}, gerror.Wrap(err, "wechat login")
	}
	openID := strings.TrimSpace(loginSession.OpenID)
	if openID == "" {
		return authmodel.AuthResult{}, consts.ErrWechatLoginInvalid
	}
	phone, err := wechat.ResolvePhoneCode(ctx, phoneCode)
	if err != nil {
		return authmodel.AuthResult{}, gerror.Wrap(err, "wechat phone login")
	}
	phone = strings.TrimSpace(phone)
	if err := validatePhone(phone); err != nil {
		return authmodel.AuthResult{}, err
	}

	var byOpenID *wechatUserRecord
	err = dao.Users.Ctx(ctx).
		Fields("id, phone, password_hash, status, role, user_type, wechat_openid").
		Where("wechat_openid", openID).
		Scan(&byOpenID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return authmodel.AuthResult{}, gerror.Wrap(err, "query wechat user")
	}
	if byOpenID != nil && byOpenID.Phone != phone {
		return authmodel.AuthResult{}, consts.ErrWechatPhoneConflict
	}

	user := byOpenID
	if user == nil {
		var byPhone *wechatUserRecord
		err = dao.Users.Ctx(ctx).
			Fields("id, phone, password_hash, status, role, user_type, wechat_openid").
			Where(do.Users{Phone: phone}).
			Scan(&byPhone)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return authmodel.AuthResult{}, gerror.Wrap(err, "query phone user")
		}
		if byPhone != nil && byPhone.WechatOpenID != "" && byPhone.WechatOpenID != openID {
			return authmodel.AuthResult{}, consts.ErrWechatPhoneConflict
		}
		user = byPhone
	}

	if user == nil {
		id, err := dao.Users.Ctx(ctx).Data(g.Map{
			"phone":         phone,
			"password_hash": "",
			"wechat_openid": openID,
			"status":        0,
			"role":          0,
			"user_type":     1,
		}).InsertAndGetId()
		if err != nil {
			if isDuplicateKey(err) {
				return authmodel.AuthResult{}, consts.ErrWechatPhoneConflict
			}
			return authmodel.AuthResult{}, gerror.Wrap(err, "create wechat user")
		}
		user = &wechatUserRecord{Id: uint64(id), Phone: phone, Status: 0, Role: 0, UserType: 1, WechatOpenID: openID}
	} else if user.WechatOpenID == "" {
		if _, err := dao.Users.Ctx(ctx).Data(g.Map{"wechat_openid": openID}).Where(do.Users{Id: user.Id}).Update(); err != nil {
			if isDuplicateKey(err) {
				return authmodel.AuthResult{}, consts.ErrWechatPhoneConflict
			}
			return authmodel.AuthResult{}, gerror.Wrap(err, "bind wechat user")
		}
		user.WechatOpenID = openID
	}
	if user.Status == 2 {
		return authmodel.AuthResult{}, consts.ErrUserDisabled
	}

	result, err := authutil.IssueUserToken(ctx, authmodel.User{
		ID: int64(user.Id), Phone: user.Phone, Role: uint8(user.Role),
		UserType: uint8(user.UserType), Status: uint8(user.Status),
	})
	if err != nil {
		return authmodel.AuthResult{}, err
	}
	if err := s.setUserSession(ctx, result.AccessToken, result.User); err != nil {
		return authmodel.AuthResult{}, gerror.Wrap(err, "create user session")
	}
	return result, nil
}

type wechatUserRecord struct {
	Id           uint64 `orm:"id"`
	Phone        string `orm:"phone"`
	PasswordHash string `orm:"password_hash"`
	WechatOpenID string `orm:"wechat_openid"`
	Status       uint   `orm:"status"`
	Role         uint   `orm:"role"`
	UserType     uint   `orm:"user_type"`
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
	if err := validatePhone(phone); err != nil {
		return err
	}
	if len(password) < 8 || len(password) > 72 {
		return &consts.InvalidParameterError{Message: "password must contain 8 to 72 characters"}
	}
	return nil
}

func validatePhone(phone string) error {
	if len(phone) < 6 || len(phone) > 20 {
		return &consts.InvalidParameterError{Message: "phone must contain 6 to 20 digits"}
	}
	for _, r := range phone {
		if r < '0' || r > '9' {
			return &consts.InvalidParameterError{Message: "phone must contain digits only"}
		}
	}
	return nil
}

func isDuplicateKey(err error) bool {
	var mysqlErr *mysqlDriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
