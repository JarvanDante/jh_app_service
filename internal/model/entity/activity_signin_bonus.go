// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ActivitySigninBonus is the golang structure for table activity_signin_bonus.
type ActivitySigninBonus struct {
	Id           uint        `json:"id"           orm:"id"            description:""`
	SiteCode     string      `json:"siteCode"     orm:"site_code"     description:"站点CODE"`
	ActivityId   int         `json:"activityId"   orm:"activity_id"   description:"活动ID"`
	DayNo        int         `json:"dayNo"        orm:"day_no"        description:"日号(1-31)"`
	RewardAmount float64     `json:"rewardAmount" orm:"reward_amount" description:"当日固定金额"`
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"    description:"创建时间"`
	UpdatedAt    *gtime.Time `json:"updatedAt"    orm:"updated_at"    description:"更新时间"`
}
