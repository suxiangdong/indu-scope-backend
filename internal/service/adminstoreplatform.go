package service

import (
	"context"

	storePlatformModel "github.com/suxiangdong/indu-scope-backend/internal/model/storeplatform"
)

type IAdminStorePlatform interface {
	List(ctx context.Context) (storePlatformModel.ListResult, error)
	Create(ctx context.Context, input storePlatformModel.Input) (storePlatformModel.Item, error)
	Update(ctx context.Context, id uint64, input storePlatformModel.Input) (storePlatformModel.Item, error)
	Delete(ctx context.Context, id uint64) error
}

var localAdminStorePlatform IAdminStorePlatform

func AdminStorePlatform() IAdminStorePlatform {
	if localAdminStorePlatform == nil {
		panic("implement not found for interface IAdminStorePlatform, forgot register?")
	}
	return localAdminStorePlatform
}

func RegisterAdminStorePlatform(i IAdminStorePlatform) {
	localAdminStorePlatform = i
}
