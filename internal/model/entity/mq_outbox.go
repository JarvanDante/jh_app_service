// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MqOutbox is the golang structure for table mq_outbox.
type MqOutbox struct {
	Id            uint64      `json:"id"            orm:"id"             description:"主键ID"`
	EventType     string      `json:"eventType"     orm:"event_type"     description:"事件类型，如 user_created / balance_transfer"`
	AggregateType string      `json:"aggregateType" orm:"aggregate_type" description:"聚合根类型，如 user / balance"`
	AggregateId   string      `json:"aggregateId"   orm:"aggregate_id"   description:"聚合根ID，如 user_id / trade_no"`
	SiteCode      string      `json:"siteCode"      orm:"site_code"      description:"站点代码"`
	Exchange      string      `json:"exchange"      orm:"exchange"       description:"MQ exchange"`
	RoutingKey    string      `json:"routingKey"    orm:"routing_key"    description:"MQ routing key"`
	Payload       string      `json:"payload"       orm:"payload"        description:"发送到 MQ 的消息体"`
	Status        int         `json:"status"        orm:"status"         description:"状态: 0=待发送, 1=已发送, 2=发送失败"`
	RetryCount    int         `json:"retryCount"    orm:"retry_count"    description:"重试次数"`
	NextRetryAt   *gtime.Time `json:"nextRetryAt"   orm:"next_retry_at"  description:"下次重试时间"`
	LastError     string      `json:"lastError"     orm:"last_error"     description:"最后一次发送错误信息"`
	CreatedAt     *gtime.Time `json:"createdAt"     orm:"created_at"     description:"创建时间"`
	UpdatedAt     *gtime.Time `json:"updatedAt"     orm:"updated_at"     description:"更新时间"`
}
