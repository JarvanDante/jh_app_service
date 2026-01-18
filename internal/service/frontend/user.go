package frontend

import (
	"context"

	v1 "jh_app_service/api/frontend/user/v1"
)

type (
	IUser interface {
		GetUserInfo(ctx context.Context, req *v1.GetUserInfoReq) (*v1.GetUserInfoRes, error)
	}
)

var (
	localUser IUser
)

func User() IUser {
	if localUser == nil {
		panic("implement not found for interface IUser, forgot register?")
	}
	return localUser
}

func RegisterUser(i IUser) {
	localUser = i
}
