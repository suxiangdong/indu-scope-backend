package dao

import "github.com/suxiangdong/indu-scope-backend/internal/dao/internal"

type categoryAttributeGroupBindingsDao struct {
	*internal.CategoryAttributeGroupBindingsDao
}

var CategoryAttributeGroupBindings = categoryAttributeGroupBindingsDao{internal.NewCategoryAttributeGroupBindingsDao()}
