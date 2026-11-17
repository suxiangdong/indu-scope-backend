package service

import (
	"context"

	userprofileModel "github.com/suxiangdong/indu-scope-backend/internal/model/userprofile"
)

type IUserProfile interface {
	Profile(ctx context.Context, userID uint64) (userprofileModel.Profile, error)
}

var localUserProfile IUserProfile

func UserProfile() IUserProfile {
	if localUserProfile == nil {
		panic("implement not found for interface IUserProfile, forgot register?")
	}
	return localUserProfile
}

func RegisterUserProfile(i IUserProfile) {
	localUserProfile = i
}
