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
	Id            interface{} // 主键ID
	EventType     interface{} // 事件类型，如 user_created / balance_transfer
	AggregateType interface{} // 聚合根类型，如 user / balance
	AggregateId   interface{} // 聚合根ID，如 user_id / trade_no
	SiteCode      interface{} // 站点代码
	Exchange      interface{} // MQ exchange
	RoutingKey    interface{} // MQ routing key
	Payload       interface{} // 发送到 MQ 的消息体
	Status        interface{} // 状态: 0=待发送, 1=已发送, 2=发送失败
	RetryCount    interface{} // 重试次数
	NextRetryAt   *gtime.Time // 下次重试时间
	LastError     interface{} // 最后一次发送错误信息
	CreatedAt     *gtime.Time // 创建时间
	UpdatedAt     *gtime.Time // 更新时间
}
