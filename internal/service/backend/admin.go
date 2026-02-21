package backend

import (
	"context"
	v1 "jh_app_service/api/backend/admin/v1"

	"github.com/gogf/gf/v2/os/gtime"

	"jh_app_service/internal/dao"
	"jh_app_service/internal/model/do"
	"jh_app_service/internal/model/entity"
)

type (
	IAdmin interface {
		Login(ctx context.Context, req *v1.LoginReq) (*v1.LoginRes, error)
		RefreshToken(ctx context.Context, req *v1.RefreshTokenReq) (*v1.RefreshTokenRes, error)
		GetInfo(ctx context.Context, req *v1.GetInfoReq) (*v1.GetInfoRes, error)
		Menus(ctx context.Context, req *v1.MenusReq) (*v1.MenusRes, error)
		GetAdminList(ctx context.Context, req *v1.GetAdminListReq) (*v1.GetAdminListRes, error)
		CreateAdmin(ctx context.Context, req *v1.CreateAdminReq) (*v1.CreateAdminRes, error)
		UpdateAdmin(ctx context.Context, req *v1.UpdateAdminReq) (*v1.UpdateAdminRes, error)
		DeleteAdmin(ctx context.Context, req *v1.DeleteAdminReq) (*v1.DeleteAdminRes, error)
		Logout(ctx context.Context, req *v1.LogoutReq) (*v1.LogoutRes, error)
		ChangePassword(ctx context.Context, req *v1.ChangePasswordReq) (*v1.ChangePasswordRes, error)
		GetAdminLogs(ctx context.Context, req *v1.GetAdminLogsReq) (*v1.GetAdminLogsRes, error)
		GenerateGoogle2FA(ctx context.Context, req *v1.GenerateGoogle2FAReq) (*v1.GenerateGoogle2FARes, error)
		BindGoogle2FA(ctx context.Context, req *v1.BindGoogle2FAReq) (*v1.BindGoogle2FARes, error)
	}
)

var (
	localAdmin IAdmin
)

func Admin() IAdmin {
	if localAdmin == nil {
		panic("implement not found for interface IAdmin, forgot register?")
	}
	return localAdmin
}

func RegisterAdmin(i IAdmin) {
	localAdmin = i
}

type sAdmin struct{}

// GetByUsername 根据用户名获取管理员
func (s *sAdmin) GetByUsername(ctx context.Context, username string, siteId int) (*entity.Admin, error) {
	var admin *entity.Admin
	err := dao.Admin.Ctx(ctx).Where(do.Admin{
		Username: username,
		SiteId:   siteId,
		Status:   1, // 正常状态
	}).Scan(&admin)

	return admin, err
}

// GetById 根据ID获取管理员
func (s *sAdmin) GetById(ctx context.Context, id uint) (*entity.Admin, error) {
	var admin *entity.Admin
	err := dao.Admin.Ctx(ctx).Where(do.Admin{
		Id:     id,
		Status: 1, // 正常状态
	}).Scan(&admin)

	return admin, err
}

// UpdateLastLoginInfo 更新最后登录信息
func (s *sAdmin) UpdateLastLoginInfo(ctx context.Context, adminId uint, ip string) error {
	_, err := dao.Admin.Ctx(ctx).Where(do.Admin{Id: adminId}).Update(do.Admin{
		LastLoginIp:   ip,
		LastLoginTime: gtime.Now(),
	})
	return err
}

// AddLog 添加管理员操作日志
func (s *sAdmin) AddLog(ctx context.Context, admin *entity.Admin, message string, ip string) error {
	_, err := dao.AdminLog.Ctx(ctx).Insert(do.AdminLog{
		SiteId:        admin.SiteId,
		AdminId:       int(admin.Id),
		AdminUsername: admin.Username,
		Ip:            ip,
		Remark:        message,
		CreatedAt:     gtime.Now(),
	})
	return err
}

// CreateAdmin 创建管理员
func (s *sAdmin) CreateAdmin(ctx context.Context, siteId int, username, nickname, password string, roleId, status int) error {
	// 插入管理员记录
	_, err := dao.Admin.Ctx(ctx).Insert(do.Admin{
		SiteId:      siteId,
		Username:    username,
		Nickname:    nickname,
		Password:    password, // 密码应该已经在控制器中加密
		AdminRoleId: roleId,
		Status:      status,
		CreatedAt:   gtime.Now(),
		UpdatedAt:   gtime.Now(),
	})

	return err
}

// CheckUsernameExists 检查用户名是否已存在
func (s *sAdmin) CheckUsernameExists(ctx context.Context, username string, siteId int) (bool, error) {
	var admin *entity.Admin
	err := dao.Admin.Ctx(ctx).Where(do.Admin{
		Username: username,
		SiteId:   siteId,
	}).Scan(&admin)

	if err != nil {
		return false, err
	}

	return admin != nil, nil
}
