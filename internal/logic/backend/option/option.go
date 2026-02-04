package option

import (
	"context"
	"jh_app_service/api/backend/option/v1"
	"jh_app_service/internal/dao"
	"jh_app_service/internal/middleware"
	"jh_app_service/internal/model/do"
	"jh_app_service/internal/model/entity"
	"jh_app_service/internal/service/backend"
)

type (
	sOption struct{}
)

func init() {
	backend.RegisterOption(&sOption{})
}

// GetUserGradeList 获取会员等级列表
func (s *sOption) GetUserGradeList(ctx context.Context, req *v1.UserGradeListRequest) (*v1.UserGradeListResponse, error) {
	middleware.LogWithTrace(ctx, "info", "获取会员等级列表请求")

	// 默认站点ID为1
	siteId := int32(1)

	// 查询会员等级列表
	var userGrades []*entity.UserGrade
	err := dao.UserGrade.Ctx(ctx).Where(do.UserGrade{
		SiteId: siteId,
		Status: 1, // 只查询启用状态的等级
	}).OrderAsc("id").Scan(&userGrades)

	if err != nil {
		middleware.LogWithTrace(ctx, "error", "查询会员等级列表失败: %v", err)
		return nil, err
	}

	// 转换为响应格式
	grades := make([]*v1.UserGrade, 0, len(userGrades))
	for _, grade := range userGrades {
		grades = append(grades, &v1.UserGrade{
			Id:     int32(grade.Id),
			Name:   grade.Name,
			Sort:   int32(grade.Id), // 使用ID作为排序
			Status: int32(grade.Status),
		})
	}

	return &v1.UserGradeListResponse{
		Grades: grades,
	}, nil
}

// GetAdminRoleList 获取后台角色列表
func (s *sOption) GetAdminRoleList(ctx context.Context, req *v1.AdminRoleListRequest) (*v1.AdminRoleListResponse, error) {
	middleware.LogWithTrace(ctx, "info", "获取后台角色列表请求")

	// 默认站点ID为1
	siteId := int32(1)

	// 查询后台角色列表
	var adminRoles []*entity.AdminRole
	err := dao.AdminRole.Ctx(ctx).Where(do.AdminRole{
		SiteId: siteId,
		Status: 1, // 只查询启用状态的角色
	}).OrderAsc("id").Scan(&adminRoles)

	if err != nil {
		middleware.LogWithTrace(ctx, "error", "查询后台角色列表失败: %v", err)
		return nil, err
	}

	// 转换为响应格式
	roles := make([]*v1.AdminRole, 0, len(adminRoles))
	for _, role := range adminRoles {
		roles = append(roles, &v1.AdminRole{
			Id:          int32(role.Id),
			Name:        role.Name,
			Description: role.Name, // 使用名称作为描述，如果需要可以添加description字段
			Status:      int32(role.Status),
		})
	}

	middleware.LogWithTrace(ctx, "info", "获取后台角色列表成功 - 总数: %d", len(roles))

	return &v1.AdminRoleListResponse{
		Roles: roles,
	}, nil
}
