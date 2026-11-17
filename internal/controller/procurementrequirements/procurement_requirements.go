package procurementrequirements

import "github.com/suxiangdong/indu-scope-backend/api/procurementrequirements"

type ControllerV1 struct{}

func NewV1() procurementrequirements.IProcurementRequirementsV1 { return &ControllerV1{} }
