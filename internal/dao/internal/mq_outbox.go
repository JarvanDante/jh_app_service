// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MqOutboxDao is the data access object for the table mq_outbox.
type MqOutboxDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  MqOutboxColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// MqOutboxColumns defines and stores column names for the table mq_outbox.
type MqOutboxColumns struct {
	Id            string // 主键ID
	EventType     string // 事件类型，如 user_created / balance_transfer
	AggregateType string // 聚合根类型，如 user / balance
	AggregateId   string // 聚合根ID，如 user_id / trade_no
	SiteCode      string // 站点代码
	Exchange      string // MQ exchange
	RoutingKey    string // MQ routing key
	Payload       string // 发送到 MQ 的消息体
	Status        string // 状态: 0=待发送, 1=已发送, 2=发送失败
	RetryCount    string // 重试次数
	NextRetryAt   string // 下次重试时间
	LastError     string // 最后一次发送错误信息
	CreatedAt     string // 创建时间
	UpdatedAt     string // 更新时间
}

// mqOutboxColumns holds the columns for the table mq_outbox.
var mqOutboxColumns = MqOutboxColumns{
	Id:            "id",
	EventType:     "event_type",
	AggregateType: "aggregate_type",
	AggregateId:   "aggregate_id",
	SiteCode:      "site_code",
	Exchange:      "exchange",
	RoutingKey:    "routing_key",
	Payload:       "payload",
	Status:        "status",
	RetryCount:    "retry_count",
	NextRetryAt:   "next_retry_at",
	LastError:     "last_error",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
}

// NewMqOutboxDao creates and returns a new DAO object for table data access.
func NewMqOutboxDao(handlers ...gdb.ModelHandler) *MqOutboxDao {
	return &MqOutboxDao{
		group:    "default",
		table:    "mq_outbox",
		columns:  mqOutboxColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *MqOutboxDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *MqOutboxDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *MqOutboxDao) Columns() MqOutboxColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *MqOutboxDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *MqOutboxDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *MqOutboxDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
