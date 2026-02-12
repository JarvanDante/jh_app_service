// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// User is the golang structure of table user for DAO operations like Where/Data.
type User struct {
	g.Meta           `orm:"table:user, do:true"`
	Id               any         //
	SiteId           any         //
	GradeId          any         // 等级id
	AgentId          any         // 代理id
	InviterId        any         // 邀请人ID
	Username         any         //
	Password         any         //
	PayPassword      any         //
	Status           any         //
	RegisterIp       any         //
	RegisterTime     *gtime.Time //
	RegisterUrl      any         // 注册来源
	LastLoginIp      any         // 最后登录ip
	LastLoginTime    *gtime.Time // 最后登录时间
	LastLoginAddress any         // 最后登录地址
	Mobile           any         // 手机
	Email            any         // 邮箱
	Birthday         *gtime.Time //
	IsOnline         any         //
	Remark           any         // 备注
	CreatedAt        *gtime.Time //
	UpdatedAt        *gtime.Time //
	PayTimes         any         // 充值次数
	HasPass          any         // 是否设置密码
}
