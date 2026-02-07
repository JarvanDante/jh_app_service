// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UserGradeDao is the data access object for the table user_grade.
type UserGradeDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  UserGradeColumns   // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// UserGradeColumns defines and stores column names for the table user_grade.
type UserGradeColumns struct {
	Id                   string //
	SiteId               string // 站点ID
	Name                 string // 会员等级名称
	RebatePercentSports  string // 额外返水比例: 体育
	RebatePercentLottery string // 额外返水比例: 彩票
	RebatePercentLive    string // 额外返水比例: 真人视讯
	RebatePercentEgame   string // 额外返水比例: 电子游戏
	RebatePercentPoker   string // 额外返水比例：棋牌
	Status               string // 状态.1=可用；0=禁用
	CreatedAt            string //
	UpdatedAt            string //
	PaymentLimit         string // 充值要求
	BetLimit             string // 投注要求
	MoneyLimit           string // 升级送
}

// userGradeColumns holds the columns for the table user_grade.
var userGradeColumns = UserGradeColumns{
	Id:                   "id",
	SiteId:               "site_id",
	Name:                 "name",
	RebatePercentSports:  "rebate_percent_sports",
	RebatePercentLottery: "rebate_percent_lottery",
	RebatePercentLive:    "rebate_percent_live",
	RebatePercentEgame:   "rebate_percent_egame",
	RebatePercentPoker:   "rebate_percent_poker",
	Status:               "status",
	CreatedAt:            "created_at",
	UpdatedAt:            "updated_at",
	PaymentLimit:         "payment_limit",
	BetLimit:             "bet_limit",
	MoneyLimit:           "money_limit",
}

// NewUserGradeDao creates and returns a new DAO object for table data access.
func NewUserGradeDao(handlers ...gdb.ModelHandler) *UserGradeDao {
	return &UserGradeDao{
		group:    "default",
		table:    "user_grade",
		columns:  userGradeColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *UserGradeDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *UserGradeDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *UserGradeDao) Columns() UserGradeColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *UserGradeDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *UserGradeDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *UserGradeDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
