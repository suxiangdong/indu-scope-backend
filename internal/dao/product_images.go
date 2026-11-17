package dao

import "github.com/suxiangdong/indu-scope-backend/internal/dao/internal"

type productImagesDao struct{ *internal.ProductImagesDao }

var ProductImages = productImagesDao{internal.NewProductImagesDao()}
