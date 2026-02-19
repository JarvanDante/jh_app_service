// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// AgentSettlementDaily is the golang structure of table agent_settlement_daily for DAO operations like Where/Data.
type AgentSettlementDaily struct {
	g.Meta           `orm:"table:agent_settlement_daily, do:true"`
	Id               any         //
	SiteId           any         //
	AgentId          any         //
	StatDate         *gtime.Time //
	ValidBetAmount   any         //
	WinLoss          any         //
	CommissionAmount any         //
	Status           any         // 0待结算 1已结算
	SettledAt        *gtime.Time //
	CreatedAt        *gtime.Time //
	UpdatedAt        *gtime.Time //
}
