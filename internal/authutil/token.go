package authutil

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gsession"
	"github.com/golang-jwt/jwt/v5"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	authmodel "github.com/suxiangdong/indu-scope-backend/internal/model/auth"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

// IssueUserToken creates a user JWT without an expiration time.
func IssueUserToken(ctx context.Context, user authmodel.User) (authmodel.AuthResult, error) {
	secret, err := jwtSecret(ctx)
	if err != nil {
		return authmodel.AuthResult{}, err
	}
	return issueUserToken(secret, user)
}

func issueUserToken(secret []byte, user authmodel.User) (authmodel.AuthResult, error) {
	claims := jwt.RegisteredClaims{
		Issuer:   consts.UserTokenIssuer,
		ID:       gsession.NewSessionId(),
		Subject:  strconv.FormatInt(user.ID, 10),
		IssuedAt: jwt.NewNumericDate(time.Now()),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		return authmodel.AuthResult{}, gerror.Wrap(err, "sign access token")
	}
	return authmodel.AuthResult{AccessToken: token, User: user}, nil
}

// IssueAdminToken creates an administrator JWT without an expiration time.
func IssueAdminToken(ctx context.Context, user authmodel.AdminUser) (authmodel.AdminAuthResult, error) {
	secret, err := jwtSecret(ctx)
	if err != nil {
		return authmodel.AdminAuthResult{}, err
	}
	claims := jwt.RegisteredClaims{
		Issuer:   consts.AdminTokenIssuer,
		ID:       gsession.NewSessionId(),
		Subject:  strconv.FormatInt(user.ID, 10),
		IssuedAt: jwt.NewNumericDate(time.Now()),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		return authmodel.AdminAuthResult{}, gerror.Wrap(err, "sign admin access token")
	}
	return authmodel.AdminAuthResult{AccessToken: token, User: user}, nil
}

// ParseToken validates a JWT signature and issuer. Session validation is
// deliberately handled by the caller because user and admin sessions differ.
func ParseToken(ctx context.Context, accessToken, issuer string) (jwt.RegisteredClaims, error) {
	var claims jwt.RegisteredClaims
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return claims, consts.ErrTokenInvalid
	}
	secret, err := jwtSecret(ctx)
	if err != nil {
		return claims, err
	}
	token, err := jwt.ParseWithClaims(accessToken, &claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, consts.ErrTokenInvalid
		}
		return secret, nil
	})
	if err != nil || !token.Valid || claims.Issuer != issuer || claims.ID == "" || claims.Subject == "" {
		return jwt.RegisteredClaims{}, consts.ErrTokenInvalid
	}
	return claims, nil
}

// BindSession binds the access token to the request-scoped GoFrame session.
func BindSession(ctx context.Context, accessToken string) error {
	value := service.Bizctx().Get(ctx)
	if value == nil || value.Session == nil {
		return gerror.New("request session is unavailable")
	}
	if err := value.Session.SetId(accessToken); err != nil {
		return gerror.Wrap(err, "bind access token to session")
	}
	return nil
}

func jwtSecret(ctx context.Context) ([]byte, error) {
	value, err := g.Cfg().Get(ctx, "jwt.secret")
	if err != nil {
		return nil, gerror.Wrap(err, "read jwt.secret from config")
	}
	if value == nil || strings.TrimSpace(value.String()) == "" {
		return nil, gerror.New("jwt.secret is required in config.yaml")
	}
	secret := strings.TrimSpace(value.String())
	if len(secret) < 32 {
		return nil, gerror.New("jwt.secret must contain at least 32 characters")
	}
	return []byte(secret), nil
}
