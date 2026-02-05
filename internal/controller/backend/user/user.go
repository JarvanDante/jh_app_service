package user

import (
	"context"
	v1 "jh_app_service/api/backend/user/v1"
	"jh_app_service/internal/service/backend"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

type Controller struct {
	v1.UnimplementedUserServer
}

func Register(s *grpcx.GrpcServer) {
	v1.RegisterUserServer(s.Server, &Controller{})
}

// GetUserList 获取用户列表
func (*Controller) GetUserList(ctx context.Context, req *v1.GetUserListReq) (res *v1.GetUserListRes, err error) {
	return backend.User().GetUserList(ctx, req)
}

// UpdateUser 更新用户信息
func (*Controller) UpdateUser(ctx context.Context, req *v1.UpdateUserReq) (res *v1.UpdateUserRes, err error) {
	return backend.User().UpdateUser(ctx, req)
}

// GetUserBasicInfo 获取用户基本信息
func (*Controller) GetUserBasicInfo(ctx context.Context, req *v1.GetUserBasicInfoReq) (res *v1.GetUserBasicInfoRes, err error) {
	return backend.User().GetUserBasicInfo(ctx, req)
}

// IncreasePayTimes 增加用户充值次数
func (*Controller) IncreasePayTimes(ctx context.Context, req *v1.IncreasePayTimesReq) (res *v1.IncreasePayTimesRes, err error) {
	return backend.User().IncreasePayTimes(ctx, req)
}

// GetUserGrades 获取用户等级列表
func (*Controller) GetUserGrades(ctx context.Context, req *v1.GetUserGradesReq) (res *v1.GetUserGradesRes, err error) {
	return backend.User().GetUserGrades(ctx, req)
}

// SaveUserGrades 保存用户等级
func (*Controller) SaveUserGrades(ctx context.Context, req *v1.SaveUserGradesReq) (res *v1.SaveUserGradesRes, err error) {
	return backend.User().SaveUserGrades(ctx, req)
}

// DeleteUserGrades 删除用户等级
func (*Controller) DeleteUserGrades(ctx context.Context, req *v1.DeleteUserGradesReq) (res *v1.DeleteUserGradesRes, err error) {
	return backend.User().DeleteUserGrades(ctx, req)
}

// GetUserLoginLogs 获取用户登录日志
func (*Controller) GetUserLoginLogs(ctx context.Context, req *v1.GetUserLoginLogsReq) (res *v1.GetUserLoginLogsRes, err error) {
	return backend.User().GetUserLoginLogs(ctx, req)
}
