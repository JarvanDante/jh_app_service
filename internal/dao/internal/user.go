// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UserDao is the data access object for the table user.
type UserDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  UserColumns        // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// UserColumns defines and stores column names for the table user.
type UserColumns struct {
	Id               string //
	SiteId           string //
	GradeId          string //
	LevelId          string //
	AgentId          string //
	InviterId        string // 邀请人ID
	Username         string //
	Password         string //
	PayPassword      string //
	Status           string //
	RegisterIp       string //
	RegisterTime     string //
	RegisterUrl      string // 注册来源
	LastLoginIp      string //
	LastLoginTime    string //
	LastLoginAddress string // 最后登录地址
	Mobile           string //
	Email            string // 邮箱
	Birthday         string //
	IsOnline         string //
	Remark           string // 备注
	CreatedAt        string //
	UpdatedAt        string //
	PayTimes         string // 充值次数
	HasPass          string // 是否设置密码
}

// userColumns holds the columns for the table user.
var userColumns = UserColumns{
	Id:               "id",
	SiteId:           "site_id",
	GradeId:          "grade_id",
	LevelId:          "level_id",
	AgentId:          "agent_id",
	InviterId:        "inviter_id",
	Username:         "username",
	Password:         "password",
	PayPassword:      "pay_password",
	Status:           "status",
	RegisterIp:       "register_ip",
	RegisterTime:     "register_time",
	RegisterUrl:      "register_url",
	LastLoginIp:      "last_login_ip",
	LastLoginTime:    "last_login_time",
	LastLoginAddress: "last_login_address",
	Mobile:           "mobile",
	Email:            "email",
	Birthday:         "birthday",
	IsOnline:         "is_online",
	Remark:           "remark",
	CreatedAt:        "created_at",
	UpdatedAt:        "updated_at",
	PayTimes:         "pay_times",
	HasPass:          "has_pass",
}

// NewUserDao creates and returns a new DAO object for table data access.
func NewUserDao(handlers ...gdb.ModelHandler) *UserDao {
	return &UserDao{
		group:    "default",
		table:    "user",
		columns:  userColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *UserDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *UserDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *UserDao) Columns() UserColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *UserDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *UserDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *UserDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
