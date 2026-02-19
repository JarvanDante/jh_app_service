// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// AgentSettlementDaily is the golang structure for table agent_settlement_daily.
type AgentSettlementDaily struct {
	Id               int64       `json:"id"               orm:"id"                description:""`
	SiteId           int64       `json:"siteId"           orm:"site_id"           description:""`
	AgentId          int64       `json:"agentId"          orm:"agent_id"          description:""`
	StatDate         *gtime.Time `json:"statDate"         orm:"stat_date"         description:""`
	BetAmount        float64     `json:"betAmount"        orm:"bet_amount"        description:""`
	ValidBetAmount   float64     `json:"validBetAmount"   orm:"valid_bet_amount"  description:""`
	WinLoss          float64     `json:"winLoss"          orm:"win_loss"          description:""`
	CommissionAmount float64     `json:"commissionAmount" orm:"commission_amount" description:""`
	Status           int         `json:"status"           orm:"status"            description:"0待结算 1已结算"`
	SettledAt        *gtime.Time `json:"settledAt"        orm:"settled_at"        description:""`
	CreatedAt        *gtime.Time `json:"createdAt"        orm:"created_at"        description:""`
	UpdatedAt        *gtime.Time `json:"updatedAt"        orm:"updated_at"        description:""`
}
