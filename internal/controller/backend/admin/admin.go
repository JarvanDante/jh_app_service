package admin

import (
	"context"
	v2 "jh_app_service/api/backend/admin/v1"
	"jh_app_service/internal/service/backend"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

type Controller struct {
	v2.UnimplementedAdminServer
}

func Register(s *grpcx.GrpcServer) {
	v2.RegisterAdminServer(s.Server, &Controller{})
}

// Login 管理员登录
func (*Controller) Login(ctx context.Context, req *v2.LoginReq) (res *v2.LoginRes, err error) {
	return backend.Admin().Login(ctx, req)
}

// RefreshToken 刷新token
func (*Controller) RefreshToken(ctx context.Context, req *v2.RefreshTokenReq) (res *v2.RefreshTokenRes, err error) {
	return backend.Admin().RefreshToken(ctx, req)
}

// GetInfo 获取管理员信息
func (*Controller) GetInfo(ctx context.Context, req *v2.GetInfoReq) (res *v2.GetInfoRes, err error) {
	return backend.Admin().GetInfo(ctx, req)
}

// Menus 获取菜单列表
func (*Controller) Menus(ctx context.Context, req *v2.MenusReq) (res *v2.MenusRes, err error) {
	return backend.Admin().Menus(ctx, req)
}

// CreateAdmin 创建管理员
func (*Controller) CreateAdmin(ctx context.Context, req *v2.CreateAdminReq) (res *v2.CreateAdminRes, err error) {
	return backend.Admin().CreateAdmin(ctx, req)
}

// GetAdminList 获取管理员列表
func (*Controller) GetAdminList(ctx context.Context, req *v2.GetAdminListReq) (res *v2.GetAdminListRes, err error) {
	return backend.Admin().GetAdminList(ctx, req)
}

// UpdateAdmin 更新管理员
func (*Controller) UpdateAdmin(ctx context.Context, req *v2.UpdateAdminReq) (res *v2.UpdateAdminRes, err error) {
	return backend.Admin().UpdateAdmin(ctx, req)
}

// DeleteAdmin 删除管理员
func (*Controller) DeleteAdmin(ctx context.Context, req *v2.DeleteAdminReq) (res *v2.DeleteAdminRes, err error) {
	return backend.Admin().DeleteAdmin(ctx, req)
}

// Logout 退出登录
func (*Controller) Logout(ctx context.Context, req *v2.LogoutReq) (res *v2.LogoutRes, err error) {
	return backend.Admin().Logout(ctx, req)
}

// ChangePassword 修改密码
func (*Controller) ChangePassword(ctx context.Context, req *v2.ChangePasswordReq) (res *v2.ChangePasswordRes, err error) {
	return backend.Admin().ChangePassword(ctx, req)
}

// GetAdminLogs 获取管理员日志列表
func (*Controller) GetAdminLogs(ctx context.Context, req *v2.GetAdminLogsReq) (res *v2.GetAdminLogsRes, err error) {
	return backend.Admin().GetAdminLogs(ctx, req)
}

func (*Controller) GenerateGoogle2FA(ctx context.Context, req *v2.GenerateGoogle2FAReq) (res *v2.GenerateGoogle2FARes, err error) {
	return backend.Admin().GenerateGoogle2FA(ctx, req)
}

func (*Controller) BindGoogle2FA(ctx context.Context, req *v2.BindGoogle2FAReq) (res *v2.BindGoogle2FARes, err error) {
	return backend.Admin().BindGoogle2FA(ctx, req)
}
