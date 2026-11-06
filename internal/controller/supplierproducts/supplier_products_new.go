package supplierproducts

import (
	"github.com/suxiangdong/indu-scope-backend/api/supplierproducts"
)

type ControllerV1 struct{}

func NewV1() supplierproducts.ISupplierProductsV1 {
	return &ControllerV1{}
}
