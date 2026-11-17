package analytics

import (
	"context"

	v1 "github.com/suxiangdong/indu-scope-backend/api/analytics/v1"
)

type IAnalyticsV1 interface {
	RecordClick(ctx context.Context, req *v1.RecordClickReq) (*v1.RecordClickRes, error)
}
