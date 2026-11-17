package appcontent

import "github.com/suxiangdong/indu-scope-backend/api/appcontent"

type ControllerV1 struct{}

func NewV1() appcontent.IAppContentV1 { return &ControllerV1{} }
