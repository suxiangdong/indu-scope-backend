package categories

import "github.com/suxiangdong/indu-scope-backend/api/categories"

type ControllerV1 struct{}

func NewV1() categories.ICategoriesV1 { return &ControllerV1{} }
