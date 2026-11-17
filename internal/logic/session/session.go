package session

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	authmodel "github.com/suxiangdong/indu-scope-backend/internal/model/auth"
	model "github.com/suxiangdong/indu-scope-backend/internal/model/bizctx"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

type sSession struct{}

var _ service.ISession = (*sSession)(nil)

// New creates the session service. The actual session is request-scoped and
// comes from bizctx.Context, so this service does not own a connection or pool.
func New() service.ISession {
	return &sSession{}
}

func init() {
	service.RegisterSession(New())
}

// SetUser stores the authenticated user in the current request session.
func (s *sSession) SetUser(ctx context.Context, user authmodel.User) error {
	session, err := currentSession(ctx)
	if err != nil {
		return err
	}
	return session.Set(consts.SessionUserKey, user)
}

// GetUser retrieves the authenticated user from the current request session.
func (s *sSession) GetUser(ctx context.Context) (authmodel.User, error) {
	user := s.getUser(ctx)
	if user == nil {
		return authmodel.User{}, consts.ErrTokenInvalid
	}
	return authmodel.User{
		ID:          user.ID,
		Phone:       user.Phone,
		CompanyName: user.CompanyName,
		Role:        user.Role,
		UserType:    user.UserType,
		Status:      user.Status,
	}, nil
}

// RemoveUser removes the authenticated user from the current request session.
func (s *sSession) RemoveUser(ctx context.Context) error {
	session, err := currentSession(ctx)
	if err != nil {
		return err
	}
	return session.RemoveAll()
}

// RemoveAdminUser removes the authenticated administrator from the current session.
func (s *sSession) RemoveAdminUser(ctx context.Context) error {
	return s.RemoveUser(ctx)
}

// SetAdminUser stores the authenticated administrator in the current session.
func (s *sSession) SetAdminUser(ctx context.Context, user authmodel.AdminUser) error {
	session, err := currentSession(ctx)
	if err != nil {
		return err
	}
	return session.Set(consts.SessionUserKey, user)
}

// GetAdminUser retrieves the authenticated administrator from the current session.
func (s *sSession) GetAdminUser(ctx context.Context) (authmodel.AdminUser, error) {
	user := s.getUser(ctx)
	if user == nil {
		return authmodel.AdminUser{}, consts.ErrTokenInvalid
	}
	return authmodel.AdminUser{
		ID:           user.ID,
		Username:     user.Username,
		Name:         user.Name,
		Phone:        user.Phone,
		Status:       user.Status,
		IsSuperAdmin: user.IsSuperAdmin,
	}, nil
}

func (s *sSession) getUser(ctx context.Context) *model.ContextUser {
	customCtx := service.Bizctx().Get(ctx)
	if customCtx == nil || customCtx.Session == nil {
		return nil
	}
	v := customCtx.Session.MustGet(consts.SessionUserKey)
	if v == nil || v.IsNil() {
		return nil
	}
	var user *model.ContextUser
	if err := v.Struct(&user); err != nil || user == nil || user.ID <= 0 {
		return nil
	}
	return user
}

func currentSession(ctx context.Context) (*ghttp.Session, error) {
	value := service.Bizctx().Get(ctx)
	if value == nil || value.Session == nil {
		return nil, gerror.New("request session is unavailable")
	}
	return value.Session, nil
}
