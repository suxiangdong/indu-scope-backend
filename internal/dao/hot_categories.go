package dao

import "github.com/suxiangdong/indu-scope-backend/internal/dao/internal"

type hotCategoriesDao struct{ *internal.HotCategoriesDao }

var HotCategories = hotCategoriesDao{internal.NewHotCategoriesDao()}
