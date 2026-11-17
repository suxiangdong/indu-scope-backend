package adminhotcategories

import "github.com/suxiangdong/indu-scope-backend/api/adminhotcategories"

type ControllerV1 struct{}

func NewV1() adminhotcategories.IAdminHotCategoriesV1 { return &ControllerV1{} }
