package service

import (
	"context"

	favoriteModel "github.com/suxiangdong/indu-scope-backend/internal/model/favorite"
)

type IFavorites interface {
	List(ctx context.Context, userID uint64, page, pageSize uint) (favoriteModel.ListResult, error)
	Toggle(ctx context.Context, userID, productID uint64) (favoriteModel.State, error)
	Status(ctx context.Context, userID, productID uint64) (favoriteModel.State, error)
}

var localFavorites IFavorites

func Favorites() IFavorites {
	if localFavorites == nil {
		panic("implement not found for interface IFavorites, forgot register?")
	}
	return localFavorites
}

func RegisterFavorites(i IFavorites) {
	localFavorites = i
}
