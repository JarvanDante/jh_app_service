package frontend

import (
	"context"

	v1 "jh_app_service/api/frontend/user/v1"
)

type (
	IUser interface {
		GetUserInfo(ctx context.Context, req *v1.GetUserInfoReq) (*v1.GetUserInfoRes, error)
		Logout(ctx context.Context, req *v1.LogoutReq) (*v1.LogoutRes, error)
		BindEmail(ctx context.Context, req *v1.BindEmailReq) (*v1.BindEmailRes, error)
		BindMobile(ctx context.Context, req *v1.BindMobileReq) (*v1.BindMobileRes, error)
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
