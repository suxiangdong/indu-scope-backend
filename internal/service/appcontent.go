package service

import (
	"context"

	appcontentModel "github.com/suxiangdong/indu-scope-backend/internal/model/appcontent"
)

type IAppContent interface {
	List(ctx context.Context) (appcontentModel.ListResult, error)
	Get(ctx context.Context, pageKey string) (appcontentModel.PageContent, error)
	Update(ctx context.Context, input appcontentModel.UpdateInput) (appcontentModel.PageContent, error)
}

var localAppContent IAppContent

func AppContent() IAppContent {
	if localAppContent == nil {
		panic("implement not found for interface IAppContent, forgot register?")
	}
	return localAppContent
}

func RegisterAppContent(i IAppContent) { localAppContent = i }
