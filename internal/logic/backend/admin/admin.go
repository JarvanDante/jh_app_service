package admin

import (
	"context"
	"fmt"
	v1 "jh_app_service/api/backend/admin/v1"
	"jh_app_service/internal/service/backend"
	"jh_app_service/internal/util"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"golang.org/x/crypto/bcrypt"
	"jh_app_service/internal/dao"
	"jh_app_service/internal/middleware"
	"jh_app_service/internal/model/do"
	"jh_app_service/internal/model/entity"
	"jh_app_service/internal/tracing"
)

type (
	sAdmin struct{}
)

func init() {
	backend.RegisterAdmin(&sAdmin{})
}

// Login 管理员登录
func (s *sAdmin) Login(ctx context.Context, req *v1.LoginReq) (*v1.LoginRes, error) {
	// 创建Jaeger span
	ctx, span := tracing.StartSpan(ctx, "admin.Login", trace.WithAttributes(
		attribute.String("username", req.Username),
		attribute.String("method", "Login"),
	))
	defer span.End()

	// 获取站点ID (这里需要根据实际情况获取，可能从上下文或配置中获取)
	siteId, exists := middleware.GetSiteIdFromContext(ctx)
	if !exists {
		tracing.AddSpanEvent(span, "admin_id_not_found")
		middleware.LogWithTrace(ctx, "error", "无法获取管理员ID")
		return nil, fmt.Errorf("未登录或登录已过期")
	}
	tracing.SetSpanAttributes(span, attribute.Int("site_id", int(siteId)))

	// 数据库查询span
	ctx, dbSpan := tracing.StartSpan(ctx, "db.query.admin", trace.WithAttributes(
		attribute.String("db.operation", "select"),
		attribute.String("db.table", "admin"),
	))

	var admin *entity.Admin
	err := dao.Admin.Ctx(ctx).Where("username = ? AND site_id = ?", req.Username, siteId).Scan(&admin)

	dbSpan.End()

	if err != nil {
		tracing.SetSpanError(span, err)
		tracing.SetSpanError(dbSpan, err)
		middleware.LogWithTrace(ctx, "error", "数据库查询错误: %v", err)
		return nil, fmt.Errorf("数据库查询错误: %v", err)
	}

	if admin == nil {
		tracing.AddSpanEvent(span, "admin_not_found", attribute.String("username", req.Username))
		middleware.LogWithTrace(ctx, "warning", "未找到管理员记录 - 用户名: %s, 站点ID: %d", req.Username, siteId)
		return nil, fmt.Errorf("用户名或密码错误")
	}

	tracing.SetSpanAttributes(span,
		attribute.Int64("admin_id", int64(admin.Id)),
		attribute.Int("admin_status", admin.Status),
	)

	// 检查状态（状态检查放在找到记录之后）
	if admin.Status != 1 {
		tracing.AddSpanEvent(span, "admin_status_invalid",
			attribute.String("username", req.Username),
			attribute.Int("status", admin.Status),
		)
		middleware.LogWithTrace(ctx, "warning", "管理员状态异常 - 用户名: %s, 状态: %d", req.Username, admin.Status)
		return nil, fmt.Errorf("账号已被禁用")
	}

	// 验证密码span
	ctx, authSpan := tracing.StartSpan(ctx, "auth.password_verify")
	err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password))
	authSpan.End()

	if err != nil {
		tracing.SetSpanError(span, err)
		tracing.AddSpanEvent(span, "password_verification_failed", attribute.String("username", req.Username))
		middleware.LogWithTrace(ctx, "warning", "密码验证失败 - 用户名: %s, 错误: %v", req.Username, err)
		return nil, fmt.Errorf("用户名或密码错误")
	}

	middleware.SetAdminIdToContext(ctx, admin.Id)
	// 验证Google 2FA (如果开启)
	if admin.SwitchGoogle2Fa == 1 {
		tracing.AddSpanEvent(span, "google_2fa_required")
		if req.Code == "" {
			tracing.AddSpanEvent(span, "google_2fa_code_missing")
			return nil, fmt.Errorf("请输入动态验证码")
		}
		// 这里需要实现Google 2FA验证逻辑
		// valid := validateGoogle2FA(admin.Google2FaSecret, req.Code)
		// if !valid {
		//     return nil, fmt.Errorf("动态验证码错误")
		// }
	}

	// 生成JWT token span
	ctx, tokenSpan := tracing.StartSpan(ctx, "auth.generate_jwt_token")
	token, err := s.generateJWTToken(admin)
	tokenSpan.End()

	if err != nil {
		tracing.SetSpanError(span, err)
		tracing.SetSpanError(tokenSpan, err)
		return nil, fmt.Errorf("生成token失败: %v", err)
	}

	// 更新最后登录信息span
	ctx, updateSpan := tracing.StartSpan(ctx, "db.update.admin_login_info", trace.WithAttributes(
		attribute.String("db.operation", "update"),
		attribute.String("db.table", "admin"),
	))
	_, err = dao.Admin.Ctx(ctx).Where(do.Admin{Id: admin.Id}).Update(do.Admin{
		LastLoginIp:   s.getClientIP(ctx),
		LastLoginTime: gtime.Now(),
	})
	updateSpan.End()

	if err != nil {
		tracing.SetSpanError(updateSpan, err)
		middleware.LogWithTrace(ctx, "error", "更新登录信息失败: %v", err)
	}

	// 记录登录日志span
	ctx, logSpan := tracing.StartSpan(ctx, "db.insert.admin_log", trace.WithAttributes(
		attribute.String("db.operation", "insert"),
		attribute.String("db.table", "admin_log"),
	))
	err = s.addAdminLog(ctx, admin, "登录成功")
	logSpan.End()

	if err != nil {
		tracing.SetSpanError(logSpan, err)
		middleware.LogWithTrace(ctx, "error", "记录登录日志失败: %v", err)
	}

	// 获取socket地址 (从配置中获取)
	socketAddr := g.Cfg().MustGet(ctx, "workerman.host", "").String()
	if port := g.Cfg().MustGet(ctx, "workerman.port", "").String(); port != "" {
		socketAddr = socketAddr + ":" + port
	}

	res := &v1.LoginRes{
		Token:  token,
		Socket: socketAddr,
	}

	tracing.AddSpanEvent(span, "login_success",
		attribute.String("username", req.Username),
		attribute.Int64("admin_id", int64(admin.Id)),
	)
	tracing.SetSpanAttributes(span, attribute.Bool("success", true))

	middleware.LogWithTraceAndFields(ctx, "info", "登录成功", g.Map{
		"username": req.Username,
		"admin_id": admin.Id,
		"site_id":  siteId,
	})
	return res, nil
}

