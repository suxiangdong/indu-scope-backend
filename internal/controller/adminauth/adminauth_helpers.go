package adminauth

import (
	"context"
	"errors"
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"

	"github.com/suxiangdong/indu-scope-backend/api/adminauth/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	authmodel "github.com/suxiangdong/indu-scope-backend/internal/model/auth"
)

func toAuthData(result authmodel.AdminAuthResult) v1.AuthData {
	return v1.AuthData{
		AccessToken: result.AccessToken,
		TokenType:   "Bearer",
		User:        toUserInfo(result.User),
	}
}

func toUserInfo(user authmodel.AdminUser) v1.AdminUserInfo {
	return v1.AdminUserInfo{
		ID:           user.ID,
		Username:     user.Username,
		Name:         user.Name,
		Phone:        user.Phone,
		Status:       user.Status,
		IsSuperAdmin: user.IsSuperAdmin,
	}
}

func bearerToken(authorization string) (string, bool) {
	parts := strings.Fields(authorization)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

func toAPIError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, consts.ErrAdminInvalidCredentials),
		errors.Is(err, consts.ErrAdminDisabled),
		errors.Is(err, consts.ErrAdminTokenInvalid),
		errors.Is(err, consts.ErrTokenInvalid):
		return gerror.NewCode(gcode.CodeNotAuthorized, "invalid admin credentials or token")
	}
	var parameterError *consts.InvalidParameterError
	if errors.As(err, &parameterError) {
		return gerror.NewCode(gcode.CodeInvalidParameter, parameterError.Error())
	}
	glog.Errorf(ctx, "admin auth internal error: %+v", err)
	return gerror.NewCode(gcode.CodeInternalError, "internal server error")
}

func authorizationToken(ctx context.Context) (string, error) {
	authorization := g.RequestFromCtx(ctx).Header.Get("Authorization")
	token, ok := bearerToken(authorization)
	if !ok {
		return "", gerror.NewCode(gcode.CodeNotAuthorized, "authorization header is required")
	}
	return token, nil
}
