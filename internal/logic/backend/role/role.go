package role

import (
	"context"
	"fmt"
	"jh_app_service/api/backend/role/v1"
	"jh_app_service/internal/service/backend"
	"strconv"
	"strings"

	"jh_app_service/internal/dao"
	"jh_app_service/internal/middleware"
	"jh_app_service/internal/model/do"
	"jh_app_service/internal/model/entity"
)

type (
	sRole struct{}
)

func init() {
	backend.RegisterRole(&sRole{})
}

// GetRoleList 获取角色列表
func (s *sRole) GetRoleList(ctx context.Context, req *v1.GetRoleListReq) (*v1.GetRoleListRes, error) {
	middleware.LogWithTrace(ctx, "info", "获取角色列表请求 - SiteId: %d", req.SiteId)

	// 安全修复(A5)：站点ID 取自管理员 token 的权威站点ID，忽略请求体里的 site_id
	authSiteId, ok := middleware.GetAdminSiteIdFromContext(ctx)
	if !ok || authSiteId == 0 {
		middleware.LogWithTrace(ctx, "error", "未获取管理员站点ID")
		return nil, fmt.Errorf("未获取站点ID")
	}
	siteId := int32(authSiteId)

	// 查询角色列表
	var roles []*entity.AdminRole
	err := dao.AdminRole.Ctx(ctx).Where(do.AdminRole{
		SiteId: siteId,
		Status: 1, // 只查询启用的角色
	}).OrderAsc(dao.AdminRole.Columns().Id).Scan(&roles)

	if err != nil {
		middleware.LogWithTrace(ctx, "error", "查询角色列表失败: %v", err)
		return nil, fmt.Errorf("查询角色列表失败: %v", err)
	}

	// 构建响应数据
	var roleList []*v1.RoleInfo
	for _, role := range roles {
		roleInfo := &v1.RoleInfo{
			Id:          int32(role.Id),
			SiteId:      int32(role.SiteId),
			Name:        role.Name,
			Status:      int32(role.Status),
			Permissions: role.Permissions,
		}

		// 处理时间字段
		if role.CreatedAt != nil {
			roleInfo.CreatedAt = role.CreatedAt.Unix()
		}
		if role.UpdatedAt != nil {
			roleInfo.UpdatedAt = role.UpdatedAt.Unix()
		}

		roleList = append(roleList, roleInfo)
	}

	middleware.LogWithTrace(ctx, "info", "获取角色列表成功 - SiteId: %d, 角色数量: %d", siteId, len(roleList))

	return &v1.GetRoleListRes{
		Roles: roleList,
	}, nil
}

// CreateRole 创建角色
func (s *sRole) CreateRole(ctx context.Context, req *v1.CreateRoleReq) (*v1.CreateRoleRes, error) {
	middleware.LogWithTrace(ctx, "info", "创建角色请求 - SiteId: %d, Name: %s", req.SiteId, req.Name)

	// 参数验证
	if req.Name == "" {
		return &v1.CreateRoleRes{
			Success: false,
			Message: "角色名称不能为空",
		}, nil
	}

	if len(req.Name) < 2 || len(req.Name) > 50 {
		return &v1.CreateRoleRes{
			Success: false,
			Message: "角色名称长度必须在2-50个字符之间",
		}, nil
	}

	// 安全修复(A5)：站点ID 取自管理员 token 的权威站点ID，忽略请求体里的 site_id
	authSiteId, ok := middleware.GetAdminSiteIdFromContext(ctx)
	if !ok || authSiteId == 0 {
		return &v1.CreateRoleRes{
			Success: false,
			Message: "未获取站点ID",
		}, nil
	}
	siteId := int32(authSiteId)

	// 检查角色名称是否已存在
	count, err := dao.AdminRole.Ctx(ctx).Where(do.AdminRole{
		SiteId: siteId,
		Name:   req.Name,
	}).Count()
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "检查角色名称重复失败: %v", err)
		return &v1.CreateRoleRes{
			Success: false,
			Message: "系统错误，请稍后重试",
		}, nil
	}

	if count > 0 {
		return &v1.CreateRoleRes{
			Success: false,
			Message: "角色名称已存在",
		}, nil
	}

	// 创建角色
	_, err = dao.AdminRole.Ctx(ctx).Data(do.AdminRole{
		SiteId:      siteId,
		Name:        req.Name,
		Status:      1, // 默认启用
		Permissions: "",
	}).Insert()

	if err != nil {
		middleware.LogWithTrace(ctx, "error", "创建角色失败: %v", err)
		return &v1.CreateRoleRes{
			Success: false,
			Message: "创建角色失败",
		}, nil
	}

	middleware.LogWithTrace(ctx, "info", "创建角色成功 - SiteId: %d, Name: %s", siteId, req.Name)

	return &v1.CreateRoleRes{
		Success: true,
		Message: "创建成功",
	}, nil
}

