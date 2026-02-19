// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AgentDomainDao is the data access object for the table agent_domain.
type AgentDomainDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  AgentDomainColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// AgentDomainColumns defines and stores column names for the table agent_domain.
type AgentDomainColumns struct {
	Id        string //
	SiteId    string //
	AgentId   string //
	Domain    string // 代理域名/子域名
	Status    string // 1启用 0禁用
	CreatedAt string //
	UpdatedAt string //
}

// agentDomainColumns holds the columns for the table agent_domain.
var agentDomainColumns = AgentDomainColumns{
	Id:        "id",
	SiteId:    "site_id",
	AgentId:   "agent_id",
	Domain:    "domain",
	Status:    "status",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
}

// NewAgentDomainDao creates and returns a new DAO object for table data access.
func NewAgentDomainDao(handlers ...gdb.ModelHandler) *AgentDomainDao {
	return &AgentDomainDao{
		group:    "default",
		table:    "agent_domain",
		columns:  agentDomainColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AgentDomainDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AgentDomainDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AgentDomainDao) Columns() AgentDomainColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AgentDomainDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AgentDomainDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AgentDomainDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
