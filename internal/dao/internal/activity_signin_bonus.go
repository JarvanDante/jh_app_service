// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ActivitySigninBonusDao is the data access object for the table activity_signin_bonus.
type ActivitySigninBonusDao struct {
	table    string                     // table is the underlying table name of the DAO.
	group    string                     // group is the database configuration group name of the current DAO.
	columns  ActivitySigninBonusColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler         // handlers for customized model modification.
}

// ActivitySigninBonusColumns defines and stores column names for the table activity_signin_bonus.
type ActivitySigninBonusColumns struct {
	Id           string //
	SiteCode     string // 站点CODE
	ActivityId   string // 活动ID
	DayNo        string // 日号(1-31)
	RewardAmount string // 当日固定金额
	CreatedAt    string // 创建时间
	UpdatedAt    string // 更新时间
}

// activitySigninBonusColumns holds the columns for the table activity_signin_bonus.
var activitySigninBonusColumns = ActivitySigninBonusColumns{
	Id:           "id",
	SiteCode:     "site_code",
	ActivityId:   "activity_id",
	DayNo:        "day_no",
	RewardAmount: "reward_amount",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
}

// NewActivitySigninBonusDao creates and returns a new DAO object for table data access.
func NewActivitySigninBonusDao(handlers ...gdb.ModelHandler) *ActivitySigninBonusDao {
	return &ActivitySigninBonusDao{
		group:    "default",
		table:    "activity_signin_bonus",
		columns:  activitySigninBonusColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ActivitySigninBonusDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ActivitySigninBonusDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ActivitySigninBonusDao) Columns() ActivitySigninBonusColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ActivitySigninBonusDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ActivitySigninBonusDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ActivitySigninBonusDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
