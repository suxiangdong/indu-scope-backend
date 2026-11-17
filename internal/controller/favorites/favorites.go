package favorites

import "github.com/suxiangdong/indu-scope-backend/api/favorites"

type ControllerV1 struct{}

func NewV1() favorites.IFavoritesV1 { return &ControllerV1{} }