// RefreshToken 刷新令牌
func (s *sAdmin) RefreshToken(ctx context.Context, req *v1.RefreshTokenReq) (*v1.RefreshTokenRes, error) {
	// 创建Jaeger span
	ctx, span := tracing.StartSpan(ctx, "admin.RefreshToken", trace.WithAttributes(
		attribute.String("method", "RefreshToken"),
	))
	defer span.End()

	middleware.LogWithTrace(ctx, "info", "刷新token请求")

	// 从上下文中获取当前管理员ID
	adminId, exists := middleware.GetAdminIdFromContext(ctx)
	if !exists {
		tracing.AddSpanEvent(span, "admin_id_not_found")
		middleware.LogWithTrace(ctx, "error", "无法获取管理员ID")
		return nil, fmt.Errorf("未登录或登录已过期")
	}

	tracing.SetSpanAttributes(span, attribute.Int("admin_id", int(adminId)))
	middleware.LogWithTrace(ctx, "info", "获取到管理员ID: %d", adminId)

	// 查询管理员信息
	ctx, querySpan := tracing.StartSpan(ctx, "db.query.admin", trace.WithAttributes(
		attribute.String("db.operation", "select"),
		attribute.String("db.table", "admin"),
	))

	var admin *entity.Admin
	err := dao.Admin.Ctx(ctx).Where(do.Admin{Id: adminId}).Scan(&admin)
	querySpan.End()

	if err != nil {
		tracing.SetSpanError(span, err)
		tracing.SetSpanError(querySpan, err)
		middleware.LogWithTrace(ctx, "error", "查询管理员信息失败: %v", err)
		return nil, fmt.Errorf("查询管理员信息失败: %v", err)
	}

	if admin == nil {
		tracing.AddSpanEvent(span, "admin_not_found", attribute.Int("admin_id", int(adminId)))
		middleware.LogWithTrace(ctx, "warning", "管理员不存在 - ID: %d", adminId)
		return nil, fmt.Errorf("管理员不存在")
	}

	// 检查管理员状态
	if admin.Status != 1 {
		tracing.AddSpanEvent(span, "admin_status_invalid",
			attribute.Int("admin_id", int(adminId)),
			attribute.Int("status", admin.Status),
		)
		middleware.LogWithTrace(ctx, "warning", "管理员状态异常 - ID: %d, 状态: %d", adminId, admin.Status)
		return nil, fmt.Errorf("账号已被禁用")
	}

	// 生成新的JWT token
	ctx, tokenSpan := tracing.StartSpan(ctx, "auth.generate_jwt_token")
	token, err := s.generateJWTToken(admin)
	tokenSpan.End()

	if err != nil {
		tracing.SetSpanError(span, err)
		tracing.SetSpanError(tokenSpan, err)
		middleware.LogWithTrace(ctx, "error", "生成token失败: %v", err)
		return nil, fmt.Errorf("生成token失败: %v", err)
	}

	// 更新最后活跃时间（可选）
	ctx, updateSpan := tracing.StartSpan(ctx, "db.update.admin_last_active", trace.WithAttributes(
		attribute.String("db.operation", "update"),
		attribute.String("db.table", "admin"),
	))
	_, err = dao.Admin.Ctx(ctx).Where(do.Admin{Id: adminId}).Update(do.Admin{
		UpdatedAt: gtime.Now(),
	})
	updateSpan.End()

	if err != nil {
		tracing.SetSpanError(updateSpan, err)
		middleware.LogWithTrace(ctx, "error", "更新管理员活跃时间失败: %v", err)
		// 这个错误不影响token刷新，只记录日志
	}

	// 记录刷新日志
	ctx, logSpan := tracing.StartSpan(ctx, "db.insert.admin_log", trace.WithAttributes(
		attribute.String("db.operation", "insert"),
		attribute.String("db.table", "admin_log"),
	))
	err = s.addAdminLog(ctx, admin, "刷新token")
	logSpan.End()

	if err != nil {
		tracing.SetSpanError(logSpan, err)
		middleware.LogWithTrace(ctx, "error", "记录刷新日志失败: %v", err)
		// 这个错误不影响token刷新，只记录日志
	}

	res := &v1.RefreshTokenRes{
		Token: token,
	}

	tracing.AddSpanEvent(span, "token_refresh_success",
		attribute.Int("admin_id", int(adminId)),
		attribute.String("username", admin.Username),
	)
	tracing.SetSpanAttributes(span, attribute.Bool("success", true))

	middleware.LogWithTrace(ctx, "info", "刷新token成功 - 管理员ID: %d, 用户名: %s", adminId, admin.Username)
	return res, nil
}

// GetInfo 获取管理员信息
func (s *sAdmin) GetInfo(ctx context.Context, req *v1.GetInfoReq) (*v1.GetInfoRes, error) {
	// 创建Jaeger span
	ctx, span := tracing.StartSpan(ctx, "admin.GetInfo", trace.WithAttributes(
		attribute.String("method", "GetInfo"),
	))
	defer span.End()

	middleware.LogWithTrace(ctx, "info", "获取管理员信息请求")

	// 从上下文中获取当前管理员ID
	adminId, exists := middleware.GetAdminIdFromContext(ctx)
	if !exists {
		tracing.AddSpanEvent(span, "admin_id_not_found")
		middleware.LogWithTrace(ctx, "error", "无法获取管理员ID")
		return nil, fmt.Errorf("未登录或登录已过期")
	}

	tracing.SetSpanAttributes(span, attribute.Int("admin_id", int(adminId)))
	middleware.LogWithTrace(ctx, "info", "获取到管理员ID: %d", adminId)

	// 查询管理员信息
	ctx, querySpan := tracing.StartSpan(ctx, "db.query.admin", trace.WithAttributes(
		attribute.String("db.operation", "select"),
		attribute.String("db.table", "admin"),
	))

	var admin *entity.Admin
	err := dao.Admin.Ctx(ctx).Where(do.Admin{Id: adminId}).Scan(&admin)
	querySpan.End()

	if err != nil {
		tracing.SetSpanError(span, err)
		tracing.SetSpanError(querySpan, err)
		middleware.LogWithTrace(ctx, "error", "查询管理员信息失败: %v", err)
		return nil, fmt.Errorf("查询管理员信息失败: %v", err)
	}

	if admin == nil {
		tracing.AddSpanEvent(span, "admin_not_found", attribute.Int("admin_id", int(adminId)))
		middleware.LogWithTrace(ctx, "warning", "管理员不存在 - ID: %d", adminId)
		return nil, fmt.Errorf("管理员不存在")
	}

	// 检查管理员状态
	if admin.Status != 1 {
		tracing.AddSpanEvent(span, "admin_status_invalid",
			attribute.Int("admin_id", int(adminId)),
			attribute.Int("status", admin.Status),
		)
		middleware.LogWithTrace(ctx, "warning", "管理员状态异常 - ID: %d, 状态: %d", adminId, admin.Status)
		return nil, fmt.Errorf("账号已被禁用")
	}

	// 查询管理员角色信息
	ctx, roleSpan := tracing.StartSpan(ctx, "db.query.admin_role", trace.WithAttributes(
		attribute.String("db.operation", "select"),
		attribute.String("db.table", "admin_role"),
	))

	var role *entity.AdminRole
	err = dao.AdminRole.Ctx(ctx).Where(do.AdminRole{Id: uint(admin.AdminRoleId)}).Scan(&role)
	roleSpan.End()

	if err != nil {
		tracing.SetSpanError(roleSpan, err)
		middleware.LogWithTrace(ctx, "error", "查询管理员角色信息失败: %v", err)
		// 角色查询失败不影响基本信息返回，使用默认值
	}

	// 构建角色列表
	roles := []string{}
	if role != nil {
		roles = append(roles, role.Name)
	} else {
		roles = append(roles, "未知角色")
	}

	// 查询菜单权限 - 这里需要根据实际的权限系统实现
	// 暂时返回一个示例菜单结构
	menus := s.buildMenus(ctx, admin)

	// 构建响应
	res := &v1.GetInfoRes{
		Roles:        roles,
		Name:         admin.Nickname,                                                    // 使用昵称作为显示名称
		Avatar:       "https://wpimg.wallstcn.com/577965b9-bb9e-4e02-9f0c-095b41417191", // 默认头像
		Introduction: fmt.Sprintf("管理员 %s", admin.Username),
		Menus:        menus,
	}

	tracing.AddSpanEvent(span, "get_info_success",
		attribute.Int("admin_id", int(adminId)),
		attribute.String("username", admin.Username),
		attribute.Int("menu_count", len(menus)),
	)
	tracing.SetSpanAttributes(span, attribute.Bool("success", true))

	middleware.LogWithTrace(ctx, "info", "获取管理员信息成功 - ID: %d, 用户名: %s, 菜单数量: %d",
		adminId, admin.Username, len(menus))

	return res, nil
}

