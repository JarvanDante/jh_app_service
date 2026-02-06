// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SiteActivity is the golang structure for table site_activity.
type SiteActivity struct {
	Id          uint        `json:"id"          orm:"id"           description:""`
	SiteCode    string      `json:"siteCode"    orm:"site_code"    description:"站点CODE"`
	ActivityId  int         `json:"activityId"  orm:"activity_id"  description:"活动ID"`
	Name        string      `json:"name"        orm:"name"         description:"活动名称"`
	Code        string      `json:"code"        orm:"code"         description:"活动编码(全局唯一)"`
	Description string      `json:"description" orm:"description"  description:"活动描述"`
	StartTime   *gtime.Time `json:"startTime"   orm:"start_time"   description:"开始时间"`
	EndTime     *gtime.Time `json:"endTime"     orm:"end_time"     description:"结束时间"`
	FixedParams string      `json:"fixedParams" orm:"fixed_params" description:"固定参数(JSON)"`
	Status      int         `json:"status"      orm:"status"       description:"活动是否打开或者或者关闭。1=打开；0=关闭"`
	Sort        int         `json:"sort"        orm:"sort"         description:"排序"`
	Table       string      `json:"table"       orm:"table"        description:"表名定义不同属性"`
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"   description:""`
	UpdatedAt   *gtime.Time `json:"updatedAt"   orm:"updated_at"   description:""`
}
