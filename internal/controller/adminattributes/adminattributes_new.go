package adminattributes

import (
	"github.com/suxiangdong/indu-scope-backend/api/adminattributes"
)

func NewV1() adminattributes.IAdminattributesV1 {
	return &ControllerV1{}
}
