package middleware

import (
	"context"
	"strconv"
	"strings"

	"jh_app_service/internal/registry"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// DBInterceptor 数据库拦截器 - 完全无侵入的动态数据库路由
type DBInterceptor struct {
	dbManager *registry.DBConfigManagerV2
}

// NewDBInterceptor 创建数据库拦截器
func NewDBInterceptor() *DBInterceptor {
	return &DBInterceptor{
		dbManager: registry.GetDBConfigManagerV2(),
	}
}

// UnaryServerInterceptor 一元RPC拦截器
func (interceptor *DBInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 创建带有动态数据库的上下文
		newCtx := interceptor.createDynamicDBContext(ctx)

		// 调用原始处理器
		return handler(newCtx, req)
	}
}

// StreamServerInterceptor 流式RPC拦截器
func (interceptor *DBInterceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// 创建带有动态数据库的上下文
		newCtx := interceptor.createDynamicDBContext(ss.Context())

		// 包装ServerStream
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
			ctx:          newCtx,
		}

		// 调用原始处理器
		return handler(srv, wrappedStream)
	}
}

// createDynamicDBContext 创建带有动态数据库的上下文
func (interceptor *DBInterceptor) createDynamicDBContext(ctx context.Context) context.Context {
	// 获取站点代码
	siteCode := interceptor.getSiteCodeFromContext(ctx)
	g.Log().Debugf(ctx, "DB拦截器: 获取到站点代码: %s", siteCode)

	if siteCode == "" {
		// 没有站点信息，返回原始上下文
		g.Log().Debugf(ctx, "DB拦截器: 没有站点信息，使用默认数据库")
		return ctx
	}

	// 将站点代码设置到上下文中，供DAO层使用
	newCtx := context.WithValue(ctx, "dynamic_db_group", "site_"+siteCode)
	g.Log().Debugf(ctx, "DB拦截器: 设置动态数据库组: site_%s", siteCode)
	return newCtx
}

// getSiteCodeFromContext 从上下文获取站点代码
func (interceptor *DBInterceptor) getSiteCodeFromContext(ctx context.Context) string {
	// 首先尝试从上下文中获取site_code
	if siteCode := ctx.Value("site_code"); siteCode != nil {
		if siteCodeStr, ok := siteCode.(string); ok {
			g.Log().Debugf(ctx, "DB拦截器: 从上下文获取site_code: %s", siteCodeStr)
			return siteCodeStr
		}
	}

	// 如果没有site_code，尝试从gRPC metadata获取site_id
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if siteIDs := md.Get("site_id"); len(siteIDs) > 0 {
			if siteIDStr := strings.TrimSpace(siteIDs[0]); siteIDStr != "" {
				g.Log().Debugf(ctx, "DB拦截器: 从metadata获取site_id: %s", siteIDStr)
				// 转换为int64
				if siteID, err := strconv.ParseInt(siteIDStr, 10, 64); err == nil && siteID > 0 {
					// 根据site_id查找对应的site_code
					siteCode := interceptor.dbManager.GetSiteCodeBySiteID(siteID)
					g.Log().Debugf(ctx, "DB拦截器: site_id=%d 对应 site_code=%s", siteID, siteCode)
					return siteCode
				} else {
					g.Log().Warningf(ctx, "DB拦截器: site_id转换失败: %s, error: %v", siteIDStr, err)
				}
			}
		} else {
			g.Log().Debugf(ctx, "DB拦截器: metadata中没有site_id")
		}
	} else {
		g.Log().Debugf(ctx, "DB拦截器: 无法从上下文获取metadata")
	}

	return ""
}

// getDynamicDB 获取动态数据库连接
func (interceptor *DBInterceptor) getDynamicDB(ctx context.Context, siteCode string) gdb.DB {
	// 创建数据库配置组名称
	groupName := "site_" + siteCode

	// 检查配置是否存在
	if gdb.GetConfig(groupName) != nil {
		g.Log().Debugf(ctx, "使用站点数据库: %s", groupName)
		return g.DB(groupName)
	}

	g.Log().Warningf(ctx, "站点数据库配置不存在，使用默认数据库: %s", groupName)
	return nil // 返回nil表示使用默认数据库
}

// wrappedServerStream 包装的ServerStream
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context 返回包装的上下文
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
