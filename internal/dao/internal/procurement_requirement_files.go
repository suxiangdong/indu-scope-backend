package internal

import "github.com/gogf/gf/v2/database/gdb"

type ProcurementRequirementFilesDao struct{ *TableDao }

func NewProcurementRequirementFilesDao(handlers ...gdb.ModelHandler) *ProcurementRequirementFilesDao {
	return &ProcurementRequirementFilesDao{TableDao: NewTableDao("procurement_requirement_files", handlers...)}
}
