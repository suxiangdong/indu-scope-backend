package internal

import "github.com/gogf/gf/v2/database/gdb"

type ProcurementRequirementAttributeValuesDao struct{ *TableDao }

func NewProcurementRequirementAttributeValuesDao(handlers ...gdb.ModelHandler) *ProcurementRequirementAttributeValuesDao {
	return &ProcurementRequirementAttributeValuesDao{TableDao: NewTableDao("procurement_requirement_attribute_values", handlers...)}
}
