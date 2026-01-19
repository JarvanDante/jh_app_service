// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// UserGrade is the golang structure of table user_grade for DAO operations like Where/Data.
type UserGrade struct {
	g.Meta               `orm:"table:user_grade, do:true"`
	Id                   any         //
	SiteId               any         // 站点ID
	Name                 any         // 会员等级名称
	PointsUpgrade        any         // 升级消耗积分
	BonusUpgrade         any         // 额外返水比例: 体育
	BonusBirthday        any         // 生日彩金
	RebatePercentSports  any         // 额外返水比例: 体育
	RebatePercentLottery any         // 额外返水比例: 彩票
	RebatePercentLive    any         // 额外返水比例: 真人视讯
	RebatePercentEgame   any         // 额外返水比例: 电子游戏
	RebatePercentPoker   any         // 额外返水比例：棋牌
	FieldsDisable        any         // 字段开关，用来关闭哪些字段
	AutoProviding        any         // 哪些字段的业务是自动发放的
	Status               any         // 状态.1=可用；0=禁用
	CreatedAt            *gtime.Time //
	UpdatedAt            *gtime.Time //
	PaymentLimit         any         // 充值要求
	BetLimit             any         // 投注要求
	MoneyLimit           any         // 升级送
}
