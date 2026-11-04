// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ================================================================================

package service

import (
	"context"

	categorymodel "github.com/suxiangdong/indu-scope-backend/internal/model/category"
)

type (
	IAdminCategories interface {
		// List returns all categories ordered by parent, sort order, and ID.
		List(ctx context.Context) (categorymodel.CategoryListResult, error)
		// Create creates a category.
		Create(ctx context.Context, parentID *uint64, name string, sortOrder uint, status uint) (categorymodel.CategoryItem, error)
		// Update updates a category and its hierarchy placement.
		Update(ctx context.Context, id uint64, parentID *uint64, name string, sortOrder uint, status uint) (categorymodel.CategoryItem, error)
		// Delete deletes an unused leaf category.
		Delete(ctx context.Context, id uint64) error
	}
)

var (
	localAdminCategories IAdminCategories
)

func AdminCategories() IAdminCategories {
	if localAdminCategories == nil {
		panic("implement not found for interface IAdminCategories, forgot register?")
	}
	return localAdminCategories
}

func RegisterAdminCategories(i IAdminCategories) {
	localAdminCategories = i
}
