package adminanalytics

import (
	"context"

	v1 "github.com/suxiangdong/indu-scope-backend/api/adminanalytics/v1"
)

type IAdminAnalyticsV1 interface {
	Overview(ctx context.Context, req *v1.OverviewReq) (*v1.OverviewRes, error)
}