// buildMenus 构建菜单结构 - 这里是一个示例实现
func (s *sAdmin) buildMenus(ctx context.Context, admin *entity.Admin) []*v1.MenuInfo {
	// 这里应该根据管理员的角色权限查询实际的菜单
	// 暂时返回一个示例菜单结构，参考 go_service 项目的返回格式

	menus := []*v1.MenuInfo{
		{
			Id:   1,
			Name: "系统管理",
			Path: "system",
			Type: 1,
			Sort: 0,
			Children: []*v1.MenuInfo{
				{
					Id:   2,
					Name: "管理员管理",
					Path: "system/admin",
					Type: 1,
					Sort: 0,
					Children: []*v1.MenuInfo{
						{
							Id:   3,
							Name: "管理员列表",
							Path: "system/admin/list",
							Type: 1,
							Sort: 0,
						},
					},
				},
				{
					Id:   4,
					Name: "角色管理",
					Path: "system/role",
					Type: 1,
					Sort: 1,
				},
			},
		},
		{
			Id:   5,
			Name: "用户管理",
			Path: "user",
			Type: 1,
			Sort: 1,
			Children: []*v1.MenuInfo{
				{
					Id:   6,
					Name: "用户列表",
					Path: "user/list",
					Type: 1,
					Sort: 0,
				},
			},
		},
	}

	return menus
}

// CreateAdmin 创建管理员
func (s *sAdmin) CreateAdmin(ctx context.Context, req *v1.CreateAdminReq) (*v1.CreateAdminRes, error) {
	// 创建Jaeger span
	ctx, span := tracing.StartSpan(ctx, "admin.CreateAdmin", trace.WithAttributes(
		attribute.String("username", req.Username),
		attribute.String("nickname", req.Nickname),
		attribute.String("method", "CreateAdmin"),
	))
	defer span.End()

	// 获取站点ID (这里需要根据实际情况获取，可能从上下文或配置中获取)
	siteId := 1 // 临时硬编码，实际应该从请求中获取
	tracing.SetSpanAttributes(span, attribute.Int("site_id", siteId))

	middleware.LogWithTrace(ctx, "info", "创建管理员请求 - 用户名: %s, 昵称: %s", req.Username, req.Nickname)

	// 手动验证参数 - 不依赖 protobuf 验证标签
	if err := s.validateCreateAdminRequest(req); err != nil {
		tracing.AddSpanEvent(span, "validation_failed", attribute.String("reason", err.Error()))
		middleware.LogWithTrace(ctx, "error", "创建管理员参数验证失败: %v", err)
		return nil, err
	}

	// 检查用户名是否已存在
	middleware.LogWithTrace(ctx, "info", "检查用户名是否存在 - 用户名: %s, 站点ID: %d", req.Username, siteId)

	ctx, checkSpan := tracing.StartSpan(ctx, "db.query.check_username_exists", trace.WithAttributes(
		attribute.String("db.operation", "select"),
		attribute.String("db.table", "admin"),
		attribute.String("username", req.Username),
	))

	var existingAdmin *entity.Admin
	err := dao.Admin.Ctx(ctx).Where(do.Admin{
		Username: req.Username,
		SiteId:   siteId,
	}).Scan(&existingAdmin)

	checkSpan.End()

	if err != nil {
		tracing.SetSpanError(span, err)
		tracing.SetSpanError(checkSpan, err)
		middleware.LogWithTrace(ctx, "error", "检查用户名存在性时数据库查询失败: %v", err)
		return nil, fmt.Errorf("数据库查询错误: %v", err)
	}

	if existingAdmin != nil {
		tracing.AddSpanEvent(span, "username_already_exists", attribute.String("username", req.Username))
		middleware.LogWithTrace(ctx, "warning", "用户名已存在 - 用户名: %s, 站点ID: %d", req.Username, siteId)
		return nil, fmt.Errorf("用户名已经被使用")
	}

	// 加密密码
	middleware.LogWithTrace(ctx, "info", "开始加密密码 - 用户名: %s", req.Username)

	ctx, hashSpan := tracing.StartSpan(ctx, "auth.hash_password")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	hashSpan.End()

	if err != nil {
		tracing.SetSpanError(span, err)
		tracing.SetSpanError(hashSpan, err)
		middleware.LogWithTrace(ctx, "error", "密码加密失败 - 用户名: %s, 错误: %v", req.Username, err)
		return nil, fmt.Errorf("密码加密失败: %v", err)
	}

	// 创建管理员
	middleware.LogWithTrace(ctx, "info", "开始创建管理员 - 用户名: %s, 昵称: %s, 角色: %d, 状态: %d",
		req.Username, req.Nickname, req.Role, req.Status)

	// 确保状态值正确 (如果为0则设为1)
	status := int(req.Status)
	if status == 0 {
		status = 1
	}

	tracing.SetSpanAttributes(span,
		attribute.Int("role", int(req.Role)),
		attribute.Int("status", status),
	)

	ctx, insertSpan := tracing.StartSpan(ctx, "db.insert.admin", trace.WithAttributes(
		attribute.String("db.operation", "insert"),
		attribute.String("db.table", "admin"),
	))

	_, err = dao.Admin.Ctx(ctx).Insert(do.Admin{
		SiteId:      siteId,
		Username:    req.Username,
		Nickname:    req.Nickname,
		Password:    string(hashedPassword),
		AdminRoleId: int(req.Role),
		Status:      status,
		CreatedAt:   gtime.Now(),
		UpdatedAt:   gtime.Now(),
	})

	insertSpan.End()

	if err != nil {
		tracing.SetSpanError(span, err)
		tracing.SetSpanError(insertSpan, err)
		middleware.LogWithTrace(ctx, "error", "创建管理员数据库操作失败 - 用户名: %s, 错误: %v", req.Username, err)
		return nil, fmt.Errorf("创建管理员失败: %v", err)
	}

	tracing.AddSpanEvent(span, "admin_created_successfully",
		attribute.String("username", req.Username),
		attribute.String("nickname", req.Nickname),
	)
	tracing.SetSpanAttributes(span, attribute.Bool("success", true))

	middleware.LogWithTrace(ctx, "info", "创建管理员成功 - 用户名: %s, 实际状态: %d", req.Username, status)

	// 记录操作日志
	// 这里需要获取当前操作的管理员信息，暂时跳过
	// err = s.addAdminLog(ctx, currentAdmin, "添加员工："+req.Username)

	res := &v1.CreateAdminRes{}
	return res, nil
}

