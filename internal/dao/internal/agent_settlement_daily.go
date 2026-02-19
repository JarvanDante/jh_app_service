// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AgentSettlementDailyDao is the data access object for the table agent_settlement_daily.
type AgentSettlementDailyDao struct {
	table    string                      // table is the underlying table name of the DAO.
	group    string                      // group is the database configuration group name of the current DAO.
	columns  AgentSettlementDailyColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler          // handlers for customized model modification.
}

// AgentSettlementDailyColumns defines and stores column names for the table agent_settlement_daily.
type AgentSettlementDailyColumns struct {
	Id               string //
	SiteId           string //
	AgentId          string //
	StatDate         string //
	ValidBetAmount   string //
	WinLoss          string //
	CommissionAmount string //
	Status           string // 0待结算 1已结算
	SettledAt        string //
	CreatedAt        string //
	UpdatedAt        string //
}

// agentSettlementDailyColumns holds the columns for the table agent_settlement_daily.
var agentSettlementDailyColumns = AgentSettlementDailyColumns{
	Id:               "id",
	SiteId:           "site_id",
	AgentId:          "agent_id",
	StatDate:         "stat_date",
	ValidBetAmount:   "valid_bet_amount",
	WinLoss:          "win_loss",
	CommissionAmount: "commission_amount",
	Status:           "status",
	SettledAt:        "settled_at",
	CreatedAt:        "created_at",
	UpdatedAt:        "updated_at",
}

// NewAgentSettlementDailyDao creates and returns a new DAO object for table data access.
func NewAgentSettlementDailyDao(handlers ...gdb.ModelHandler) *AgentSettlementDailyDao {
	return &AgentSettlementDailyDao{
		group:    "default",
		table:    "agent_settlement_daily",
		columns:  agentSettlementDailyColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AgentSettlementDailyDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AgentSettlementDailyDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AgentSettlementDailyDao) Columns() AgentSettlementDailyColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AgentSettlementDailyDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AgentSettlementDailyDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AgentSettlementDailyDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
