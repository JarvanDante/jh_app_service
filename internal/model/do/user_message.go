// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// UserMessage is the golang structure of table user_message for DAO operations like Where/Data.
type UserMessage struct {
	g.Meta    `orm:"table:user_message, do:true"`
	Id        any         //
	SiteId    any         // 站点ID
	UserId    any         // 用户ID
	MessageId any         // 消息模板ID
	EventCode any         // 业务触发编码(冗余，便于查询)
	BizId     any         // 业务流水号(充值单号/升级记录等)
	Amount    any         // 涉及的金额
	GradeName any         // 会员等级名称
	Title     any         // 消息标题(快照)
	Content   any         // 消息内容(快照)
	Status    any         // 0=未读 1=已读
	ReadAt    *gtime.Time // 阅读时间
	DeleteAt  *gtime.Time // 删除时间(软删)
	CreatedAt *gtime.Time //
	UpdatedAt *gtime.Time //
}
