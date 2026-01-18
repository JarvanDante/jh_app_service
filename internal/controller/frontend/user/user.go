package user

import (
	"context"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"

	v1 "jh_app_service/api/frontend/user/v1"
	"jh_app_service/internal/service/frontend"
)

type Controller struct {
	v1.UnimplementedUserServer
}

func Register(s *grpcx.GrpcServer) {
	v1.RegisterUserServer(s.Server, &Controller{})
}

// GetUserInfo 获取用户信息
func (c *Controller) GetUserInfo(ctx context.Context, req *v1.GetUserInfoReq) (*v1.GetUserInfoRes, error) {
	return frontend.User().GetUserInfo(ctx, req)
}

// Logout 用户退出登录
func (c *Controller) Logout(ctx context.Context, req *v1.LogoutReq) (*v1.LogoutRes, error) {
	return frontend.User().Logout(ctx, req)
}
