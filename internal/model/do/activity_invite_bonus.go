// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// ActivityInviteBonus is the golang structure of table activity_invite_bonus for DAO operations like Where/Data.
type ActivityInviteBonus struct {
	g.Meta              `orm:"table:activity_invite_bonus, do:true"`
	Id                  any         //
	SiteCode            any         // 站点CODE
	ActivityId          any         // 活动ID
	TriggerEvent        any         // 发放触发:1充值总和2首充
	RewardType          any         // 奖励类型:1=固定,2=比例
	MinRechargeAmount   any         // 触发最小计算金额
	InviterRewardAmount any         // 邀请人奖励金额
	CreatedAt           *gtime.Time //
	UpdatedAt           *gtime.Time //
}
