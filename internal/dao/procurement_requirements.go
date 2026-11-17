package dao

import "github.com/suxiangdong/indu-scope-backend/internal/dao/internal"

type procurementRequirementsDao struct {
	*internal.ProcurementRequirementsDao
}

var ProcurementRequirements = procurementRequirementsDao{internal.NewProcurementRequirementsDao()}
