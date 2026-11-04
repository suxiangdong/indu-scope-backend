// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"

	model "github.com/suxiangdong/indu-scope-backend/internal/model/bizctx"
)

type (
	IBizctx interface {
		// Set attaches the custom business context to the standard Go context.
		Set(ctx context.Context, value *model.Context) context.Context
		// Get retrieves the custom business context from the standard Go context.
		Get(ctx context.Context) *model.Context
	}
)

var (
	localBizctx IBizctx
)

func Bizctx() IBizctx {
	if localBizctx == nil {
		panic("implement not found for interface IBizctx, forgot register?")
	}
	return localBizctx
}

func RegisterBizctx(i IBizctx) {
	localBizctx = i
}