// validateCreateAdminRequest 验证创建管理员请求参数
func (s *sAdmin) validateCreateAdminRequest(req *v1.CreateAdminReq) error {
	// 验证必填字段
	if req.Username == "" {
		return fmt.Errorf("用户名不能为空")
	}
	if req.Password == "" {
		return fmt.Errorf("密码不能为空")
	}
	if req.Nickname == "" {
		return fmt.Errorf("昵称不能为空")
	}

	// 验证用户名长度和格式
	if len(req.Username) < 4 || len(req.Username) > 12 {
		return fmt.Errorf("用户名长度必须在4-12个字符之间")
	}

	// 验证密码长度
	if len(req.Password) < 6 || len(req.Password) > 20 {
		return fmt.Errorf("密码长度必须在6-20个字符之间")
	}

	// 验证昵称长度
	if len(req.Nickname) < 2 || len(req.Nickname) > 20 {
		return fmt.Errorf("昵称长度必须在2-20个字符之间")
	}

	// 验证角色
	if req.Role <= 0 {
		return fmt.Errorf("请选择有效的角色")
	}

	// 验证状态
	if req.Status < 0 || req.Status > 1 {
		return fmt.Errorf("状态值无效")
	}

	return nil
}

// GetAdminList 获取管理员列表
func (s *sAdmin) GetAdminList(ctx context.Context, req *v1.GetAdminListReq) (*v1.GetAdminListRes, error) {
	// 创建Jaeger span
	ctx, span := tracing.StartSpan(ctx, "admin.GetAdminList", trace.WithAttributes(
		attribute.String("method", "GetAdminList"),
	))
	defer span.End()

	// 获取站点ID
	siteId := 1 // 临时硬编码，实际应该从请求中获取
	tracing.SetSpanAttributes(span, attribute.Int("site_id", siteId))

	middleware.LogWithTrace(ctx, "info", "获取管理员列表请求 - 用户名: %s, 状态: %d, 页码: %d, 大小: %d",
		req.Username, req.Status, req.Page, req.Size)

	// 设置默认分页参数
	page := int(req.Page)
	size := int(req.Size)
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}

	// 构建查询条件
	query := dao.Admin.Ctx(ctx).Where("site_id = ?", siteId)

	if req.Username != "" {
		query = query.Where("username LIKE ?", "%"+req.Username+"%")
	}

	if req.Status > 0 {
		query = query.Where("status = ?", req.Status)
	}

	// 查询总数
	ctx, countSpan := tracing.StartSpan(ctx, "db.query.admin_count", trace.WithAttributes(
		attribute.String("db.operation", "count"),
		attribute.String("db.table", "admin"),
	))

	total, err := query.Count()
	countSpan.End()

	if err != nil {
		tracing.SetSpanError(span, err)
		tracing.SetSpanError(countSpan, err)
		middleware.LogWithTrace(ctx, "error", "查询管理员总数失败: %v", err)
		return nil, fmt.Errorf("查询管理员总数失败: %v", err)
	}

	// 查询列表数据
	ctx, listSpan := tracing.StartSpan(ctx, "db.query.admin_list", trace.WithAttributes(
		attribute.String("db.operation", "select"),
		attribute.String("db.table", "admin"),
	))

	var admins []*entity.Admin
	err = query.Page(page, size).OrderDesc("id").Scan(&admins)
	listSpan.End()

	if err != nil {
		tracing.SetSpanError(span, err)
		tracing.SetSpanError(listSpan, err)
		middleware.LogWithTrace(ctx, "error", "查询管理员列表失败: %v", err)
		return nil, fmt.Errorf("查询管理员列表失败: %v", err)
	}

	// 转换为响应格式
	var adminList []*v1.AdminInfo

	// 预先获取所有角色信息，避免N+1查询问题
	var roleIds []interface{}
	for _, admin := range admins {
		if admin.AdminRoleId > 0 {
			roleIds = append(roleIds, admin.AdminRoleId)
		}
	}

	roleMap := make(map[int]string)
	if len(roleIds) > 0 {
		var roles []*entity.AdminRole
		err := dao.AdminRole.Ctx(ctx).WhereIn(dao.AdminRole.Columns().Id, roleIds).Scan(&roles)
		if err == nil {
			for _, role := range roles {
				roleMap[int(role.Id)] = role.Name
			}
		}
	}

	for _, admin := range admins {
		adminInfo := &v1.AdminInfo{
			Id:            int32(admin.Id),
			Username:      admin.Username,
			Nickname:      admin.Nickname,
			Role:          int32(admin.AdminRoleId),
			Status:        int32(admin.Status),
			LastLoginIp:   admin.LastLoginIp, // 直接赋值，即使为空字符串
			LastLoginTime: admin.LastLoginTime.Format("2006-01-02 15:04:05"),
			CreatedAt:     admin.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		// 设置角色名称
		if roleName, exists := roleMap[admin.AdminRoleId]; exists {
			adminInfo.RoleName = roleName
		} else {
			adminInfo.RoleName = ""
		}

		adminList = append(adminList, adminInfo)
	}

	// TODO: 获取操作二次验证权限
	// google2faAccess := AdminRole::validRolePermission($this->admin->admin_role_id, 'bind-google2fa');

	res := &v1.GetAdminListRes{
		List:            adminList,
		Total:           int32(total),
		Page:            int32(page),
		Size:            int32(size),
		Google2FaAccess: false, // 暂时设为false，需要实现权限检查
	}

	tracing.SetSpanAttributes(span, attribute.Bool("success", true))
	middleware.LogWithTrace(ctx, "info", "获取管理员列表成功 - 总数: %d, 当前页: %d", total, page)

	return res, nil
}

