package backend

import (
	"context"
	v1 "jh_app_service/api/backend/user/v1"
)

type (
	IUser interface {
		IncreasePayTimes(ctx context.Context, req *v1.IncreasePayTimesReq) (*v1.IncreasePayTimesRes, error)
		GetUserList(ctx context.Context, req *v1.GetUserListReq) (*v1.GetUserListRes, error)
		UpdateUser(ctx context.Context, req *v1.UpdateUserReq) (*v1.UpdateUserRes, error)
		GetUserBasicInfo(ctx context.Context, req *v1.GetUserBasicInfoReq) (*v1.GetUserBasicInfoRes, error)

		// UserGrade相关方法
		GetUserGrades(ctx context.Context, req *v1.GetUserGradesReq) (*v1.GetUserGradesRes, error)
		SaveUserGrades(ctx context.Context, req *v1.SaveUserGradesReq) (*v1.SaveUserGradesRes, error)
		DeleteUserGrades(ctx context.Context, req *v1.DeleteUserGradesReq) (*v1.DeleteUserGradesRes, error)

		// UserLoginLog相关方法
		GetUserLoginLogs(ctx context.Context, req *v1.GetUserLoginLogsReq) (*v1.GetUserLoginLogsRes, error)
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
