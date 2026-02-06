// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// SiteActivity is the golang structure of table site_activity for DAO operations like Where/Data.
type SiteActivity struct {
	g.Meta      `orm:"table:site_activity, do:true"`
	Id          any         //
	SiteCode    any         // 站点CODE
	ActivityId  any         // 活动ID
	Name        any         // 活动名称
	Code        any         // 活动编码(全局唯一)
	Description any         // 活动描述
	StartTime   *gtime.Time // 开始时间
	EndTime     *gtime.Time // 结束时间
	FixedParams any         // 固定参数(JSON)
	Status      any         // 活动是否打开或者或者关闭。1=打开；0=关闭
	Uri         any         // 活动uri地址
	Sort        any         // 排序
	Table       any         // 表名定义不同属性
	CreatedAt   *gtime.Time //
	UpdatedAt   *gtime.Time //
}
