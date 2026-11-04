package bizctx

import (
	"context"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	model "github.com/suxiangdong/indu-scope-backend/internal/model/bizctx"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

type sBizctx struct{}

var _ service.IBizctx = (*sBizctx)(nil)

// New creates the request business-context service.
func New() service.IBizctx {
	return &sBizctx{}
}

func init() {
	service.RegisterBizctx(New())
}

// Set attaches the custom business context to the standard Go context.
func (s *sBizctx) Set(ctx context.Context, value *model.Context) context.Context {
	return context.WithValue(ctx, consts.BizctxContextKey, value)
}

// Get retrieves the custom business context from the standard Go context.
func (s *sBizctx) Get(ctx context.Context) *model.Context {
	if ctx == nil {
		return nil
	}
	value, _ := ctx.Value(consts.BizctxContextKey).(*model.Context)
	return value
}
