package adminproducts

import "github.com/suxiangdong/indu-scope-backend/api/adminproducts"

type ControllerV1 struct{}

func NewV1() adminproducts.IAdminProductsV1 { return &ControllerV1{} }
