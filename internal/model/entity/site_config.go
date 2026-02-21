// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// SiteConfig is the golang structure for table site_config.
type SiteConfig struct {
	Id                   uint        `json:"id"                   orm:"id"                     description:""`
	SiteId               int         `json:"siteId"               orm:"site_id"                description:"站点ID"`
	SwitchRegister       int         `json:"switchRegister"       orm:"switch_register"        description:"是否开放注册。1=开放;0=关闭"`
	SwitchGrade          int         `json:"switchGrade"          orm:"switch_grade"           description:"是否开放会员等级。1=开放;0=关闭"`
	RegisterTimeInterval int         `json:"registerTimeInterval" orm:"register_time_interval" description:"同一IP重复注册。设定时间内，同一IP将无法进行多次注册。0或留空表示不限制"`
	DefaultGradeId       int         `json:"defaultGradeId"       orm:"default_grade_id"       description:"默认用户等级ID"`
	DefaultLevelId       int         `json:"defaultLevelId"       orm:"default_level_id"       description:"默认用户层级ID"`
	DefaultAgentId       int         `json:"defaultAgentId"       orm:"default_agent_id"       description:"默认代理ID"`
	IsClose              int         `json:"isClose"              orm:"is_close"               description:"是否关站。1=是；0=否"`
	MobileLogo           string      `json:"mobileLogo"           orm:"mobile_logo"            description:"手机端logo"`
	MinWithdraw          int         `json:"minWithdraw"          orm:"min_withdraw"           description:"单笔最低提现金额"`
	MaxWithdraw          int         `json:"maxWithdraw"          orm:"max_withdraw"           description:"单笔最高提现金额"`
	CloseReason          string      `json:"closeReason"          orm:"close_reason"           description:"关站原因"`
	UrlService           string      `json:"urlService"           orm:"url_service"            description:"客服链接"`
	UrlTg                string      `json:"urlTg"                orm:"url_tg"                 description:"telegram链接"`
	UrlMobile            string      `json:"urlMobile"            orm:"url_mobile"             description:"手机域名"`
	WebUrl               string      `json:"webUrl"               orm:"web_url"                description:"网站主域名"`
	CreatedAt            *gtime.Time `json:"createdAt"            orm:"created_at"             description:""`
	UpdatedAt            *gtime.Time `json:"updatedAt"            orm:"updated_at"             description:""`
	SwitchSign           string      `json:"switchSign"           orm:"switch_sign"            description:""`
}
