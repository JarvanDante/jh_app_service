// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// ActivitySigninBonus is the golang structure of table activity_signin_bonus for DAO operations like Where/Data.
type ActivitySigninBonus struct {
	g.Meta       `orm:"table:activity_signin_bonus, do:true"`
	Id           any         //
	SiteCode     any         // 站点CODE
	ActivityId   any         // 活动ID
	DayNo        any         // 日号(1-31)
	RewardAmount any         // 当日固定金额
	CreatedAt    *gtime.Time // 创建时间
	UpdatedAt    *gtime.Time // 更新时间
}