// UpdateAdmin 更新管理员
func (s *sAdmin) UpdateAdmin(ctx context.Context, req *v1.UpdateAdminReq) (*v1.UpdateAdminRes, error) {
	// 创建Jaeger span
	ctx, span := tracing.StartSpan(ctx, "admin.UpdateAdmin", trace.WithAttributes(
		attribute.Int("admin_id", int(req.Id)),
		attribute.String("method", "UpdateAdmin"),
	))
	defer span.End()

	// 获取站点ID
	siteId := 1 // 临时硬编码，实际应该从请求中获取
	tracing.SetSpanAttributes(span, attribute.Int("site_id", siteId))

	middleware.LogWithTrace(ctx, "info", "更新管理员请求 - ID: %d", req.Id)

	if req.Id <= 0 {
		tracing.AddSpanEvent(span, "validation_failed", attribute.String("reason", "invalid_admin_id"))
		middleware.LogWithTrace(ctx, "error", "更新管理员参数验证失败 - 无效的管理员ID: %d", req.Id)
		return nil, fmt.Errorf("无效的管理员ID")
	}

	// 验证参数
	if req.Password != "" && (len(req.Password) < 6 || len(req.Password) > 20) {
		tracing.AddSpanEvent(span, "validation_failed", attribute.String("reason", "password_length_invalid"))
		middleware.LogWithTrace(ctx, "error", "更新管理员参数验证失败 - 密码长度不符合要求")
		return nil, fmt.Errorf("密码长度必须在6-20个字符之间")
	}

	if req.Nickname != "" && (len(req.Nickname) < 2 || len(req.Nickname) > 20) {
		tracing.AddSpanEvent(span, "validation_failed", attribute.String("reason", "nickname_length_invalid"))
		middleware.LogWithTrace(ctx, "error", "更新管理员参数验证失败 - 昵称长度不符合要求")
		return nil, fmt.Errorf("昵称长度必须在2-20个字符之间")
	}

	// 检查管理员是否存在
	ctx, checkSpan := tracing.StartSpan(ctx, "db.query.check_admin_exists", trace.WithAttributes(
		attribute.String("db.operation", "select"),
		attribute.String("db.table", "admin"),
	))

	var admin *entity.Admin
	err := dao.Admin.Ctx(ctx).Where(do.Admin{
		Id:     int(req.Id),
		SiteId: siteId,
	}).Scan(&admin)

	checkSpan.End()

	if err != nil {
		tracing.SetSpanError(span, err)
		tracing.SetSpanError(checkSpan, err)
		middleware.LogWithTrace(ctx, "error", "查询管理员信息失败: %v", err)
		return nil, fmt.Errorf("查询管理员信息失败: %v", err)
	}

	if admin == nil {
		tracing.AddSpanEvent(span, "admin_not_found", attribute.Int("admin_id", int(req.Id)))
		middleware.LogWithTrace(ctx, "warning", "管理员不存在 - ID: %d", req.Id)
		return nil, fmt.Errorf("管理员不存在")
	}

	// 构建更新数据
	updateData := do.Admin{
		UpdatedAt: gtime.Now(),
	}

	if req.Password != "" {
		// 加密密码
		ctx, hashSpan := tracing.StartSpan(ctx, "auth.hash_password")
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		hashSpan.End()

		if err != nil {
			tracing.SetSpanError(span, err)
			tracing.SetSpanError(hashSpan, err)
			middleware.LogWithTrace(ctx, "error", "密码加密失败: %v", err)
			return nil, fmt.Errorf("密码加密失败: %v", err)
		}
		updateData.Password = string(hashedPassword)
	}

	if req.Nickname != "" {
		updateData.Nickname = req.Nickname
	}

	if req.Role > 0 {
		updateData.AdminRoleId = int(req.Role)
	}

	if req.Status >= 0 {
		updateData.Status = int(req.Status)
	}

	// 更新管理员
	ctx, updateSpan := tracing.StartSpan(ctx, "db.update.admin", trace.WithAttributes(
		attribute.String("db.operation", "update"),
		attribute.String("db.table", "admin"),
	))

	_, err = dao.Admin.Ctx(ctx).Where(do.Admin{Id: int(req.Id)}).Update(updateData)
	updateSpan.End()

	if err != nil {
		tracing.SetSpanError(span, err)
		tracing.SetSpanError(updateSpan, err)
		middleware.LogWithTrace(ctx, "error", "更新管理员失败: %v", err)
		return nil, fmt.Errorf("更新管理员失败: %v", err)
	}

	tracing.AddSpanEvent(span, "admin_updated_successfully", attribute.Int("admin_id", int(req.Id)))
	tracing.SetSpanAttributes(span, attribute.Bool("success", true))

	middleware.LogWithTrace(ctx, "info", "更新管理员成功 - ID: %d, 用户名: %s", req.Id, admin.Username)

	// 记录操作日志
	err = s.addAdminLog(ctx, admin, "编辑员工："+admin.Username)
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "记录操作日志失败: %v", err)
	}

	res := &v1.UpdateAdminRes{}
	return res, nil
}

