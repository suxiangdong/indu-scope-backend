package authutil

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	authmodel "github.com/suxiangdong/indu-scope-backend/internal/model/auth"
)

func TestIssueUserToken(t *testing.T) {
	secret := []byte("01234567890123456789012345678901")
	result, err := issueUserToken(secret, authmodel.User{
		ID: 42, Phone: "13800138000", Role: 1, UserType: 1, Status: 0,
	})
	if err != nil {
		t.Fatalf("issueUserToken() error = %v", err)
	}
	if result.AccessToken == "" {
		t.Fatalf("issueUserToken() returned empty token: %+v", result)
	}

	token, err := jwt.ParseWithClaims(result.AccessToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		return secret, nil
	})
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	claims := token.Claims.(*jwt.RegisteredClaims)
	if !token.Valid || claims.Subject != "42" || claims.Issuer != consts.UserTokenIssuer || claims.ID == "" || claims.ExpiresAt != nil {
		t.Fatalf("unexpected claims: %+v", token.Claims)
	}
}
