// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ActivityInviteBonus is the golang structure for table activity_invite_bonus.
type ActivityInviteBonus struct {
	Id                  uint        `json:"id"                  orm:"id"                    description:""`
	SiteCode            string      `json:"siteCode"            orm:"site_code"             description:"站点CODE"`
	ActivityId          int         `json:"activityId"          orm:"activity_id"           description:"活动ID"`
	TriggerEvent        int         `json:"triggerEvent"        orm:"trigger_event"         description:"发放触发:1充值总和2首充"`
	RewardType          int         `json:"rewardType"          orm:"reward_type"           description:"奖励类型:1=固定,2=比例"`
	MinRechargeAmount   float64     `json:"minRechargeAmount"   orm:"min_recharge_amount"   description:"触发最小计算金额"`
	InviterRewardAmount float64     `json:"inviterRewardAmount" orm:"inviter_reward_amount" description:"邀请人奖励金额"`
	CreatedAt           *gtime.Time `json:"createdAt"           orm:"created_at"            description:""`
	UpdatedAt           *gtime.Time `json:"updatedAt"           orm:"updated_at"            description:""`
}
