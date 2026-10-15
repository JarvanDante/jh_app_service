package middleware

import (
	"context"
)

const (
	SiteIdContextKey   = "site_id"
	SiteCodeContextKey = "site_code"
	AdminIdContextKey  = "admin_id"
	UserIdContextKey   = "user_id"
	ClientIPContextKey = "client_ip"
)

// GetAdminIdFromContext 从上下文中获取管理员ID
// 优先从 gRPC metadata 获取，如果没有再从 context 获取
func GetAdminIdFromContext(ctx context.Context) (uint, bool) {
	// 首先尝试从 gRPC metadata 获取
	if adminId, ok := GetAdminIdFromGRPCMetadata(ctx); ok {
		return adminId, true
	}

	// 如果 gRPC metadata 中没有，尝试从 context 获取
	if adminId, ok := ctx.Value(AdminIdContextKey).(uint); ok {
		return adminId, true
	}
	return 0, false
}

//GetSiteIdFromContext
/**
 * @desc：获取siteId -优先从 gRPC metadata 获取，如果没有再从 context 获取
 * @param ctx
 * @return uint
 * @return bool
 * @author : Carson
 */
func GetSiteIdFromContext(ctx context.Context) (uint, bool) {
	// 首先尝试从 gRPC metadata 获取
	if siteId, ok := GetSiteIdFromGRPCMetadata(ctx); ok {
		return siteId, true
	}

	// 如果 gRPC metadata 中没有，尝试从 context 获取
	if siteId, ok := ctx.Value(SiteIdContextKey).(uint); ok {
		return siteId, true
	}
	return 0, false
}

//GetAdminSiteIdFromContext
/**
 * @desc：获取后台管理员的权威站点ID（租户隔离唯一可信来源）。
 *        仅从 gRPC metadata 的 admin_site_id 读取，该值由网关基于管理员 token 注入，不可伪造。
 *        后台接口必须用它作为租户边界，禁止信任请求体里的 site_id。
 * @param ctx
 * @return uint  权威站点ID
 * @return bool  是否成功获取
 */
func GetAdminSiteIdFromContext(ctx context.Context) (uint, bool) {
	return GetAdminSiteIdFromGRPCMetadata(ctx)
}

// ResolveTenantSiteId 解析当前请求的租户站点ID。
// 后台走 admin_site_id；前台/服务间调用走网关或上游注入的 site_id。
// 都不信任请求体里的 site_id。写接口仍应只用 GetAdminSiteIdFromContext。
func ResolveTenantSiteId(ctx context.Context) (uint, bool) {
	if siteId, ok := GetAdminSiteIdFromContext(ctx); ok && siteId > 0 {
		return siteId, true
	}
	if siteId, ok := GetSiteIdFromContext(ctx); ok && siteId > 0 {
		return siteId, true
	}
	return 0, false
}

//GetUserIdFromContext
/**
 * @desc：获取userId -优先从 gRPC metadata 获取，如果没有再从 context 获取
 * @param ctx
 * @return uint
 * @return bool
 * @author : Carson
 */
func GetUserIdFromContext(ctx context.Context) (uint, bool) {
	// 首先尝试从 gRPC metadata 获取
	if userId, ok := GetUserIdFromGRPCMetadata(ctx); ok {
		return userId, true
	}

	// 如果 gRPC metadata 中没有，尝试从 context 获取
	if userId, ok := ctx.Value(UserIdContextKey).(uint); ok {
		return userId, true
	}
	return 0, false
}

//GetSiteCodeFromContext
/**
 * @desc：获取siteCode -优先从 gRPC metadata 获取，如果没有再从 context 获取
 * @param ctx
 * @return uint
 * @return bool
 * @author : Carson
 */
func GetSiteCodeFromContext(ctx context.Context) (string, bool) {
	// 首先尝试从 gRPC metadata 获取
	if siteCode, ok := GetSiteCodeFromGRPCMetadata(ctx); ok {
		return siteCode, true
	}

	// 如果 gRPC metadata 中没有，尝试从 context 获取
	if siteCode, ok := ctx.Value(SiteCodeContextKey).(string); ok {
		return siteCode, true
	}
	return "", false
}

// SetAdminIdToContext 将管理员ID设置到上下文中
func SetAdminIdToContext(ctx context.Context, adminId uint) context.Context {
	return context.WithValue(ctx, AdminIdContextKey, adminId)
}

// GetClientIPFromContext 从上下文中获取客户端IP
// 优先从 gRPC metadata 获取，如果没有再从 context 获取
func GetClientIPFromContext(ctx context.Context) string {
	// 首先尝试从 gRPC metadata 获取
	if ip := GetClientIPFromGRPCMetadata(ctx); ip != "127.0.0.1" {
		return ip
	}

	// 如果 gRPC metadata 中没有，尝试从 context 获取
	if ip, ok := ctx.Value(ClientIPContextKey).(string); ok {
		return ip
	}
	return "127.0.0.1"
}

// SetClientIPToContext 将客户端IP设置到上下文中
func SetClientIPToContext(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, ClientIPContextKey, ip)
}