// DeleteAdmin 删除管理员
func (s *sAdmin) DeleteAdmin(ctx context.Context, req *v1.DeleteAdminReq) (*v1.DeleteAdminRes, error) {
	// 创建Jaeger span
	ctx, span := tracing.StartSpan(ctx, "admin.DeleteAdmin", trace.WithAttributes(
		attribute.Int("admin_id", int(req.Id)),
		attribute.String("method", "DeleteAdmin"),
	))
	defer span.End()

	// 获取站点ID
	siteId := 1 // 临时硬编码，实际应该从请求中获取
	tracing.SetSpanAttributes(span, attribute.Int("site_id", siteId))

	middleware.LogWithTrace(ctx, "info", "删除管理员请求 - ID: %d", req.Id)

	if req.Id <= 0 {
		tracing.AddSpanEvent(span, "validation_failed", attribute.String("reason", "invalid_admin_id"))
		middleware.LogWithTrace(ctx, "error", "删除管理员参数验证失败 - 无效的管理员ID: %d", req.Id)
		return nil, fmt.Errorf("无效的管理员ID")
	}

	// 获取当前操作的管理员ID，防止自己删除自己
	currentAdminId, exists := middleware.GetAdminIdFromContext(ctx)
	if exists && currentAdminId == uint(req.Id) {
		tracing.AddSpanEvent(span, "validation_failed", attribute.String("reason", "cannot_delete_self"))
		middleware.LogWithTrace(ctx, "warning", "不能删除自己 - 当前管理员ID: %d, 要删除的ID: %d", currentAdminId, req.Id)
		return nil, fmt.Errorf("不能删除自己")
	}

	// 检查管理员是否存在
	ctx, checkSpan := tracing.StartSpan(ctx, "db.query.check_admin_exists", trace.WithAttributes(
		attribute.String("db.operation", "select"),
		attribute.String("db.table", "admin"),
	))

	var admin *entity.Admin
	err := dao.Admin.Ctx(ctx).Where(do.Admin{
		Id:     int(req.Id),
		SiteId: siteId,
	}).Scan(&admin)

	checkSpan.End()

	if err != nil {
		tracing.SetSpanError(span, err)
		tracing.SetSpanError(checkSpan, err)
		middleware.LogWithTrace(ctx, "error", "查询管理员信息失败: %v", err)
		return nil, fmt.Errorf("查询管理员信息失败: %v", err)
	}

	if admin == nil {
		tracing.AddSpanEvent(span, "admin_not_found", attribute.Int("admin_id", int(req.Id)))
		middleware.LogWithTrace(ctx, "warning", "管理员不存在 - ID: %d", req.Id)
		return nil, fmt.Errorf("管理员不存在")
	}

	// 软删除管理员 (设置 delete_at 字段)
	ctx, deleteSpan := tracing.StartSpan(ctx, "db.update.admin_soft_delete", trace.WithAttributes(
		attribute.String("db.operation", "update"),
		attribute.String("db.table", "admin"),
	))

	_, err = dao.Admin.Ctx(ctx).Where(do.Admin{Id: int(req.Id)}).Update(do.Admin{
		DeleteAt:  gtime.Now(),
		UpdatedAt: gtime.Now(),
	})

	deleteSpan.End()

	if err != nil {
		tracing.SetSpanError(span, err)
		tracing.SetSpanError(deleteSpan, err)
		middleware.LogWithTrace(ctx, "error", "删除管理员失败: %v", err)
		return nil, fmt.Errorf("删除管理员失败: %v", err)
	}

	tracing.AddSpanEvent(span, "admin_deleted_successfully", attribute.Int("admin_id", int(req.Id)))
	tracing.SetSpanAttributes(span, attribute.Bool("success", true))

	middleware.LogWithTrace(ctx, "info", "删除管理员成功 - ID: %d, 用户名: %s", req.Id, admin.Username)

	// 记录操作日志
	err = s.addAdminLog(ctx, admin, "删除员工："+admin.Username)
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "记录操作日志失败: %v", err)
	}

	res := &v1.DeleteAdminRes{}
	return res, nil
}

// Logout 退出登录
func (s *sAdmin) Logout(ctx context.Context, req *v1.LogoutReq) (*v1.LogoutRes, error) {
	middleware.LogWithTrace(ctx, "info", "管理员退出登录请求")

	// 从 gRPC metadata 中获取当前管理员信息
	adminId, exists := middleware.GetAdminIdFromContext(ctx)
	middleware.LogWithTrace(ctx, "info", "从上下文中获取管理员ID: %d, exists: %v", adminId, exists)

	// 同时尝试直接从 gRPC metadata 获取，用于调试
	if directAdminId, directExists := middleware.GetAdminIdFromGRPCMetadata(ctx); directExists {
		middleware.LogWithTrace(ctx, "info", "直接从 gRPC metadata 获取管理员ID: %d", directAdminId)
	} else {
		middleware.LogWithTrace(ctx, "warn", "无法直接从 gRPC metadata 获取管理员ID")
	}

	if !exists {
		middleware.LogWithTrace(ctx, "error", "无法获取管理员ID")
		return &v1.LogoutRes{
			Success: false,
			Message: "未登录或登录已过期",
		}, nil
	}

	// 获取管理员信息用于日志记录
	var admin *entity.Admin
	err := dao.Admin.Ctx(ctx).Where(do.Admin{Id: adminId}).Scan(&admin)
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "查询管理员信息失败: %v", err)
	}

	// 获取客户端IP
	clientIP := middleware.GetClientIPFromContext(ctx)
	middleware.LogWithTrace(ctx, "info", "客户端IP: %s", clientIP)

	// 在实际的JWT实现中，这里可能需要将token加入黑名单
	// 但由于JWT是无状态的，通常在客户端删除token即可
	// 这里我们只记录退出日志

	if admin != nil {
		// 记录退出日志
		_, err = dao.AdminLog.Ctx(ctx).Insert(do.AdminLog{
			SiteId:        admin.SiteId,
			AdminId:       int(admin.Id),
			AdminUsername: admin.Username,
			Ip:            clientIP,
			Remark:        "管理员退出登录",
		})
		if err != nil {
			middleware.LogWithTrace(ctx, "error", "记录退出日志失败: %v", err)
		}

		middleware.LogWithTrace(ctx, "info", "管理员退出登录成功 - Username: %s, IP: %s", admin.Username, clientIP)
	}

	return &v1.LogoutRes{
		Success: true,
		Message: "退出成功",
	}, nil
}

// ChangePassword 修改密码
func (s *sAdmin) ChangePassword(ctx context.Context, req *v1.ChangePasswordReq) (*v1.ChangePasswordRes, error) {
	middleware.LogWithTrace(ctx, "info", "管理员修改密码请求")

	// 参数验证
	if req.OldPassword == "" {
		return &v1.ChangePasswordRes{
			Success: false,
			Message: "请输入旧密码",
		}, nil
	}

	if req.NewPassword == "" {
		return &v1.ChangePasswordRes{
			Success: false,
			Message: "请输入新密码",
		}, nil
	}

	if len(req.NewPassword) < 6 || len(req.NewPassword) > 20 {
		return &v1.ChangePasswordRes{
			Success: false,
			Message: "新密码长度必须在6-20个字符之间",
		}, nil
	}

	if req.OldPassword == req.NewPassword {
		return &v1.ChangePasswordRes{
			Success: false,
			Message: "新密码不能与旧密码相同",
		}, nil
	}

	// 从上下文中获取当前管理员信息
	adminId, exists := middleware.GetAdminIdFromContext(ctx)
	if !exists {
		middleware.LogWithTrace(ctx, "error", "无法获取管理员ID")
		return &v1.ChangePasswordRes{
			Success: false,
			Message: "未登录或登录已过期",
		}, nil
	}

	// 获取管理员信息
	var admin *entity.Admin
	err := dao.Admin.Ctx(ctx).Where(do.Admin{Id: adminId}).Scan(&admin)
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "查询管理员信息失败: %v", err)
		return &v1.ChangePasswordRes{
			Success: false,
			Message: "系统错误，请稍后重试",
		}, nil
	}

	if admin == nil {
		return &v1.ChangePasswordRes{
			Success: false,
			Message: "管理员不存在",
		}, nil
	}

	// 验证旧密码
	err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.OldPassword))
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "旧密码验证失败: %v", err)
		return &v1.ChangePasswordRes{
			Success: false,
			Message: "旧密码错误",
		}, nil
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "密码加密失败: %v", err)
		return &v1.ChangePasswordRes{
			Success: false,
			Message: "系统错误，请稍后重试",
		}, nil
	}

	// 更新密码
	_, err = dao.Admin.Ctx(ctx).Where(do.Admin{Id: adminId}).Update(do.Admin{
		Password: string(hashedPassword),
	})
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "更新密码失败: %v", err)
		return &v1.ChangePasswordRes{
			Success: false,
			Message: "修改密码失败",
		}, nil
	}

	// 记录操作日志
	_, err = dao.AdminLog.Ctx(ctx).Insert(do.AdminLog{
		SiteId:        admin.SiteId,
		AdminId:       int(admin.Id),
		AdminUsername: admin.Username,
		Ip:            middleware.GetClientIPFromContext(ctx),
		Remark:        "管理员修改密码",
	})
	if err != nil {
		middleware.LogWithTrace(ctx, "error", "记录操作日志失败: %v", err)
	}

	middleware.LogWithTrace(ctx, "info", "管理员修改密码成功 - Username: %s", admin.Username)

	return &v1.ChangePasswordRes{
		Success: true,
		Message: "修改密码成功",
	}, nil
}

