// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// AgentDomain is the golang structure of table agent_domain for DAO operations like Where/Data.
type AgentDomain struct {
	g.Meta    `orm:"table:agent_domain, do:true"`
	Id        any         //
	SiteId    any         //
	AgentId   any         //
	Domain    any         // 代理域名/子域名
	Status    any         // 1启用 0禁用
	CreatedAt *gtime.Time //
	UpdatedAt *gtime.Time //
}
