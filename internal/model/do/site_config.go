// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SiteConfig is the golang structure of table site_config for DAO operations like Where/Data.
type SiteConfig struct {
	g.Meta               `orm:"table:site_config, do:true"`
	Id                   any         //
	SiteId               any         // 站点ID
	SwitchRegister       any         // 是否开放注册。1=开放;0=关闭
	SwitchGrade          any         // 是否开放会员等级。1=开放;0=关闭
	RegisterTimeInterval any         // 同一IP重复注册。设定时间内，同一IP将无法进行多次注册。0或留空表示不限制
	DefaultGradeId       any         // 默认用户等级ID
	DefaultLevelId       any         // 默认用户层级ID
	DefaultAgentId       any         // 默认代理ID
	IsClose              any         // 是否关站。1=是；0=否
	MobileLogo           any         // 手机端logo
	MinWithdraw          any         // 单笔最低提现金额
	MaxWithdraw          any         // 单笔最高提现金额
	CloseReason          any         // 关站原因
	UrlService           any         // 客服链接
	UrlTg                any         // telegram链接
	UrlMobile            any         // 手机域名
	CreatedAt            *gtime.Time //
	UpdatedAt            *gtime.Time //
	SwitchSign           any         //
}
