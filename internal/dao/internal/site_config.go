// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SiteConfigDao is the data access object for the table site_config.
type SiteConfigDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SiteConfigColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SiteConfigColumns defines and stores column names for the table site_config.
type SiteConfigColumns struct {
	Id                   string //
	SiteId               string // 站点ID
	SwitchRegister       string // 是否开放注册。1=开放;0=关闭
	SwitchGrade          string // 是否开放会员等级。1=开放;0=关闭
	RegisterTimeInterval string // 同一IP重复注册。设定时间内，同一IP将无法进行多次注册。0或留空表示不限制
	DefaultGradeId       string // 默认用户等级ID
	DefaultLevelId       string // 默认用户层级ID
	DefaultAgentId       string // 默认代理ID
	IsClose              string // 是否关站。1=是；0=否
	MobileLogo           string // 手机端logo
	MinWithdraw          string // 单笔最低提现金额
	MaxWithdraw          string // 单笔最高提现金额
	CloseReason          string // 关站原因
	UrlService           string // 客服链接
	UrlTg                string // telegram链接
	UrlMobile            string // 手机域名
	WebUrl               string // 网站主域名
	CreatedAt            string //
	UpdatedAt            string //
	SwitchSign           string //
}

// siteConfigColumns holds the columns for the table site_config.
var siteConfigColumns = SiteConfigColumns{
	Id:                   "id",
	SiteId:               "site_id",
	SwitchRegister:       "switch_register",
	SwitchGrade:          "switch_grade",
	RegisterTimeInterval: "register_time_interval",
	DefaultGradeId:       "default_grade_id",
	DefaultLevelId:       "default_level_id",
	DefaultAgentId:       "default_agent_id",
	IsClose:              "is_close",
	MobileLogo:           "mobile_logo",
	MinWithdraw:          "min_withdraw",
	MaxWithdraw:          "max_withdraw",
	CloseReason:          "close_reason",
	UrlService:           "url_service",
	UrlTg:                "url_tg",
	UrlMobile:            "url_mobile",
	WebUrl:               "web_url",
	CreatedAt:            "created_at",
	UpdatedAt:            "updated_at",
	SwitchSign:           "switch_sign",
}

// NewSiteConfigDao creates and returns a new DAO object for table data access.
func NewSiteConfigDao(handlers ...gdb.ModelHandler) *SiteConfigDao {
	return &SiteConfigDao{
		group:    "default",
		table:    "site_config",
		columns:  siteConfigColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SiteConfigDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SiteConfigDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SiteConfigDao) Columns() SiteConfigColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SiteConfigDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SiteConfigDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SiteConfigDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
