package appcontent

import (
	"context"
	"github.com/suxiangdong/indu-scope-backend/api/appcontent/v1"
)

type IAppContentV1 interface {
	Get(ctx context.Context, req *v1.GetReq) (*v1.GetRes, error)
	Agreements(ctx context.Context, req *v1.AgreementsReq) (*v1.AgreementsRes, error)
}
