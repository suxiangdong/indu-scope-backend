package dao

import "github.com/suxiangdong/indu-scope-backend/internal/dao/internal"

type attributeGroupsDao struct {
	*internal.AttributeGroupsDao
}

var AttributeGroups = attributeGroupsDao{internal.NewAttributeGroupsDao()}
