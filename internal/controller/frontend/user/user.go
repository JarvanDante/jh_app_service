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

// BindEmail 绑定邮箱
func (c *Controller) BindEmail(ctx context.Context, req *v1.BindEmailReq) (*v1.BindEmailRes, error) {
	return frontend.User().BindEmail(ctx, req)
}

// BindMobile 绑定手机
func (c *Controller) BindMobile(ctx context.Context, req *v1.BindMobileReq) (*v1.BindMobileRes, error) {
	return frontend.User().BindMobile(ctx, req)
}

// ChangeLoginPassword 修改登录密码
func (c *Controller) ChangeLoginPassword(ctx context.Context, req *v1.ChangeLoginPasswordReq) (*v1.ChangeLoginPasswordRes, error) {
	return frontend.User().ChangeLoginPassword(ctx, req)
}

// BindLoginPassword 绑定登录密码（针对快速注册用户）
func (c *Controller) BindLoginPassword(ctx context.Context, req *v1.BindLoginPasswordReq) (*v1.BindLoginPasswordRes, error) {
	return frontend.User().BindLoginPassword(ctx, req)
}

// GetGradeList 获取会员等级列表
func (c *Controller) GetGradeList(ctx context.Context, req *v1.GetGradeListReq) (*v1.GetGradeListRes, error) {
	return frontend.User().GetGradeList(ctx, req)
}

// SetPayPassword 设置支付密码
func (c *Controller) SetPayPassword(ctx context.Context, req *v1.SetPayPasswordReq) (*v1.SetPayPasswordRes, error) {
	return frontend.User().SetPayPassword(ctx, req)
}

// VerifyPayPassword 验证支付密码
func (c *Controller) VerifyPayPassword(ctx context.Context, req *v1.VerifyPayPasswordReq) (*v1.VerifyPayPasswordRes, error) {
	return frontend.User().VerifyPayPassword(ctx, req)
}

// ChangePayPassword 修改支付密码
func (c *Controller) ChangePayPassword(ctx context.Context, req *v1.ChangePayPasswordReq) (*v1.ChangePayPasswordRes, error) {
	return frontend.User().ChangePayPassword(ctx, req)
}

// GetInviteInfo 获取邀请信息
func (c *Controller) GetInviteInfo(ctx context.Context, req *v1.GetInviteInfoReq) (*v1.GetInviteInfoRes, error) {
	return frontend.User().GetInviteInfo(ctx, req)
}

// GetInviteList 获取邀请列表
func (c *Controller) GetInviteList(ctx context.Context, req *v1.GetInviteListReq) (*v1.GetInviteListRes, error) {
	return frontend.User().GetInviteList(ctx, req)
}