// UpdateRole 更新角色
func (s *sRole) UpdateRole(ctx context.Context, req *v1.UpdateRoleReq) (*v1.UpdateRoleRes, error) {
	middleware.LogWithTrace(ctx, "info", "更新角色请求 - Id: %d, SiteId: %d, Name: %s", req.Id, req.SiteId, req.Name)

	// 参数验证
	if req.Id <= 0 {
		return &v1.UpdateRoleRes{
			Success: false,
			Message: "角色ID无效",
		}, nil
	}

	if req.Name == "" {
		return &v1.UpdateRoleRes{
			Success: false,
			Message: "角色名称不能为空",
		}, nil
	}

	if len(req.Name) < 2 || len(req.Name) > 50 {
		return &v1.UpdateRoleRes{
			Success: false,
			Message: "角色名称长度必须在2-50个字符之间",
		}, nil
	}

	// 安全修复(A5)：站点ID 取自管理员 token 的权威站点ID，忽略请求体里的 site_id
	authSiteId, ok := middleware.GetAdminSiteIdFromContext(ctx)
	if !ok || authSiteId == 0 {
		return &v1.UpdateRoleRes{
			Success: false,
			Message: "未获取站点ID",
		}, nil
	}
	siteId := int32(authSiteId)

	// 检查角色是否存在
	var role *entity.AdminRole
	err := dao.AdminRole.Ctx(ctx).Where(do.AdminRole{
		Id:     req.Id,
		SiteId: siteId,
	}).Scan(&role)
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "查询角色失败: %v", err)
		return &v1.UpdateRoleRes{
			Success: false,
			Message: "系统错误，请稍后重试",
		}, nil
	}

	if role == nil {
		return &v1.UpdateRoleRes{
			Success: false,
			Message: "角色不存在",
		}, nil
	}

	// 检查角色名称是否已被其他角色使用
	count, err := dao.AdminRole.Ctx(ctx).Where(do.AdminRole{
		SiteId: siteId,
		Name:   req.Name,
	}).WhereNot(dao.AdminRole.Columns().Id, req.Id).Count()
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "检查角色名称重复失败: %v", err)
		return &v1.UpdateRoleRes{
			Success: false,
			Message: "系统错误，请稍后重试",
		}, nil
	}

	if count > 0 {
		return &v1.UpdateRoleRes{
			Success: false,
			Message: "角色名称已存在",
		}, nil
	}

	// 更新角色
	_, err = dao.AdminRole.Ctx(ctx).Where(do.AdminRole{
		Id:     req.Id,
		SiteId: siteId,
	}).Data(do.AdminRole{
		Name: req.Name,
	}).Update()

	if err != nil {
		middleware.LogWithTrace(ctx, "error", "更新角色失败: %v", err)
		return &v1.UpdateRoleRes{
			Success: false,
			Message: "更新角色失败",
		}, nil
	}

	middleware.LogWithTrace(ctx, "info", "更新角色成功 - Id: %d, Name: %s", req.Id, req.Name)

	return &v1.UpdateRoleRes{
		Success: true,
		Message: "更新成功",
	}, nil
}

