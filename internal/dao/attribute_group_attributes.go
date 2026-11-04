package dao

import "github.com/suxiangdong/indu-scope-backend/internal/dao/internal"

type attributeGroupAttributesDao struct {
	*internal.AttributeGroupAttributesDao
}

var AttributeGroupAttributes = attributeGroupAttributesDao{internal.NewAttributeGroupAttributesDao()}
