package dao

import "github.com/suxiangdong/indu-scope-backend/internal/dao/internal"

type categoryAttributeGroupsDao struct {
	*internal.CategoryAttributeGroupsDao
}

var CategoryAttributeGroups = categoryAttributeGroupsDao{internal.NewCategoryAttributeGroupsDao()}
