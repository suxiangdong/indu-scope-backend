package userprofile

import "github.com/suxiangdong/indu-scope-backend/api/userprofiles"

type ControllerV1 struct{}

func NewV1() userprofiles.IUserProfilesV1 { return &ControllerV1{} }
