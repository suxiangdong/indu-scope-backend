package dao

import "github.com/suxiangdong/indu-scope-backend/internal/dao/internal"

type procurementRequirementAttributeValueOptionsDao struct {
	*internal.ProcurementRequirementAttributeValueOptionsDao
}

var ProcurementRequirementAttributeValueOptions = procurementRequirementAttributeValueOptionsDao{internal.NewProcurementRequirementAttributeValueOptionsDao()}
