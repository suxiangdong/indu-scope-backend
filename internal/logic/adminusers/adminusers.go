package adminusers

import (
	"context"
	"errors"
	"strings"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/errors/gerror"
	"golang.org/x/crypto/bcrypt"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	"github.com/suxiangdong/indu-scope-backend/internal/dao"
	adminusermodel "github.com/suxiangdong/indu-scope-backend/internal/model/adminuser"
	"github.com/suxiangdong/indu-scope-backend/internal/model/do"
	"github.com/suxiangdong/indu-scope-backend/internal/model/entity"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

// sAdminUsers implements administrator account management for the console.
type sAdminUsers struct{}

var _ service.IAdminUsers = (*sAdminUsers)(nil)

// New creates the administrator management service.
func New() service.IAdminUsers {
	return &sAdminUsers{}
}

func init() {
	service.RegisterAdminUsers(New())
}

// List returns all administrator accounts.
func (s *sAdminUsers) List(ctx context.Context) (adminusermodel.AdminUserListResult, error) {
	cols := dao.AdminUsers.Columns()
	query := dao.AdminUsers.Ctx(ctx).Fields(
		cols.Id, cols.Username, cols.Name, cols.Phone, cols.Status,
		cols.LastLoginAt, cols.CreatedAt, cols.UpdatedAt,
	).OrderDesc(cols.CreatedAt)
	var records []*entity.AdminUsers
	if err := query.Scan(&records); err != nil {
		return adminusermodel.AdminUserListResult{}, gerror.Wrap(err, "list admin users")
	}

	items := make([]adminusermodel.AdminUserItem, 0, len(records))
	for _, record := range records {
		if record != nil {
			items = append(items, toAdminUserItem(record))
		}
	}
	return adminusermodel.AdminUserListResult{Items: items}, nil
}

// Create creates an administrator account.
func (s *sAdminUsers) Create(ctx context.Context, username, name, phone, password string, status uint) (adminusermodel.AdminUserItem, error) {
	username, name, phone = strings.TrimSpace(username), strings.TrimSpace(name), strings.TrimSpace(phone)
	if err := validateAdminUser(username, name, phone, password, status, true); err != nil {
		return adminusermodel.AdminUserItem{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return adminusermodel.AdminUserItem{}, gerror.Wrap(err, "hash admin password")
	}
	id, err := dao.AdminUsers.Ctx(ctx).Data(do.AdminUsers{
		Username: username, Name: name, Phone: phone,
		PasswordHash: string(hash), Status: status,
	}).InsertAndGetId()
	if err != nil {
		if isDuplicateKey(err) {
			return adminusermodel.AdminUserItem{}, consts.ErrAdminUserAlreadyExists
		}
		return adminusermodel.AdminUserItem{}, gerror.Wrap(err, "create admin user")
	}

	var record *entity.AdminUsers
	if err := dao.AdminUsers.Ctx(ctx).
		Fields(dao.AdminUsers.Columns().Id, dao.AdminUsers.Columns().Username, dao.AdminUsers.Columns().Name,
			dao.AdminUsers.Columns().Phone, dao.AdminUsers.Columns().Status, dao.AdminUsers.Columns().CreatedAt,
			dao.AdminUsers.Columns().UpdatedAt).
		Where(do.AdminUsers{Id: id}).Scan(&record); err != nil || record == nil {
		if err != nil {
			return adminusermodel.AdminUserItem{}, gerror.Wrap(err, "query created admin user")
		}
		return adminusermodel.AdminUserItem{}, consts.ErrAdminUserNotFound
	}
	return toAdminUserItem(record), nil
}

// Update updates an administrator account. An empty password keeps the old password.
func (s *sAdminUsers) Update(ctx context.Context, id uint64, username, name, phone, password string, status uint) (adminusermodel.AdminUserItem, error) {
	username, name, phone = strings.TrimSpace(username), strings.TrimSpace(name), strings.TrimSpace(phone)
	if id == 0 {
		return adminusermodel.AdminUserItem{}, &consts.InvalidParameterError{Message: "id is required"}
	}
	if err := validateAdminUser(username, name, phone, password, status, false); err != nil {
		return adminusermodel.AdminUserItem{}, err
	}
	if count, err := dao.AdminUsers.Ctx(ctx).Where(do.AdminUsers{Id: id}).Count(); err != nil {
		return adminusermodel.AdminUserItem{}, gerror.Wrap(err, "query admin user")
	} else if count == 0 {
		return adminusermodel.AdminUserItem{}, consts.ErrAdminUserNotFound
	}

	data := do.AdminUsers{Username: username, Name: name, Phone: phone, Status: status}
	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return adminusermodel.AdminUserItem{}, gerror.Wrap(err, "hash admin password")
		}
		data.PasswordHash = string(hash)
	}
	if _, err := dao.AdminUsers.Ctx(ctx).Where(do.AdminUsers{Id: id}).Data(data).Update(); err != nil {
		if isDuplicateKey(err) {
			return adminusermodel.AdminUserItem{}, consts.ErrAdminUserAlreadyExists
		}
		return adminusermodel.AdminUserItem{}, gerror.Wrap(err, "update admin user")
	}
	var record *entity.AdminUsers
	if err := dao.AdminUsers.Ctx(ctx).
		Fields(dao.AdminUsers.Columns().Id, dao.AdminUsers.Columns().Username, dao.AdminUsers.Columns().Name,
			dao.AdminUsers.Columns().Phone, dao.AdminUsers.Columns().Status, dao.AdminUsers.Columns().LastLoginAt,
			dao.AdminUsers.Columns().CreatedAt, dao.AdminUsers.Columns().UpdatedAt).
		Where(do.AdminUsers{Id: id}).Scan(&record); err != nil {
		return adminusermodel.AdminUserItem{}, gerror.Wrap(err, "query updated admin user")
	}
	if record == nil {
		return adminusermodel.AdminUserItem{}, consts.ErrAdminUserNotFound
	}
	return toAdminUserItem(record), nil
}

