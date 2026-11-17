package adminprocurementrequirements

import "github.com/suxiangdong/indu-scope-backend/api/adminprocurementrequirements"

type ControllerV1 struct{}

func NewV1() adminprocurementrequirements.IAdminProcurementRequirementsV1 { return &ControllerV1{} }
