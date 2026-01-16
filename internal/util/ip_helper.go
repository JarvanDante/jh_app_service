package util

import (
	"context"
	"strings"

	"google.golang.org/grpc/metadata"
)

// GetClientIP 从gRPC metadata中获取客户端真实IP
// Gateway会将client_ip添加到metadata中
func GetClientIP(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "127.0.0.1"
	}

	// 从metadata获取client_ip
	if ips := md.Get("client_ip"); len(ips) > 0 {
		ip := ips[0]
		// 如果是多个IP（通过代理），取第一个
		if idx := strings.Index(ip, ","); idx > 0 {
			ip = strings.TrimSpace(ip[:idx])
		}
		if ip != "" {
			return ip
		}
	}

	return "127.0.0.1"
}
