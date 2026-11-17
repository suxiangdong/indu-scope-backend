package service

import (
	"context"

	analyticsModel "github.com/suxiangdong/indu-scope-backend/internal/model/analytics"
)

type IAnalytics interface {
	RecordClick(ctx context.Context, userID uint64, contentType analyticsModel.ContentType, contentID uint64) error
	Overview(ctx context.Context, days uint) (analyticsModel.Overview, error)
}

var localAnalytics IAnalytics

func Analytics() IAnalytics {
	if localAnalytics == nil {
		panic("implement not found for interface IAnalytics, forgot register?")
	}
	return localAnalytics
}

func RegisterAnalytics(i IAnalytics) {
	localAnalytics = i
}
