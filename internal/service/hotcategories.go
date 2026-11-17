package service

import (
	"context"

	hotcategoryModel "github.com/suxiangdong/indu-scope-backend/internal/model/hotcategory"
)

type IHotCategories interface {
	List(ctx context.Context, filter hotcategoryModel.ListFilter) (hotcategoryModel.ListResult, error)
	Create(ctx context.Context, categoryID uint64, sortOrder uint, status uint8) (hotcategoryModel.Item, error)
	Update(ctx context.Context, id, categoryID uint64, sortOrder uint, status uint8) (hotcategoryModel.Item, error)
	Delete(ctx context.Context, id uint64) error
	UpdateStatus(ctx context.Context, id uint64, status uint8) (hotcategoryModel.Item, error)
	Sort(ctx context.Context, items []hotcategoryModel.SortItem) (hotcategoryModel.ListResult, error)
	ListPublic(ctx context.Context) (hotcategoryModel.ListResult, error)
}

var localHotCategories IHotCategories

func HotCategories() IHotCategories {
	if localHotCategories == nil {
		panic("implement not found for interface IHotCategories, forgot register?")
	}
	return localHotCategories
}

func RegisterHotCategories(i IHotCategories) {
	localHotCategories = i
}
