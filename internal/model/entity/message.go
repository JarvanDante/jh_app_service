// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Message is the golang structure for table message.
type Message struct {
	Id        uint        `json:"id"        orm:"id"         description:""`
	SiteId    int         `json:"siteId"    orm:"site_id"    description:"站点ID"`
	AdminId   int         `json:"adminId"   orm:"admin_id"   description:"员工ID"`
	EventCode string      `json:"eventCode" orm:"event_code" description:"业务触发编码"`
	Title     string      `json:"title"     orm:"title"      description:"消息标题"`
	Content   string      `json:"content"   orm:"content"    description:"消息内容"`
	Status    int         `json:"status"    orm:"status"     description:"状态0=关闭，1=启用"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:""`
}
