package favorites

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"

	"github.com/suxiangdong/indu-scope-backend/api/favorites/v1"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	favoriteModel "github.com/suxiangdong/indu-scope-backend/internal/model/favorite"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (*v1.ListRes, error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	var page, pageSize uint
	if req != nil {
		page, pageSize = req.Page, req.PageSize
	}
	result, err := service.Favorites().List(ctx, userID, page, pageSize)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	items := make([]v1.FavoriteProductInfo, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toAPI(item))
	}
	return &v1.ListRes{Items: items, Total: result.Total, Page: result.Page, PageSize: result.PageSize}, nil
}

func (c *ControllerV1) Toggle(ctx context.Context, req *v1.ToggleReq) (*v1.ToggleRes, error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "product_id is required")
	}
	state, err := service.Favorites().Toggle(ctx, userID, req.ProductID)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.ToggleRes{FavoriteState: toAPIState(state)}, nil
}

func (c *ControllerV1) Status(ctx context.Context, req *v1.StatusReq) (*v1.StatusRes, error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "product_id is required")
	}
	state, err := service.Favorites().Status(ctx, userID, req.ProductID)
	if err != nil {
		return nil, toAPIError(ctx, err)
	}
	return &v1.StatusRes{FavoriteState: toAPIState(state)}, nil
}

func currentUserID(ctx context.Context) (uint64, error) {
	value := service.Bizctx().Get(ctx)
	if value == nil || value.User == nil {
		return 0, gerror.NewCode(gcode.CodeNotAuthorized, "authorization is required")
	}
	if value.User.Status != 1 || value.User.Role != 1 {
		return 0, gerror.NewCode(gcode.CodeNotAuthorized, "buyer authorization is required")
	}
	return uint64(value.User.ID), nil
}

func toAPI(item favoriteModel.ProductItem) v1.FavoriteProductInfo {
	return v1.FavoriteProductInfo{
		FavoriteID: item.ID, ProductID: item.ProductID, ProductName: item.ProductName,
		SupplierUserID: item.SupplierUserID,
		CategoryID:     item.CategoryID, CategoryName: item.CategoryName,
		StockQuantity: item.StockQuantity, StockUnitName: item.StockUnitName,
		CoverURL: item.CoverURL, ImageCount: item.ImageCount, Status: item.Status,
		FavoritedAt: item.FavoritedAt,
	}
}

func toAPIState(state favoriteModel.State) v1.FavoriteState {
	return v1.FavoriteState{ProductID: state.ProductID, Favorited: state.Favorited, FavoriteID: state.FavoriteID}
}

func toAPIError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, consts.ErrProductNotFound), errors.Is(err, consts.ErrProductUnavailable):
		return gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
	case errors.Is(err, consts.ErrTokenInvalid):
		return gerror.NewCode(gcode.CodeNotAuthorized, err.Error())
	}
	var parameterError *consts.InvalidParameterError
	if errors.As(err, &parameterError) {
		return gerror.NewCode(gcode.CodeInvalidParameter, parameterError.Error())
	}
	glog.Errorf(ctx, "favorites internal error: %+v", err)
	return gerror.NewCode(gcode.CodeInternalError, "internal server error")
}
