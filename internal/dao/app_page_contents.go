package dao

import "github.com/suxiangdong/indu-scope-backend/internal/dao/internal"

type appPageContentsDao struct{ *internal.AppPageContentsDao }

var AppPageContents = appPageContentsDao{internal.NewAppPageContentsDao()}
