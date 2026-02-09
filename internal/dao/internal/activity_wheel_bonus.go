// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ActivityWheelBonusDao is the data access object for the table activity_wheel_bonus.
type ActivityWheelBonusDao struct {
	table    string                    // table is the underlying table name of the DAO.
	group    string                    // group is the database configuration group name of the current DAO.
	columns  ActivityWheelBonusColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler        // handlers for customized model modification.
}

// ActivityWheelBonusColumns defines and stores column names for the table activity_wheel_bonus.
type ActivityWheelBonusColumns struct {
	Id                string //
	SiteCode          string // 站点CODE
	ActivityId        string // 活动ID
	SegmentNo         string // 转盘扇区序号(1~12)
	PrizeName         string // 奖项名称
	RewardAmount      string // 奖金额(谢谢参与填0)
	Weight            string // 抽中权重(越大概率越高)
	UserDailyWinLimit string // 单用户每日可中次数(0不限)
	MinVipLevel       string // 最低VIP等级(0不限)
	Icon              string // 奖项图标
	CreatedAt         string // 创建时间
	UpdatedAt         string // 更新时间
}

// activityWheelBonusColumns holds the columns for the table activity_wheel_bonus.
var activityWheelBonusColumns = ActivityWheelBonusColumns{
	Id:                "id",
	SiteCode:          "site_code",
	ActivityId:        "activity_id",
	SegmentNo:         "segment_no",
	PrizeName:         "prize_name",
	RewardAmount:      "reward_amount",
	Weight:            "weight",
	UserDailyWinLimit: "user_daily_win_limit",
	MinVipLevel:       "min_vip_level",
	Icon:              "icon",
	CreatedAt:         "created_at",
	UpdatedAt:         "updated_at",
}

// NewActivityWheelBonusDao creates and returns a new DAO object for table data access.
func NewActivityWheelBonusDao(handlers ...gdb.ModelHandler) *ActivityWheelBonusDao {
	return &ActivityWheelBonusDao{
		group:    "default",
		table:    "activity_wheel_bonus",
		columns:  activityWheelBonusColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ActivityWheelBonusDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ActivityWheelBonusDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ActivityWheelBonusDao) Columns() ActivityWheelBonusColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ActivityWheelBonusDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ActivityWheelBonusDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ActivityWheelBonusDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
