// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AgentDao is the data access object for the table agent.
type AgentDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  AgentColumns       // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// AgentColumns defines and stores column names for the table agent.
type AgentColumns struct {
	Id         string //
	SiteId     string // 站点ID
	ParentId   string // 上级代理ID，0=顶级
	AgentCode  string // 代理编码(站点内唯一)
	AgentName  string // 代理名称
	Status     string // 1启用 0禁用
	SettleType string // 1亏损分成 2流水分成
	Rate       string // 默认佣金比例
	Remark     string // 备注
	CreatedAt  string //
	UpdatedAt  string //
}

// agentColumns holds the columns for the table agent.
var agentColumns = AgentColumns{
	Id:         "id",
	SiteId:     "site_id",
	ParentId:   "parent_id",
	AgentCode:  "agent_code",
	AgentName:  "agent_name",
	Status:     "status",
	SettleType: "settle_type",
	Rate:       "rate",
	Remark:     "remark",
	CreatedAt:  "created_at",
	UpdatedAt:  "updated_at",
}

// NewAgentDao creates and returns a new DAO object for table data access.
func NewAgentDao(handlers ...gdb.ModelHandler) *AgentDao {
	return &AgentDao{
		group:    "default",
		table:    "agent",
		columns:  agentColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AgentDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AgentDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AgentDao) Columns() AgentColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AgentDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AgentDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AgentDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
