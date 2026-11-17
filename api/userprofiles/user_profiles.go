package userprofiles

import (
	"context"

	"github.com/suxiangdong/indu-scope-backend/api/userprofiles/v1"
)

type IUserProfilesV1 interface {
	Profile(ctx context.Context, req *v1.ProfileReq) (*v1.ProfileRes, error)
}
