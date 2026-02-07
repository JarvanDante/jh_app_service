// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// User is the golang structure for table user.
type User struct {
	Id               uint        `json:"id"               orm:"id"                 description:""`
	SiteId           int         `json:"siteId"           orm:"site_id"            description:""`
	GradeId          int         `json:"gradeId"          orm:"grade_id"           description:""`
	LevelId          int         `json:"levelId"          orm:"level_id"           description:""`
	AgentId          int         `json:"agentId"          orm:"agent_id"           description:""`
	InviterId        int         `json:"inviterId"        orm:"inviter_id"         description:"邀请人ID"`
	Username         string      `json:"username"         orm:"username"           description:""`
	Password         string      `json:"password"         orm:"password"           description:""`
	PayPassword      string      `json:"payPassword"      orm:"pay_password"       description:""`
	Status           int         `json:"status"           orm:"status"             description:""`
	RegisterIp       string      `json:"registerIp"       orm:"register_ip"        description:""`
	RegisterTime     *gtime.Time `json:"registerTime"     orm:"register_time"      description:""`
	RegisterUrl      string      `json:"registerUrl"      orm:"register_url"       description:"注册来源"`
	LastLoginIp      string      `json:"lastLoginIp"      orm:"last_login_ip"      description:""`
	LastLoginTime    *gtime.Time `json:"lastLoginTime"    orm:"last_login_time"    description:""`
	LastLoginAddress string      `json:"lastLoginAddress" orm:"last_login_address" description:"最后登录地址"`
	Mobile           string      `json:"mobile"           orm:"mobile"             description:""`
	Email            string      `json:"email"            orm:"email"              description:"邮箱"`
	Birthday         *gtime.Time `json:"birthday"         orm:"birthday"           description:""`
	IsOnline         int         `json:"isOnline"         orm:"is_online"          description:""`
	Remark           string      `json:"remark"           orm:"remark"             description:"备注"`
	CreatedAt        *gtime.Time `json:"createdAt"        orm:"created_at"         description:""`
	UpdatedAt        *gtime.Time `json:"updatedAt"        orm:"updated_at"         description:""`
	PayTimes         int         `json:"payTimes"         orm:"pay_times"          description:"充值次数"`
	HasPass          int         `json:"hasPass"          orm:"has_pass"           description:"是否设置密码"`
}
