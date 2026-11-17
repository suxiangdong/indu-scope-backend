package adminanalytics

import "github.com/suxiangdong/indu-scope-backend/api/adminanalytics"

type ControllerV1 struct{}

func NewV1() adminanalytics.IAdminAnalyticsV1 { return &ControllerV1{} }
