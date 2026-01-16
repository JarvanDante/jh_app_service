package frontend

import (
	"context"

	v1 "jh_app_service/api/frontend/public/v1"
)

type (
	IPublic interface {
		QuickRegister(ctx context.Context, req *v1.QuickRegisterReq) (*v1.QuickRegisterRes, error)
		Login(ctx context.Context, req *v1.LoginReq) (*v1.LoginRes, error)
	}
)

var (
	localPublic IPublic
)

func Public() IPublic {
	if localPublic == nil {
		panic("implement not found for interface IPublic, forgot register?")
	}
	return localPublic
}

func RegisterPublic(i IPublic) {
	localPublic = i
}
