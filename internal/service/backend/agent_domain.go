package backend

import (
	"context"
	v1 "jh_app_service/api/backend/agent_domain/v1"
)

type IAgentDomain interface {
	GetAgentDomainList(ctx context.Context, req *v1.GetAgentDomainListReq) (*v1.GetAgentDomainListRes, error)
	CreateAgentDomain(ctx context.Context, req *v1.CreateAgentDomainReq) (*v1.CreateAgentDomainRes, error)
	UpdateAgentDomain(ctx context.Context, req *v1.UpdateAgentDomainReq) (*v1.UpdateAgentDomainRes, error)
	DeleteAgentDomain(ctx context.Context, req *v1.DeleteAgentDomainReq) (*v1.DeleteAgentDomainRes, error)
}

var localAgentDomain IAgentDomain

func AgentDomain() IAgentDomain {
	if localAgentDomain == nil {
		panic("implement not found for interface IAgentDomain, forgot register?")
	}
	return localAgentDomain
}

func RegisterAgentDomain(i IAgentDomain) {
	localAgentDomain = i
}
