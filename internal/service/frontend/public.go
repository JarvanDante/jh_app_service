package frontend

import (
	"context"

	v1 "jh_app_service/api/frontend/public/v1"
)

type (
	IPublic interface {
		QuickRegister(ctx context.Context, req *v1.QuickRegisterReq) (*v1.QuickRegisterRes, error)
		Register(ctx context.Context, req *v1.RegisterReq) (*v1.RegisterRes, error)
		Login(ctx context.Context, req *v1.LoginReq) (*v1.LoginRes, error)
		GenerateCaptcha(ctx context.Context, req *v1.GenerateCaptchaReq) (*v1.GenerateCaptchaRes, error)
		RefreshToken(ctx context.Context, req *v1.RefreshTokenReq) (*v1.RefreshTokenRes, error)
		GetSiteSetting(ctx context.Context, req *v1.GetSiteSettingReq) (*v1.GetSiteSettingRes, error)
		GetAppNoticeList(ctx context.Context, req *v1.GetNoticeListReq) (*v1.GetNoticeListRes, error)
		GetAppAdList(ctx context.Context, req *v1.GetAdListReq) (*v1.GetAdListRes, error)
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
