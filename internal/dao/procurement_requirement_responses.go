package dao

import "github.com/suxiangdong/indu-scope-backend/internal/dao/internal"

type procurementRequirementResponsesDao struct {
	*internal.ProcurementRequirementResponsesDao
}

var ProcurementRequirementResponses = procurementRequirementResponsesDao{internal.NewProcurementRequirementResponsesDao()}
