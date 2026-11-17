package auth

import (
	"github.com/suxiangdong/indu-scope-backend/api/auth/v1"
	authmodel "github.com/suxiangdong/indu-scope-backend/internal/model/auth"
)

func toUserInfo(user authmodel.User) v1.UserInfo {
	return v1.UserInfo{
		ID:          user.ID,
		Phone:       user.Phone,
		CompanyName: user.CompanyName,
		Role:        user.Role,
		UserType:    user.UserType,
		Status:      user.Status,
	}
}
