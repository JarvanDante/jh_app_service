// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Agent is the golang structure for table agent.
type Agent struct {
	Id         int64       `json:"id"         orm:"id"          description:""`
	SiteId     int64       `json:"siteId"     orm:"site_id"     description:"站点ID"`
	ParentId   int64       `json:"parentId"   orm:"parent_id"   description:"上级代理ID，0=顶级"`
	AgentCode  string      `json:"agentCode"  orm:"agent_code"  description:"代理编码(站点内唯一)"`
	AgentName  string      `json:"agentName"  orm:"agent_name"  description:"代理名称"`
	Status     int         `json:"status"     orm:"status"      description:"1启用 0禁用"`
	SettleType int         `json:"settleType" orm:"settle_type" description:"1亏损分成 2流水分成"`
	Rate       float64     `json:"rate"       orm:"rate"        description:"默认佣金比例"`
	Remark     string      `json:"remark"     orm:"remark"      description:"备注"`
	CreatedAt  *gtime.Time `json:"createdAt"  orm:"created_at"  description:""`
	UpdatedAt  *gtime.Time `json:"updatedAt"  orm:"updated_at"  description:""`
}
