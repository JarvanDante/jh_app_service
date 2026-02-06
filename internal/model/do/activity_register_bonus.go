// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// ActivityRegisterBonus is the golang structure of table activity_register_bonus for DAO operations like Where/Data.
type ActivityRegisterBonus struct {
	g.Meta       `orm:"table:activity_register_bonus, do:true"`
	Id           any         //
	SiteCode     any         // 站点CODE
	ActivityId   any         // 活动ID
	RewardAmount any         // 注册奖励金额
	LimitIp      any         // 同IP注册限制次数
	LimitDevice  any         // 同设备注册限制次数
	CreatedAt    *gtime.Time //
	UpdatedAt    *gtime.Time //
}