// DeleteRole 删除角色
func (s *sRole) DeleteRole(ctx context.Context, req *v1.DeleteRoleReq) (*v1.DeleteRoleRes, error) {
	middleware.LogWithTrace(ctx, "info", "删除角色请求 - Id: %d, SiteId: %d", req.Id, req.SiteId)

	// 参数验证
	if req.Id <= 0 {
		return &v1.DeleteRoleRes{
			Success: false,
			Message: "角色ID无效",
		}, nil
	}

	// 安全修复(A5)：站点ID 取自管理员 token 的权威站点ID，忽略请求体里的 site_id
	authSiteId, ok := middleware.GetAdminSiteIdFromContext(ctx)
	if !ok || authSiteId == 0 {
		return &v1.DeleteRoleRes{
			Success: false,
			Message: "未获取站点ID",
		}, nil
	}
	siteId := int32(authSiteId)

	// 检查角色是否存在
	var role *entity.AdminRole
	err := dao.AdminRole.Ctx(ctx).Where(do.AdminRole{
		Id:     req.Id,
		SiteId: siteId,
	}).Scan(&role)
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "查询角色失败: %v", err)
		return &v1.DeleteRoleRes{
			Success: false,
			Message: "系统错误，请稍后重试",
		}, nil
	}

	if role == nil {
		return &v1.DeleteRoleRes{
			Success: false,
			Message: "角色不存在",
		}, nil
	}

	// 删除角色
	_, err = dao.AdminRole.Ctx(ctx).Where(do.AdminRole{
		Id:     req.Id,
		SiteId: siteId,
	}).Delete()

	if err != nil {
		middleware.LogWithTrace(ctx, "error", "删除角色失败: %v", err)
		return &v1.DeleteRoleRes{
			Success: false,
			Message: "删除角色失败",
		}, nil
	}

	middleware.LogWithTrace(ctx, "info", "删除角色成功 - Id: %d, Name: %s", req.Id, role.Name)

	return &v1.DeleteRoleRes{
		Success: true,
		Message: "删除成功",
	}, nil
}

// GetPermissions 获取权限列表
func (s *sRole) GetPermissions(ctx context.Context, req *v1.GetPermissionsReq) (*v1.GetPermissionsRes, error) {
	middleware.LogWithTrace(ctx, "info", "获取权限列表请求 - SiteId: %d", req.SiteId)

	// 安全修复(A5)：站点ID 取自管理员 token 的权威站点ID，忽略请求体里的 site_id
	authSiteId, ok := middleware.GetAdminSiteIdFromContext(ctx)
	if !ok || authSiteId == 0 {
		middleware.LogWithTrace(ctx, "error", "未获取管理员站点ID")
		return nil, fmt.Errorf("未获取站点ID")
	}
	siteId := int32(authSiteId)

	// 获取角色列表及其权限
	var roles []*entity.AdminRole
	err := dao.AdminRole.Ctx(ctx).Where(do.AdminRole{
		SiteId: siteId,
		Status: 1, // 只查询启用的角色
	}).OrderAsc(dao.AdminRole.Columns().Id).Scan(&roles)

	if err != nil {
		middleware.LogWithTrace(ctx, "error", "查询角色列表失败: %v", err)
		return nil, fmt.Errorf("查询角色列表失败: %v", err)
	}

	// 构建角色权限列表
	var roleList []*v1.RolePermissionInfo
	for _, role := range roles {
		permissionList := []string{}
		if role.Permissions != "" {
			permissionList = strings.Split(role.Permissions, ",")
		}

		roleInfo := &v1.RolePermissionInfo{
			Id:             int32(role.Id),
			Name:           role.Name,
			PermissionList: permissionList,
		}
		roleList = append(roleList, roleInfo)
	}

	// 获取权限树列表
	var permissions []*entity.AdminPermission
	err = dao.AdminPermission.Ctx(ctx).Where(do.AdminPermission{
		Status: 1, // 只查询启用的权限
	}).OrderAsc(dao.AdminPermission.Columns().ParentId).
		OrderAsc(dao.AdminPermission.Columns().Sort).Scan(&permissions)

	if err != nil {
		middleware.LogWithTrace(ctx, "error", "查询权限列表失败: %v", err)
		return nil, fmt.Errorf("查询权限列表失败: %v", err)
	}

	// 构建权限树
	permissionTree := s.buildPermissionTree(permissions, 0)

	middleware.LogWithTrace(ctx, "info", "获取权限列表成功 - SiteId: %d, 角色数量: %d, 权限数量: %d",
		siteId, len(roleList), len(permissions))

	return &v1.GetPermissionsRes{
		PermissionList: permissionTree,
		RoleList:       roleList,
	}, nil
}

