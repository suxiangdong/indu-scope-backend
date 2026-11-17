package internal

import "github.com/gogf/gf/v2/database/gdb"

type ProcurementRequirementResponsesDao struct{ *TableDao }

func NewProcurementRequirementResponsesDao(handlers ...gdb.ModelHandler) *ProcurementRequirementResponsesDao {
	return &ProcurementRequirementResponsesDao{TableDao: NewTableDao("procurement_requirement_responses", handlers...)}
}
