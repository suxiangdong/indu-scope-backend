package dao

import "github.com/suxiangdong/indu-scope-backend/internal/dao/internal"

type productFavoritesDao struct{ *internal.ProductFavoritesDao }

var ProductFavorites = productFavoritesDao{internal.NewProductFavoritesDao()}