// SavePermission 保存权限
func (s *sRole) SavePermission(ctx context.Context, req *v1.SavePermissionReq) (*v1.SavePermissionRes, error) {
	middleware.LogWithTrace(ctx, "info", "保存权限请求 - Id: %d, SiteId: %d, PermissionList: %s",
		req.Id, req.SiteId, req.PermissionList)

	// 参数验证
	if req.Id <= 0 {
		return &v1.SavePermissionRes{
			Success: false,
			Message: "角色ID无效",
		}, nil
	}

	// 安全修复(A5)：站点ID 取自管理员 token 的权威站点ID，忽略请求体里的 site_id
	authSiteId, ok := middleware.GetAdminSiteIdFromContext(ctx)
	if !ok || authSiteId == 0 {
		return &v1.SavePermissionRes{
			Success: false,
			Message: "未获取站点ID",
		}, nil
	}
	siteId := int32(authSiteId)

	// 检查角色是否存在
	var role *entity.AdminRole
	err := dao.AdminRole.Ctx(ctx).Where(do.AdminRole{
		Id:     req.Id,
		SiteId: siteId,
	}).Scan(&role)
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "查询角色失败: %v", err)
		return &v1.SavePermissionRes{
			Success: false,
			Message: "系统错误，请稍后重试",
		}, nil
	}

	if role == nil {
		return &v1.SavePermissionRes{
			Success: false,
			Message: "角色不存在",
		}, nil
	}

	// 验证权限ID格式
	permissionList := strings.TrimSpace(req.PermissionList)
	if permissionList != "" {
		permissionIds := strings.Split(permissionList, ",")
		for _, idStr := range permissionIds {
			if _, err := strconv.Atoi(strings.TrimSpace(idStr)); err != nil {
				return &v1.SavePermissionRes{
					Success: false,
					Message: "权限ID格式错误",
				}, nil
			}
		}
	}

	// 更新角色权限
	_, err = dao.AdminRole.Ctx(ctx).Where(do.AdminRole{
		Id:     req.Id,
		SiteId: siteId,
	}).Data(do.AdminRole{
		Permissions: permissionList,
	}).Update()

	if err != nil {
		middleware.LogWithTrace(ctx, "error", "保存权限失败: %v", err)
		return &v1.SavePermissionRes{
			Success: false,
			Message: "保存权限失败",
		}, nil
	}

	middleware.LogWithTrace(ctx, "info", "保存权限成功 - Id: %d, Name: %s, PermissionList: %s",
		req.Id, role.Name, permissionList)

	return &v1.SavePermissionRes{
		Success: true,
		Message: "保存成功",
	}, nil
}

// buildPermissionTree 构建权限树
func (s *sRole) buildPermissionTree(permissions []*entity.AdminPermission, parentId int) []*v1.PermissionInfo {
	var tree []*v1.PermissionInfo

	for _, permission := range permissions {
		if permission.ParentId == parentId {
			permissionInfo := &v1.PermissionInfo{
				Id:          int32(permission.Id),
				ParentId:    int32(permission.ParentId),
				Type:        int32(permission.Type),
				Name:        permission.Name,
				BackendUrl:  permission.BackendUrl,
				FrontendUrl: permission.FrontendUrl,
				Open:        permission.Type == 1, // 菜单类型默认展开
				Checked:     false,
			}

			// 递归构建子权限
			permissionInfo.Children = s.buildPermissionTree(permissions, int(permission.Id))

			tree = append(tree, permissionInfo)
		}
	}

	return tree
}
