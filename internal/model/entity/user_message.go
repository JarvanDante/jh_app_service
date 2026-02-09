// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// UserMessage is the golang structure for table user_message.
type UserMessage struct {
	Id        uint64      `json:"id"        orm:"id"         description:""`
	SiteId    int         `json:"siteId"    orm:"site_id"    description:"站点ID"`
	UserId    int         `json:"userId"    orm:"user_id"    description:"用户ID"`
	MessageId int         `json:"messageId" orm:"message_id" description:"消息模板ID"`
	EventCode string      `json:"eventCode" orm:"event_code" description:"业务触发编码(冗余，便于查询)"`
	BizId     string      `json:"bizId"     orm:"biz_id"     description:"业务流水号(充值单号/升级记录等)"`
	Amount    float64     `json:"amount"    orm:"amount"     description:"涉及的金额"`
	GradeName string      `json:"gradeName" orm:"grade_name" description:"会员等级名称"`
	Title     string      `json:"title"     orm:"title"      description:"消息标题(快照)"`
	Content   string      `json:"content"   orm:"content"    description:"消息内容(快照)"`
	Status    int         `json:"status"    orm:"status"     description:"0=未读 1=已读"`
	ReadAt    *gtime.Time `json:"readAt"    orm:"read_at"    description:"阅读时间"`
	DeleteAt  *gtime.Time `json:"deleteAt"  orm:"delete_at"  description:"删除时间(软删)"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:""`
}
