package dao

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// GetDynamicModel 获取动态数据库Model的通用辅助函数
func GetDynamicModel(ctx context.Context, tableName string, originalModel *gdb.Model) *gdb.Model {
	// 从上下文获取动态数据库组名
	if dbGroup := ctx.Value("dynamic_db_group"); dbGroup != nil {
		if groupName, ok := dbGroup.(string); ok && groupName != "" {
			// 检查配置是否存在
			if gdb.GetConfig(groupName) != nil {
				g.Log().Infof(ctx, "DAO Helper: 使用动态数据库组: %s, 表: %s", groupName, tableName)
				// 创建新的Model使用指定的数据库组
				return g.DB(groupName).Model(tableName).Safe().Ctx(ctx)
			} else {
				g.Log().Warningf(ctx, "DAO Helper: 动态数据库组配置不存在，使用默认数据库: %s", groupName)
			}
		}
	}

	// 返回原始Model
	return originalModel
}