// Delete permanently deletes an administrator account.
func (s *sAdminUsers) Delete(ctx context.Context, id uint64) error {
	if id == 0 {
		return &consts.InvalidParameterError{Message: "id is required"}
	}
	if value := service.Bizctx().Get(ctx); value != nil && value.AdminUser != nil && value.AdminUser.ID == int64(id) {
		return consts.ErrAdminUserSelfDelete
	}
	result, err := dao.AdminUsers.Ctx(ctx).Where(do.AdminUsers{Id: id}).Delete()
	if err != nil {
		return gerror.Wrap(err, "delete admin user")
	}
	if affected, err := result.RowsAffected(); err != nil {
		return gerror.Wrap(err, "check deleted admin user")
	} else if affected == 0 {
		return consts.ErrAdminUserNotFound
	}
	return nil
}

func validateAdminUser(username, name, phone, password string, status uint, passwordRequired bool) error {
	if username == "" || len(username) > 64 {
		return &consts.InvalidParameterError{Message: "username must contain 1 to 64 characters"}
	}
	if name == "" || len(name) > 100 {
		return &consts.InvalidParameterError{Message: "name must contain 1 to 100 characters"}
	}
	if len(phone) < 6 || len(phone) > 20 {
		return &consts.InvalidParameterError{Message: "phone must contain 6 to 20 digits"}
	}
	for _, r := range phone {
		if r < '0' || r > '9' {
			return &consts.InvalidParameterError{Message: "phone must contain digits only"}
		}
	}
	if passwordRequired && (len(password) < 6 || len(password) > 72) {
		return &consts.InvalidParameterError{Message: "password must contain 6 to 72 characters"}
	}
	if !passwordRequired && password != "" && (len(password) < 6 || len(password) > 72) {
		return &consts.InvalidParameterError{Message: "password must contain 6 to 72 characters"}
	}
	if status > 1 {
		return &consts.InvalidParameterError{Message: "status must be 0 or 1"}
	}
	return nil
}

func toAdminUserItem(record *entity.AdminUsers) adminusermodel.AdminUserItem {
	var lastLoginAt *time.Time
	if !record.LastLoginAt.IsZero() {
		value := record.LastLoginAt
		lastLoginAt = &value
	}
	return adminusermodel.AdminUserItem{
		ID: int64(record.Id), Username: record.Username, Name: record.Name, Phone: record.Phone,
		Status: uint8(record.Status), LastLoginAt: lastLoginAt,
		CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
	}
}

func isDuplicateKey(err error) bool {
	var mysqlErr *mysqlDriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
