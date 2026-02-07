// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// UserGrade is the golang structure for table user_grade.
type UserGrade struct {
	Id                   uint        `json:"id"                   orm:"id"                     description:""`
	SiteId               int         `json:"siteId"               orm:"site_id"                description:"站点ID"`
	Name                 string      `json:"name"                 orm:"name"                   description:"会员等级名称"`
	RebatePercentSports  float64     `json:"rebatePercentSports"  orm:"rebate_percent_sports"  description:"额外返水比例: 体育"`
	RebatePercentLottery float64     `json:"rebatePercentLottery" orm:"rebate_percent_lottery" description:"额外返水比例: 彩票"`
	RebatePercentLive    float64     `json:"rebatePercentLive"    orm:"rebate_percent_live"    description:"额外返水比例: 真人视讯"`
	RebatePercentEgame   float64     `json:"rebatePercentEgame"   orm:"rebate_percent_egame"   description:"额外返水比例: 电子游戏"`
	RebatePercentPoker   float64     `json:"rebatePercentPoker"   orm:"rebate_percent_poker"   description:"额外返水比例：棋牌"`
	Status               int         `json:"status"               orm:"status"                 description:"状态.1=可用；0=禁用"`
	CreatedAt            *gtime.Time `json:"createdAt"            orm:"created_at"             description:""`
	UpdatedAt            *gtime.Time `json:"updatedAt"            orm:"updated_at"             description:""`
	PaymentLimit         float64     `json:"paymentLimit"         orm:"payment_limit"          description:"充值要求"`
	BetLimit             float64     `json:"betLimit"             orm:"bet_limit"              description:"投注要求"`
	MoneyLimit           float64     `json:"moneyLimit"           orm:"money_limit"            description:"升级送"`
}
