package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"

	"github.com/suxiangdong/indu-scope-backend/api/auth/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	authmodel "github.com/suxiangdong/indu-scope-backend/internal/model/auth"
)

func toAPIData(result authmodel.AuthResult) v1.AuthData {
	return v1.AuthData{
		AccessToken: result.AccessToken,
		TokenType:   "Bearer",
		User: v1.UserInfo{
			ID:          result.User.ID,
			Phone:       result.User.Phone,
			CompanyName: result.User.CompanyName,
			Role:        result.User.Role,
			UserType:    result.User.UserType,
			Status:      result.User.Status,
		},
	}
}

func toAPIError(ctx context.Context, err error) error {
	// Keep upstream details in server logs only. WeChat error messages can
	// contain credential or implementation details that must not be exposed to
	// the mini-program.
	if isWechatAuthError(err) {
		glog.Errorf(ctx, "wechat auth error: %+v", err)
		return gerror.NewCode(gcode.CodeBusinessValidationFailed, "微信认证失败，请稍后重试")
	}
	switch {
	case errors.Is(err, consts.ErrPhoneAlreadyExists):
		return gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
	case errors.Is(err, consts.ErrVerificationTypeConflict):
		return gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
	case errors.Is(err, consts.ErrVerificationLocked):
		return gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
	case errors.Is(err, consts.ErrStorePlatformNotFound):
		return gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
	case errors.Is(err, consts.ErrInvalidCredentials),
		errors.Is(err, consts.ErrUserDisabled),
		errors.Is(err, consts.ErrTokenInvalid),
		errors.Is(err, consts.ErrWechatLoginInvalid),
		errors.Is(err, consts.ErrWechatPhoneConflict):
		return gerror.NewCode(gcode.CodeNotAuthorized, err.Error())
	}
	var parameterError *consts.InvalidParameterError
	if errors.As(err, &parameterError) {
		return gerror.NewCode(gcode.CodeInvalidParameter, parameterError.Error())
	}
	glog.Errorf(ctx, "auth internal error: %+v", err)
	return gerror.NewCode(gcode.CodeInternalError, "internal server error")
}

func isWechatAuthError(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.HasPrefix(message, "wechat login:") ||
		strings.HasPrefix(message, "wechat phone login:")
}

func authorizationToken(ctx context.Context) (string, error) {
	parts := strings.Fields(g.RequestFromCtx(ctx).Header.Get("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", gerror.NewCode(gcode.CodeNotAuthorized, "authorization header is required")
	}
	return parts[1], nil
}
