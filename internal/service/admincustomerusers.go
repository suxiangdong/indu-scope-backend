package service

import (
	"context"

	adminCustomerUserModel "github.com/suxiangdong/indu-scope-backend/internal/model/admincustomeruser"
)

type IAdminCustomerUsers interface {
	List(ctx context.Context, filter adminCustomerUserModel.ListFilter) (adminCustomerUserModel.ListResult, error)
	UpdateStatus(ctx context.Context, id uint64, status uint8) (adminCustomerUserModel.UserItem, error)
}

var localAdminCustomerUsers IAdminCustomerUsers

func AdminCustomerUsers() IAdminCustomerUsers {
	if localAdminCustomerUsers == nil {
		panic("implement not found for interface IAdminCustomerUsers, forgot register?")
	}
	return localAdminCustomerUsers
}

func RegisterAdminCustomerUsers(i IAdminCustomerUsers) {
	localAdminCustomerUsers = i
}
