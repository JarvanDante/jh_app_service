// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// MqOutbox is the golang structure of table mq_outbox for DAO operations like Where/Data.
type MqOutbox struct {
	g.Meta        `orm:"table:mq_outbox, do:true"`
	Id            any         // 主键ID
	EventType     any         // 事件类型，如 user_created / balance_transfer
	AggregateType any         // 聚合根类型，如 user / balance
	AggregateId   any         // 聚合根ID，如 user_id / trade_no
	SiteCode      any         // 站点代码
	Exchange      any         // MQ exchange
	RoutingKey    any         // MQ routing key
	Payload       any         // 发送到 MQ 的消息体
	Status        any         // 状态: 0=待发送, 1=已发送, 2=发送失败
	RetryCount    any         // 重试次数
	NextRetryAt   *gtime.Time // 下次重试时间
	LastError     any         // 最后一次发送错误信息
	CreatedAt     *gtime.Time // 创建时间
	UpdatedAt     *gtime.Time // 更新时间
}
