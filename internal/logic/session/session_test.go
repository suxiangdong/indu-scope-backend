package session

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/os/gsession"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	_ "github.com/suxiangdong/indu-scope-backend/internal/logic/bizctx"
	authmodel "github.com/suxiangdong/indu-scope-backend/internal/model/auth"
	bizmodel "github.com/suxiangdong/indu-scope-backend/internal/model/bizctx"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func TestAdminUserSessionRoundTrip(t *testing.T) {
	session := gsession.New(0, gsession.NewStorageMemory()).New(context.Background(), "admin-token")
	ctx := service.Bizctx().Set(context.Background(), &bizmodel.Context{Session: session})
	serviceImpl := &sSession{}
	want := authmodel.AdminUser{
		ID:       42,
		Username: "cffyadmin",
		Name:     "管理员",
		Phone:    "13800138000",
		Status:   1,
	}

	if err := serviceImpl.SetAdminUser(ctx, want); err != nil {
		t.Fatalf("SetAdminUser() error = %v", err)
	}
	got, err := serviceImpl.GetAdminUser(ctx)
	if err != nil {
		t.Fatalf("GetAdminUser() error = %v", err)
	}
	if got != want {
		t.Fatalf("GetAdminUser() = %+v, want %+v", got, want)
	}
	if value := session.MustGet(consts.SessionUserKey); value == nil || value.IsNil() {
		t.Fatal("session user value is empty")
	}
}
