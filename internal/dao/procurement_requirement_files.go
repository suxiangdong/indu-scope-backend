package dao

import "github.com/suxiangdong/indu-scope-backend/internal/dao/internal"

type procurementRequirementFilesDao struct {
	*internal.ProcurementRequirementFilesDao
}

var ProcurementRequirementFiles = procurementRequirementFilesDao{internal.NewProcurementRequirementFilesDao()}
