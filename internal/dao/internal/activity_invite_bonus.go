// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ActivityInviteBonusDao is the data access object for the table activity_invite_bonus.
type ActivityInviteBonusDao struct {
	table    string                     // table is the underlying table name of the DAO.
	group    string                     // group is the database configuration group name of the current DAO.
	columns  ActivityInviteBonusColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler         // handlers for customized model modification.
}

// ActivityInviteBonusColumns defines and stores column names for the table activity_invite_bonus.
type ActivityInviteBonusColumns struct {
	Id                  string //
	SiteCode            string // 站点CODE
	ActivityId          string // 活动ID
	TriggerEvent        string // 发放触发:1充值总和2首充
	RewardType          string // 奖励类型:1=固定,2=比例
	MinRechargeAmount   string // 触发最小计算金额
	InviterRewardAmount string // 邀请人奖励金额
	CreatedAt           string //
	UpdatedAt           string //
}

// activityInviteBonusColumns holds the columns for the table activity_invite_bonus.
var activityInviteBonusColumns = ActivityInviteBonusColumns{
	Id:                  "id",
	SiteCode:            "site_code",
	ActivityId:          "activity_id",
	TriggerEvent:        "trigger_event",
	RewardType:          "reward_type",
	MinRechargeAmount:   "min_recharge_amount",
	InviterRewardAmount: "inviter_reward_amount",
	CreatedAt:           "created_at",
	UpdatedAt:           "updated_at",
}

// NewActivityInviteBonusDao creates and returns a new DAO object for table data access.
func NewActivityInviteBonusDao(handlers ...gdb.ModelHandler) *ActivityInviteBonusDao {
	return &ActivityInviteBonusDao{
		group:    "default",
		table:    "activity_invite_bonus",
		columns:  activityInviteBonusColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ActivityInviteBonusDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ActivityInviteBonusDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ActivityInviteBonusDao) Columns() ActivityInviteBonusColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ActivityInviteBonusDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ActivityInviteBonusDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ActivityInviteBonusDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
