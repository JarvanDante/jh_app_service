package public

import (
	"context"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"

	v1 "jh_app_service/api/frontend/public/v1"
	"jh_app_service/internal/service/frontend"
)

type Controller struct {
	v1.UnimplementedPublicServer
}

func Register(s *grpcx.GrpcServer) {
	v1.RegisterPublicServer(s.Server, &Controller{})
}

// QuickRegister 快速注册
func (c *Controller) QuickRegister(ctx context.Context, req *v1.QuickRegisterReq) (*v1.QuickRegisterRes, error) {
	return frontend.Public().QuickRegister(ctx, req)
}

// Login 用户登录
func (c *Controller) Login(ctx context.Context, req *v1.LoginReq) (*v1.LoginRes, error) {
	return frontend.Public().Login(ctx, req)
}

// GenerateCaptcha 生成验证码
func (c *Controller) GenerateCaptcha(ctx context.Context, req *v1.GenerateCaptchaReq) (*v1.GenerateCaptchaRes, error) {
	return frontend.Public().GenerateCaptcha(ctx, req)
}

// RefreshToken 刷新token
func (c *Controller) RefreshToken(ctx context.Context, req *v1.RefreshTokenReq) (*v1.RefreshTokenRes, error) {
	return frontend.Public().RefreshToken(ctx, req)
}

// Register 用户注册
func (c *Controller) Register(ctx context.Context, req *v1.RegisterReq) (*v1.RegisterRes, error) {
	return frontend.Public().Register(ctx, req)
}

// GetSiteSetting 获取站点设置
func (c *Controller) GetSiteSetting(ctx context.Context, req *v1.GetSiteSettingReq) (*v1.GetSiteSettingRes, error) {
	return frontend.Public().GetSiteSetting(ctx, req)
}
