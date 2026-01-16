package dao

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// GetDynamicModel 获取动态数据库Model的通用辅助函数
func GetDynamicModel(ctx context.Context, tableName string, originalModel *gdb.Model) *gdb.Model {
	g.Log().Debugf(ctx, "GetDynamicModel: 表=%s, 开始检查动态数据库", tableName)

	// 从上下文获取动态数据库组名
	if dbGroup := ctx.Value("dynamic_db_group"); dbGroup != nil {
		if groupName, ok := dbGroup.(string); ok && groupName != "" {
			g.Log().Debugf(ctx, "GetDynamicModel: 找到dynamic_db_group=%s", groupName)
			// 检查配置是否存在
			if gdb.GetConfig(groupName) != nil {
				g.Log().Infof(ctx, "DAO Helper: 使用动态数据库组: %s, 表: %s", groupName, tableName)
				// 创建新的Model使用指定的数据库组
				// 注意：不使用originalModel，直接创建新的，避免继承DynamicDBHandler
				return g.DB(groupName).Model(tableName).Safe().Ctx(ctx)
			} else {
				g.Log().Warningf(ctx, "DAO Helper: 动态数据库组配置不存在，使用默认数据库: %s", groupName)
			}
		} else {
			g.Log().Debugf(ctx, "GetDynamicModel: dynamic_db_group类型错误或为空")
		}
	} else {
		g.Log().Debugf(ctx, "GetDynamicModel: 上下文中没有dynamic_db_group")
	}

	// 返回原始Model，确保context正确
	g.Log().Debugf(ctx, "GetDynamicModel: 使用原始Model")
	return originalModel.Ctx(ctx)
}