// 辅助方法

// generateJWTToken 生成JWT token
func (s *sAdmin) generateJWTToken(admin *entity.Admin) (string, error) {
	// 从配置文件获取JWT密钥
	jwtSecret := g.Cfg().MustGet(context.Background(), "jwt.secret").String()
	if jwtSecret == "" {
		return "", fmt.Errorf("JWT secret not configured")
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		//"user_id":  admin.Id, // 使用 user_id 字段名，与 Gateway 的 Claims 结构体匹配
		"user_id":  0,        // 后台使用 user_id =0
		"admin_id": admin.Id, // 保留 admin_id 用于兼容
		"username": admin.Username,
		"site_id":  admin.SiteId,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // 24小时过期
		"iat":      time.Now().Unix(),
	})

	// 签名token
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// getClientIP 获取客户端IP
func (s *sAdmin) getClientIP(ctx context.Context) string {
	// 从上下文中获取HTTP请求
	if r := g.RequestFromCtx(ctx); r != nil {
		return r.GetClientIp()
	}
	return "127.0.0.1"
}

// addAdminLog 添加管理员日志
func (s *sAdmin) addAdminLog(ctx context.Context, admin *entity.Admin, message string) error {
	_, err := dao.AdminLog.Ctx(ctx).Insert(do.AdminLog{
		SiteId:        admin.SiteId,
		AdminId:       int(admin.Id),
		AdminUsername: admin.Username,
		Ip:            s.getClientIP(ctx),
		Remark:        message,
		CreatedAt:     gtime.Now(),
	})
	return err
}

// GetAdminLogs 获取管理员日志列表
func (s *sAdmin) GetAdminLogs(ctx context.Context, req *v1.GetAdminLogsReq) (*v1.GetAdminLogsRes, error) {
	// 参数验证
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	// 创建Jaeger span
	ctx, span := tracing.StartSpan(ctx, "admin.GetAdminLogs", trace.WithAttributes(
		attribute.String("method", "GetAdminLogs"),
		attribute.String("username", req.Username),
		attribute.Int("page", int(req.Page)),
		attribute.Int("size", int(req.Size)),
	))
	defer span.End()

	// 获取站点ID (从上下文或配置中获取)
	siteId := 1 // 临时硬编码，实际应该从请求中获取
	tracing.SetSpanAttributes(span, attribute.Int("site_id", siteId))

	// 设置默认分页参数
	page := req.Page
	size := req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 50
	}

	// 构建查询条件
	query := dao.AdminLog.Ctx(ctx).Where(do.AdminLog{
		SiteId: siteId,
	})

	// 用户名筛选
	if req.Username != "" {
		query = query.Where(do.AdminLog{
			AdminUsername: req.Username,
		})
	}

	// 时间范围筛选
	if req.Start != "" && req.End != "" {
		query = query.WhereBetween("created_at", req.Start, req.End)
	}

	// 数据库查询span - 获取总数
	ctx, countSpan := tracing.StartSpan(ctx, "db.query.admin_log_count", trace.WithAttributes(
		attribute.String("db.operation", "count"),
		attribute.String("db.table", "admin_log"),
	))
	total, err := query.Count()
	countSpan.End()
	if err != nil {
		tracing.SetSpanError(span, err)
		return nil, fmt.Errorf("获取管理员日志总数失败: %v", err)
	}

	// 数据库查询span - 获取列表数据
	ctx, listSpan := tracing.StartSpan(ctx, "db.query.admin_log_list", trace.WithAttributes(
		attribute.String("db.operation", "select"),
		attribute.String("db.table", "admin_log"),
		attribute.Int("limit", int(size)),
		attribute.Int("offset", int((page-1)*size)),
	))

	// 方案1: 尝试使用数据库函数直接格式化时间
	var rawResults []map[string]interface{}
	err = query.Fields(
		"admin_username",
		"ip",
		"remark",
		"DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') as created_at_formatted",
		"UNIX_TIMESTAMP(created_at) as created_at_unix",
	).Page(int(page), int(size)).
		OrderDesc("created_at").
		Scan(&rawResults)

	if err == nil && len(rawResults) > 0 {
		// 使用数据库格式化的结果
		listSpan.End()

		var logList []*v1.AdminLogInfo
		for _, result := range rawResults {
			createdAtFormatted := result["created_at_formatted"]
			createdAtStr := ""

			if createdAtFormatted != nil {
				createdAtStr = fmt.Sprintf("%v", createdAtFormatted)
				// 检查是否为无效时间
				if createdAtStr == "0000-00-00 00:00:00" || createdAtStr == "" {
					createdAtStr = ""
				}
			}

			logList = append(logList, &v1.AdminLogInfo{
				Username:  fmt.Sprintf("%v", result["admin_username"]),
				Ip:        fmt.Sprintf("%v", result["ip"]),
				Remark:    fmt.Sprintf("%v", result["remark"]),
				CreatedAt: createdAtStr,
			})
		}

		// 确保返回空数组而不是nil
		if logList == nil {
			logList = []*v1.AdminLogInfo{}
		}

		middleware.LogWithTrace(ctx, "info", fmt.Sprintf("使用数据库格式化获取管理员日志列表成功，总数: %d，返回: %d", total, len(logList)))

		return &v1.GetAdminLogsRes{
			List:  logList,
			Count: int32(total),
		}, nil
	}

	// 查询管理员日志数据
	var logs []entity.AdminLog
	err = query.Fields("admin_username", "ip", "remark", "created_at").
		Page(int(page), int(size)).
		OrderDesc("created_at").
		Scan(&logs)
	listSpan.End()

	if err != nil {
		tracing.SetSpanError(span, err)
		return nil, fmt.Errorf("获取管理员日志列表失败: %v", err)
	}

	// 转换为响应格式
	var logList []*v1.AdminLogInfo
	for _, log := range logs {
		// 使用修复后的工具函数处理时间格式化
		createdAt := util.FormatTime(log.CreatedAt)

		logList = append(logList, &v1.AdminLogInfo{
			Username:  log.AdminUsername,
			Ip:        log.Ip,
			Remark:    log.Remark,
			CreatedAt: createdAt,
		})
	}

	// 确保返回空数组而不是nil
	if logList == nil {
		logList = []*v1.AdminLogInfo{}
	}

	tracing.SetSpanAttributes(span,
		attribute.Int("total_count", total),
		attribute.Int("returned_count", len(logList)),
	)

	middleware.LogWithTrace(ctx, "info", fmt.Sprintf("获取管理员日志列表成功，总数: %d，返回: %d", total, len(logList)))

	return &v1.GetAdminLogsRes{
		List:  logList,
		Count: int32(total),
	}, nil
}

