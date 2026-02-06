// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ActivityDepositBonus is the golang structure for table activity_deposit_bonus.
type ActivityDepositBonus struct {
	Id                uint        `json:"id"                orm:"id"                  description:""`
	SiteCode          string      `json:"siteCode"          orm:"site_code"           description:"站点CODE"`
	ActivityId        int         `json:"activityId"        orm:"activity_id"         description:"活动ID"`
	TriggerEvent      int         `json:"triggerEvent"      orm:"trigger_event"       description:"发放触发:1=充值总和,2=首充"`
	RewardType        int         `json:"rewardType"        orm:"reward_type"         description:"奖励类型:1=固定,2=比例"`
	MinRechargeAmount float64     `json:"minRechargeAmount" orm:"min_recharge_amount" description:"触发最小充值金额"`
	RewardAmount      float64     `json:"rewardAmount"      orm:"reward_amount"       description:"奖励金额/比例值"`
	CreatedAt         *gtime.Time `json:"createdAt"         orm:"created_at"          description:""`
	UpdatedAt         *gtime.Time `json:"updatedAt"         orm:"updated_at"          description:""`
}
