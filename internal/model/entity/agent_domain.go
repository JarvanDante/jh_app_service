// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// AgentDomain is the golang structure for table agent_domain.
type AgentDomain struct {
	Id        int64       `json:"id"        orm:"id"         description:""`
	SiteId    int64       `json:"siteId"    orm:"site_id"    description:""`
	AgentId   int64       `json:"agentId"   orm:"agent_id"   description:""`
	Domain    string      `json:"domain"    orm:"domain"     description:"代理域名/子域名"`
	Status    int         `json:"status"    orm:"status"     description:"1启用 0禁用"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:""`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:""`
}
