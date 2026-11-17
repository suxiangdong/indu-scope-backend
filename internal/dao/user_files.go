package dao

import "github.com/suxiangdong/indu-scope-backend/internal/dao/internal"

type userFilesDao struct{ *internal.UserFilesDao }

var UserFiles = userFilesDao{internal.NewUserFilesDao()}
