package service

import (
	"context"

	adminverificationmodel "github.com/suxiangdong/indu-scope-backend/internal/model/adminverification"
)

type IAdminVerification interface {
	List(ctx context.Context, filter adminverificationmodel.ListFilter) (adminverificationmodel.ListResult, error)
	Detail(ctx context.Context, verificationType string, id uint64) (adminverificationmodel.Detail, error)
	Review(ctx context.Context, adminID uint64, input adminverificationmodel.ReviewInput) (adminverificationmodel.Detail, error)
}

var localAdminVerification IAdminVerification

func AdminVerification() IAdminVerification {
	if localAdminVerification == nil {
		panic("implement not found for interface IAdminVerification, forgot register?")
	}
	return localAdminVerification
}

func RegisterAdminVerification(i IAdminVerification) {
	localAdminVerification = i
}
