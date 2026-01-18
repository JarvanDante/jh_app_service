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
	if adminId, ok := ctx.Value(SiteIdContextKey).(uint); ok {
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
	if siteId, ok := ctx.Value(AdminIdContextKey).(uint); ok {
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
