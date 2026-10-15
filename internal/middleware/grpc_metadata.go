package middleware

import (
	"context"
	"strconv"
	"strings"

	"google.golang.org/grpc/metadata"
)

// GetAdminIdFromGRPCMetadata 从 gRPC metadata 中获取管理员ID
func GetAdminIdFromGRPCMetadata(ctx context.Context) (uint, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, false
	}

	adminIds := md.Get("admin_id")
	if len(adminIds) == 0 {
		return 0, false
	}

	adminId, err := strconv.ParseUint(adminIds[0], 10, 32)
	if err != nil {
		return 0, false
	}

	return uint(adminId), true
}

// GetSiteIdFromGRPCMetadata 从 gRPC metadata 中获取商户ID
func GetSiteIdFromGRPCMetadata(ctx context.Context) (uint, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, false
	}

	siteIDs := md.Get("site_id")
	if len(siteIDs) == 0 {
		return 0, false
	}

	siteIDStr := strings.TrimSpace(siteIDs[0])
	if siteIDStr == "" {
		return 0, false
	}

	siteID, err := strconv.ParseInt(siteIDStr, 10, 64)
	if err != nil || siteID <= 0 {
		return 0, false
	}

	return uint(siteID), true
}

// GetAdminSiteIdFromGRPCMetadata 从 gRPC metadata 中获取管理员 token 的权威站点ID。
// 该值由网关从管理员 JWT (claims.SiteID) 注入 metadata key "admin_site_id"，
// 不接受请求头/请求体伪造，是后台租户隔离的唯一可信来源。
func GetAdminSiteIdFromGRPCMetadata(ctx context.Context) (uint, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, false
	}

	adminSiteIDs := md.Get("admin_site_id")
	if len(adminSiteIDs) == 0 {
		return 0, false
	}

	adminSiteIDStr := strings.TrimSpace(adminSiteIDs[0])
	if adminSiteIDStr == "" {
		return 0, false
	}

	adminSiteID, err := strconv.ParseInt(adminSiteIDStr, 10, 64)
	if err != nil || adminSiteID <= 0 {
		return 0, false
	}

	return uint(adminSiteID), true
}

// GetSiteCodeFromGRPCMetadata 从 gRPC metadata 中获取商户CODE
func GetSiteCodeFromGRPCMetadata(ctx context.Context) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}

	siteCodes := md.Get("site_code")
	if len(siteCodes) == 0 {
		return "", false
	}

	siteCode := strings.TrimSpace(siteCodes[0])
	if siteCode == "" {
		return "", false
	}

	return siteCode, true
}

// GetUserIdFromGRPCMetadata 从 gRPC metadata 中获取用户ID
func GetUserIdFromGRPCMetadata(ctx context.Context) (uint, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, false
	}

	userIds := md.Get("user_id")
	if len(userIds) == 0 {
		return 0, false
	}

	userId, err := strconv.ParseUint(userIds[0], 10, 32)
	if err != nil {
		return 0, false
	}

	return uint(userId), true
}

// GetClientIPFromGRPCMetadata 从 gRPC metadata 中获取客户端IP
func GetClientIPFromGRPCMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "127.0.0.1"
	}

	ips := md.Get("client_ip")
	if len(ips) == 0 {
		return "127.0.0.1"
	}

	return ips[0]
}
