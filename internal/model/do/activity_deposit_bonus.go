// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// ActivityDepositBonus is the golang structure of table activity_deposit_bonus for DAO operations like Where/Data.
type ActivityDepositBonus struct {
	g.Meta            `orm:"table:activity_deposit_bonus, do:true"`
	Id                any         //
	SiteCode          any         // 站点CODE
	ActivityId        any         // 活动ID
	TriggerEvent      any         // 发放触发:1=充值总和,2=首充
	RewardType        any         // 奖励类型:1=固定,2=比例
	MinRechargeAmount any         // 触发最小充值金额
	RewardAmount      any         // 奖励金额/比例值
	CreatedAt         *gtime.Time //
	UpdatedAt         *gtime.Time //
}
