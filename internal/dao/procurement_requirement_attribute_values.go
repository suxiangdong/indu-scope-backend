package dao

import "github.com/suxiangdong/indu-scope-backend/internal/dao/internal"

type procurementRequirementAttributeValuesDao struct {
	*internal.ProcurementRequirementAttributeValuesDao
}

var ProcurementRequirementAttributeValues = procurementRequirementAttributeValuesDao{internal.NewProcurementRequirementAttributeValuesDao()}
