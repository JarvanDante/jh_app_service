// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ActivityRegisterBonusDao is the data access object for the table activity_register_bonus.
type ActivityRegisterBonusDao struct {
	table    string                       // table is the underlying table name of the DAO.
	group    string                       // group is the database configuration group name of the current DAO.
	columns  ActivityRegisterBonusColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler           // handlers for customized model modification.
}

// ActivityRegisterBonusColumns defines and stores column names for the table activity_register_bonus.
type ActivityRegisterBonusColumns struct {
	Id           string //
	SiteCode     string // 站点CODE
	ActivityId   string // 活动ID
	RewardAmount string // 注册奖励金额
	LimitIp      string // 同IP注册限制次数
	LimitDevice  string // 同设备注册限制次数
	CreatedAt    string //
	UpdatedAt    string //
}

// activityRegisterBonusColumns holds the columns for the table activity_register_bonus.
var activityRegisterBonusColumns = ActivityRegisterBonusColumns{
	Id:           "id",
	SiteCode:     "site_code",
	ActivityId:   "activity_id",
	RewardAmount: "reward_amount",
	LimitIp:      "limit_ip",
	LimitDevice:  "limit_device",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
}

// NewActivityRegisterBonusDao creates and returns a new DAO object for table data access.
func NewActivityRegisterBonusDao(handlers ...gdb.ModelHandler) *ActivityRegisterBonusDao {
	return &ActivityRegisterBonusDao{
		group:    "default",
		table:    "activity_register_bonus",
		columns:  activityRegisterBonusColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ActivityRegisterBonusDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ActivityRegisterBonusDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ActivityRegisterBonusDao) Columns() ActivityRegisterBonusColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ActivityRegisterBonusDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ActivityRegisterBonusDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ActivityRegisterBonusDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
