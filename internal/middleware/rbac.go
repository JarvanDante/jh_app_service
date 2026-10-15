package middleware

import (
	"context"
	"strings"
	"sync"

	"jh_app_service/internal/dao"
	"jh_app_service/internal/model/do"
	"jh_app_service/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// =============================================================================
// A3 — 后台 RBAC 服务端强制（默认 audit 模式，不真正拦截）
//
// 背景：admin_role.permissions（CSV 权限ID）此前只用于前端菜单渲染，
//   admin_permission.backend_url 从不与实际 gRPC 请求比对，任何持有有效
//   管理员 token 的人都能调用任意后台接口（水平/垂直越权）。
//
// 本实现：新增一元拦截器，把当前 gRPC 方法（info.FullMethod）映射到所需的
//   权限ID，加载该 admin 角色的 permissions 判断是否放行。
//
//   - 默认 audit 模式（backend.rbacEnforce=false）：只记录“本应拒绝”日志，
//     不真正拦截，避免映射不全导致管理员被锁死。
//   - enforce 模式（backend.rbacEnforce=true）：真正返回权限不足错误。
//
// ⚠️ 启用 enforce 前必须先做的事：
//   1. 补全下面的 methodPermissionMap：把每个后台 gRPC 全方法名
//      （形如 "/admin.v1.Admin/GetAdminList"）映射到 admin_permission 表里
//      对应的权限ID。当前该表 backend_url 存的是 HTTP 路径而非 gRPC 方法名，
//      因此这里用代码内映射表，需与产品/前端确认权限点后逐条补全。
//   2. 用真实角色数据联调（尤其是超管角色应覆盖全部权限），确认不会误杀。
//   3. 观察一段时间 audit 日志（"RBAC本应拒绝"）确认无误报后再打开开关。
//
// 在映射表补全前，enforce 即使打开也只对映射表里已列出的方法生效；未列出的
// 方法一律放行（fail-open），这是为了避免锁死，属于有意为之的保守策略。
// =============================================================================

// methodPermissionMap 后台 gRPC 全方法名 -> 所需权限ID。
// TODO(A3): 补全此映射后方可安全启用 backend.rbacEnforce。留空即对所有方法放行。
// 示例（占位，需与 admin_permission 表权限ID对齐后启用）：
//
//	"/admin.v1.Admin/CreateAdmin": 2,
//	"/admin.v1.Admin/DeleteAdmin": 2,
//	"/role.v1.Role/SavePermission": 4,
var methodPermissionMap = map[string]int{}

// rbacExemptMethods 无需 RBAC 校验的方法（登录/登出/刷新/获取自身信息等）。
// 前台方法通过前缀判断豁免，见 isBackendMethod。
var rbacExemptMethods = map[string]struct{}{}

var (
	rbacEnforceOnce sync.Once
	rbacEnforce     bool
)

// isRBACEnforce 读取配置 backend.rbacEnforce（默认 false=audit）。只读一次。
func isRBACEnforce(ctx context.Context) bool {
	rbacEnforceOnce.Do(func() {
		rbacEnforce = g.Cfg().MustGet(ctx, "backend.rbacEnforce", false).Bool()
	})
	return rbacEnforce
}

// isBackendMethod 仅对后台接口做 RBAC。前台接口（public/user 前台等）不在此列。
// 这里用一个简单启发式：后台方法所在的 proto package 都在 backend 侧注册。
// 由于本服务前后台方法名空间未强制区分，保守起见：只有出现在 methodPermissionMap
// 里的方法才会被校验，其余一律视为“未纳入 RBAC”而放行（fail-open）。
func isBackendMethod(fullMethod string) bool {
	if _, ok := rbacExemptMethods[fullMethod]; ok {
		return false
	}
	_, mapped := methodPermissionMap[fullMethod]
	return mapped
}

// getAdminPermissionSet 加载当前管理员角色的权限ID集合。
func getAdminPermissionSet(ctx context.Context) (map[int]struct{}, bool) {
	adminId, ok := GetAdminIdFromContext(ctx)
	if !ok || adminId == 0 {
		return nil, false
	}

	var admin *entity.Admin
	if err := dao.Admin.Ctx(ctx).Where(do.Admin{Id: int(adminId)}).Scan(&admin); err != nil || admin == nil {
		return nil, false
	}

	var role *entity.AdminRole
	if err := dao.AdminRole.Ctx(ctx).Where(do.AdminRole{Id: uint(admin.AdminRoleId)}).Scan(&role); err != nil || role == nil {
		return nil, false
	}

	set := make(map[int]struct{})
	for _, idStr := range strings.Split(role.Permissions, ",") {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}
		if id := g.NewVar(idStr).Int(); id > 0 {
			set[id] = struct{}{}
		}
	}
	return set, true
}

// RBACInterceptor 后台权限校验一元拦截器。
// 默认 audit：只记录“本应拒绝”日志，放行请求；enforce 打开后才真正拦截。
func RBACInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 未纳入映射表的方法直接放行（fail-open，避免锁死）。
	if !isBackendMethod(info.FullMethod) {
		return handler(ctx, req)
	}

	requiredPerm := methodPermissionMap[info.FullMethod]

	permSet, ok := getAdminPermissionSet(ctx)
	allowed := ok
	if ok {
		if _, has := permSet[requiredPerm]; !has {
			allowed = false
		}
	}

	if !allowed {
		if isRBACEnforce(ctx) {
			LogWithTrace(ctx, "warning", "RBAC拒绝: method=%s 缺少权限ID=%d", info.FullMethod, requiredPerm)
			return nil, status.Error(codes.PermissionDenied, "权限不足")
		}
		// audit 模式：只记录，不拦截
		LogWithTrace(ctx, "warning", "RBAC本应拒绝(audit,放行): method=%s 缺少权限ID=%d", info.FullMethod, requiredPerm)
	}

	return handler(ctx, req)
}
