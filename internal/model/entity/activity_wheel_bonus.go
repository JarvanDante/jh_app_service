// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ActivityWheelBonus is the golang structure for table activity_wheel_bonus.
type ActivityWheelBonus struct {
	Id                uint        `json:"id"                orm:"id"                   description:""`
	SiteCode          string      `json:"siteCode"          orm:"site_code"            description:"站点CODE"`
	ActivityId        int         `json:"activityId"        orm:"activity_id"          description:"活动ID"`
	SegmentNo         uint        `json:"segmentNo"         orm:"segment_no"           description:"转盘扇区序号(1~12)"`
	PrizeName         string      `json:"prizeName"         orm:"prize_name"           description:"奖项名称"`
	RewardAmount      float64     `json:"rewardAmount"      orm:"reward_amount"        description:"奖金额(谢谢参与填0)"`
	Weight            uint        `json:"weight"            orm:"weight"               description:"抽中权重(越大概率越高)"`
	UserDailyWinLimit uint        `json:"userDailyWinLimit" orm:"user_daily_win_limit" description:"单用户每日可中次数(0不限)"`
	MinVipLevel       uint        `json:"minVipLevel"       orm:"min_vip_level"        description:"最低VIP等级(0不限)"`
	Icon              string      `json:"icon"              orm:"icon"                 description:"奖项图标"`
	CreatedAt         *gtime.Time `json:"createdAt"         orm:"created_at"           description:"创建时间"`
	UpdatedAt         *gtime.Time `json:"updatedAt"         orm:"updated_at"           description:"更新时间"`
}
