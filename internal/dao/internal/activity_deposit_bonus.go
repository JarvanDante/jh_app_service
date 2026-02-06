// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ActivityDepositBonusDao is the data access object for the table activity_deposit_bonus.
type ActivityDepositBonusDao struct {
	table    string                      // table is the underlying table name of the DAO.
	group    string                      // group is the database configuration group name of the current DAO.
	columns  ActivityDepositBonusColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler          // handlers for customized model modification.
}

// ActivityDepositBonusColumns defines and stores column names for the table activity_deposit_bonus.
type ActivityDepositBonusColumns struct {
	Id                string //
	SiteCode          string // 站点CODE
	ActivityId        string // 活动ID
	TriggerEvent      string // 发放触发:1=充值总和,2=首充
	RewardType        string // 奖励类型:1=固定,2=比例
	MinRechargeAmount string // 触发最小充值金额
	RewardAmount      string // 奖励金额/比例值
	CreatedAt         string //
	UpdatedAt         string //
}

// activityDepositBonusColumns holds the columns for the table activity_deposit_bonus.
var activityDepositBonusColumns = ActivityDepositBonusColumns{
	Id:                "id",
	SiteCode:          "site_code",
	ActivityId:        "activity_id",
	TriggerEvent:      "trigger_event",
	RewardType:        "reward_type",
	MinRechargeAmount: "min_recharge_amount",
	RewardAmount:      "reward_amount",
	CreatedAt:         "created_at",
	UpdatedAt:         "updated_at",
}

// NewActivityDepositBonusDao creates and returns a new DAO object for table data access.
func NewActivityDepositBonusDao(handlers ...gdb.ModelHandler) *ActivityDepositBonusDao {
	return &ActivityDepositBonusDao{
		group:    "default",
		table:    "activity_deposit_bonus",
		columns:  activityDepositBonusColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ActivityDepositBonusDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ActivityDepositBonusDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ActivityDepositBonusDao) Columns() ActivityDepositBonusColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ActivityDepositBonusDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ActivityDepositBonusDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ActivityDepositBonusDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
