package internal

import "github.com/gogf/gf/v2/database/gdb"

type ProcurementRequirementAttributeValueOptionsDao struct{ *TableDao }

func NewProcurementRequirementAttributeValueOptionsDao(handlers ...gdb.ModelHandler) *ProcurementRequirementAttributeValueOptionsDao {
	return &ProcurementRequirementAttributeValueOptionsDao{TableDao: NewTableDao("procurement_requirement_attribute_value_options", handlers...)}
}
