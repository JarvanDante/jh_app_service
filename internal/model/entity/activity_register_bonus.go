// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ActivityRegisterBonus is the golang structure for table activity_register_bonus.
type ActivityRegisterBonus struct {
	Id           uint        `json:"id"           orm:"id"            description:""`
	SiteCode     string      `json:"siteCode"     orm:"site_code"     description:"站点CODE"`
	ActivityId   int         `json:"activityId"   orm:"activity_id"   description:"活动ID"`
	RewardAmount float64     `json:"rewardAmount" orm:"reward_amount" description:"注册奖励金额"`
	LimitIp      int         `json:"limitIp"      orm:"limit_ip"      description:"同IP注册限制次数"`
	LimitDevice  int         `json:"limitDevice"  orm:"limit_device"  description:"同设备注册限制次数"`
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"    description:""`
	UpdatedAt    *gtime.Time `json:"updatedAt"    orm:"updated_at"    description:""`
}