// Menus 获取菜单列表
func (s *sAdmin) Menus(ctx context.Context, req *v1.MenusReq) (*v1.MenusRes, error) {
	// 创建Jaeger span
	ctx, span := tracing.StartSpan(ctx, "admin.Menus", trace.WithAttributes(
		attribute.String("method", "Menus"),
	))
	defer span.End()

	middleware.LogWithTrace(ctx, "info", "获取菜单列表请求")

	// 从上下文中获取当前管理员ID
	adminId, exists := middleware.GetAdminIdFromContext(ctx)
	if !exists {
		tracing.AddSpanEvent(span, "admin_id_not_found")
		middleware.LogWithTrace(ctx, "error", "无法获取管理员ID")
		return nil, fmt.Errorf("未登录或登录已过期")
	}

	tracing.SetSpanAttributes(span, attribute.Int("admin_id", int(adminId)))
	middleware.LogWithTrace(ctx, "info", "获取到管理员ID: %d", adminId)

	// 查询管理员信息
	ctx, querySpan := tracing.StartSpan(ctx, "db.query.admin", trace.WithAttributes(
		attribute.String("db.operation", "select"),
		attribute.String("db.table", "admin"),
	))

	var admin *entity.Admin
	err := dao.Admin.Ctx(ctx).Where(do.Admin{Id: adminId}).Scan(&admin)
	querySpan.End()

	if err != nil {
		tracing.SetSpanError(span, err)
		tracing.SetSpanError(querySpan, err)
		middleware.LogWithTrace(ctx, "error", "查询管理员信息失败: %v", err)
		return nil, fmt.Errorf("查询管理员信息失败: %v", err)
	}

	if admin == nil {
		tracing.AddSpanEvent(span, "admin_not_found", attribute.Int("admin_id", int(adminId)))
		middleware.LogWithTrace(ctx, "warning", "管理员不存在 - ID: %d", adminId)
		return nil, fmt.Errorf("管理员不存在")
	}

	// 检查管理员状态
	if admin.Status != 1 {
		tracing.AddSpanEvent(span, "admin_status_invalid",
			attribute.Int("admin_id", int(adminId)),
			attribute.Int("status", admin.Status),
		)
		middleware.LogWithTrace(ctx, "warning", "管理员状态异常 - ID: %d, 状态: %d", adminId, admin.Status)
		return nil, fmt.Errorf("账号已被禁用")
	}

	// 构建菜单列表 - 这里应该根据管理员的角色权限查询实际的菜单
	// 暂时返回一个示例菜单结构，参考 go_service 项目的返回格式
	menus := s.buildMenusForMenusAPI(ctx, admin)

	res := &v1.MenusRes{
		Menus: menus,
	}

	tracing.AddSpanEvent(span, "get_menus_success",
		attribute.Int("admin_id", int(adminId)),
		attribute.String("username", admin.Username),
		attribute.Int("menu_count", len(menus)),
	)
	tracing.SetSpanAttributes(span, attribute.Bool("success", true))

	middleware.LogWithTrace(ctx, "info", "获取菜单列表成功 - ID: %d, 用户名: %s, 菜单数量: %d",
		adminId, admin.Username, len(menus))

	return res, nil
}

// buildMenusForMenusAPI 为Menus API构建菜单结构 - 从数据库动态查询
func (s *sAdmin) buildMenusForMenusAPI(ctx context.Context, admin *entity.Admin) []*v1.MenuInfo {
	//admin.AdminRoleId
	var adminRole entity.AdminRole
	errRole := dao.AdminRole.Ctx(ctx).Where("id = ?", admin.AdminRoleId).Scan(&adminRole)
	if errRole != nil {
		middleware.LogWithTrace(ctx, "error", "查询角色失败: %v", errRole)
		return []*v1.MenuInfo{}
	}

	if len(adminRole.Permissions) == 0 {
		middleware.LogWithTrace(ctx, "error", "查询角色ID %d 的权限节点为空", admin.AdminRoleId)
		return []*v1.MenuInfo{}
	}

	// 将权限字符串分割成ID切片
	permissionIds := strings.Split(adminRole.Permissions, ",")
	if len(permissionIds) == 0 {
		middleware.LogWithTrace(ctx, "error", "角色ID %d 的权限解析为空", admin.AdminRoleId)
		return []*v1.MenuInfo{}
	}

	middleware.LogWithTrace(ctx, "info", "角色ID %d 的权限列表: %v", admin.AdminRoleId, permissionIds)

	// 查询管理员角色权限
	var rolePermissions []entity.AdminPermission
	err := dao.AdminPermission.Ctx(ctx).
		WhereIn("id", permissionIds).
		Where("status = ?", 1). // 只查询启用的权限
		OrderAsc("sort").
		OrderAsc("id").
		Scan(&rolePermissions)

	if err != nil {
		middleware.LogWithTrace(ctx, "error", "查询菜单权限失败: %v", err)
		return []*v1.MenuInfo{}
	}

	// 构建菜单树结构
	menuMap := make(map[int]*v1.MenuInfo)
	var rootMenus []*v1.MenuInfo

	// 第一遍遍历：创建所有菜单项
	for _, perm := range rolePermissions {
		menuInfo := &v1.MenuInfo{
			Id:          int32(perm.Id),
			Type:        int32(perm.Type),
			Name:        perm.Name,
			BackendUrl:  perm.BackendUrl,
			FrontendUrl: perm.FrontendUrl,
			Open:        perm.Type == 1, // 菜单类型默认展开，操作权限默认不展开
			Checked:     false,
			Path:        perm.FrontendUrl, // 使用frontend_url作为path
			Sort:        int32(perm.Sort),
			Children:    []*v1.MenuInfo{},
			Icon:        perm.Icon,
		}
		menuMap[int(perm.Id)] = menuInfo
	}

	// 第二遍遍历：构建父子关系
	for _, perm := range rolePermissions {
		menuInfo := menuMap[int(perm.Id)]
		if perm.ParentId == 0 {
			// 根菜单
			rootMenus = append(rootMenus, menuInfo)
		} else {
			// 子菜单
			if parentMenu, exists := menuMap[perm.ParentId]; exists {
				parentMenu.Children = append(parentMenu.Children, menuInfo)
			}
		}
	}

	middleware.LogWithTrace(ctx, "info", "从数据库构建菜单成功 - 总权限数: %d, 根菜单数: %d",
		len(rolePermissions), len(rootMenus))

	return rootMenus
}
