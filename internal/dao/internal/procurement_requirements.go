package internal

import "github.com/gogf/gf/v2/database/gdb"

type ProcurementRequirementsDao struct{ *TableDao }

func NewProcurementRequirementsDao(handlers ...gdb.ModelHandler) *ProcurementRequirementsDao {
	return &ProcurementRequirementsDao{TableDao: NewTableDao("procurement_requirements", handlers...)}
}
