package analytics

import "github.com/suxiangdong/indu-scope-backend/api/analytics"

type ControllerV1 struct{}

func NewV1() analytics.IAnalyticsV1 { return &ControllerV1{} }
