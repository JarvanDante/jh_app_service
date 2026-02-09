// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// ActivityWheelBonus is the golang structure of table activity_wheel_bonus for DAO operations like Where/Data.
type ActivityWheelBonus struct {
	g.Meta            `orm:"table:activity_wheel_bonus, do:true"`
	Id                any         //
	SiteCode          any         // 站点CODE
	ActivityId        any         // 活动ID
	SegmentNo         any         // 转盘扇区序号(1~12)
	PrizeName         any         // 奖项名称
	RewardAmount      any         // 奖金额(谢谢参与填0)
	Weight            any         // 抽中权重(越大概率越高)
	UserDailyWinLimit any         // 单用户每日可中次数(0不限)
	MinVipLevel       any         // 最低VIP等级(0不限)
	Icon              any         // 奖项图标
	CreatedAt         *gtime.Time // 创建时间
	UpdatedAt         *gtime.Time // 更新时间
}
