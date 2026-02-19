// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Agent is the golang structure of table agent for DAO operations like Where/Data.
type Agent struct {
	g.Meta     `orm:"table:agent, do:true"`
	Id         any         //
	SiteId     any         // 站点ID
	ParentId   any         // 上级代理ID，0=顶级
	AgentCode  any         // 代理编码(站点内唯一)
	AgentName  any         // 代理名称
	Status     any         // 1启用 0禁用
	SettleType any         // 1亏损分成 2流水分成
	Rate       any         // 默认佣金比例
	Remark     any         // 备注
	CreatedAt  *gtime.Time //
	UpdatedAt  *gtime.Time //
}
