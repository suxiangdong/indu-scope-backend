package productcatalog

import "github.com/suxiangdong/indu-scope-backend/api/products"

type ControllerV1 struct{}

func NewV1() products.IProductsV1 { return &ControllerV1{} }
