package dao

import (
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// DynamicDBHandler 动态数据库处理器 - 完全无侵入的方式
func DynamicDBHandler(model *gdb.Model) *gdb.Model {
	ctx := model.GetCtx()
	if ctx == nil {
		g.Log().Debug(nil, "DynamicDBHandler: 上下文为空")
		return model
	}

	// 从上下文获取动态数据库组名
	if dbGroup := ctx.Value("dynamic_db_group"); dbGroup != nil {
		if groupName, ok := dbGroup.(string); ok && groupName != "" {
			g.Log().Debugf(ctx, "DynamicDBHandler: 检查数据库组: %s", groupName)
			// 检查配置是否存在
			if gdb.GetConfig(groupName) != nil {
				g.Log().Infof(ctx, "DynamicDBHandler: 使用动态数据库组: %s", groupName)

				// 使用Clone方法复制Model到新的数据库组
				newDB := g.DB(groupName)
				clonedModel := model.Clone()

				// 设置新的数据库
				return clonedModel.DB(newDB)
			} else {
				g.Log().Warningf(ctx, "DynamicDBHandler: 动态数据库组配置不存在，使用默认数据库: %s", groupName)
			}
		}
	} else {
		g.Log().Debugf(ctx, "DynamicDBHandler: 上下文中没有dynamic_db_group")
	}

	// 返回原始model
	return model
}
