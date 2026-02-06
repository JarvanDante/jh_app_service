// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SiteActivityDao is the data access object for the table site_activity.
type SiteActivityDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  SiteActivityColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// SiteActivityColumns defines and stores column names for the table site_activity.
type SiteActivityColumns struct {
	Id          string //
	SiteCode    string // 站点CODE
	ActivityId  string // 活动ID
	Name        string // 活动名称
	Code        string // 活动编码(全局唯一)
	Description string // 活动描述
	StartTime   string // 开始时间
	EndTime     string // 结束时间
	FixedParams string // 固定参数(JSON)
	Status      string // 活动是否打开或者或者关闭。1=打开；0=关闭
	Uri         string // 活动uri地址
	Sort        string // 排序
	Table       string // 表名定义不同属性
	CreatedAt   string //
	UpdatedAt   string //
}

// siteActivityColumns holds the columns for the table site_activity.
var siteActivityColumns = SiteActivityColumns{
	Id:          "id",
	SiteCode:    "site_code",
	ActivityId:  "activity_id",
	Name:        "name",
	Code:        "code",
	Description: "description",
	StartTime:   "start_time",
	EndTime:     "end_time",
	FixedParams: "fixed_params",
	Status:      "status",
	Uri:         "uri",
	Sort:        "sort",
	Table:       "table",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
}

// NewSiteActivityDao creates and returns a new DAO object for table data access.
func NewSiteActivityDao(handlers ...gdb.ModelHandler) *SiteActivityDao {
	return &SiteActivityDao{
		group:    "default",
		table:    "site_activity",
		columns:  siteActivityColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SiteActivityDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SiteActivityDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SiteActivityDao) Columns() SiteActivityColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SiteActivityDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SiteActivityDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SiteActivityDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
