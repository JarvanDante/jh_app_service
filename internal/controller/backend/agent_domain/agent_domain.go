package agent_domain

import (
	"context"
	v1 "jh_app_service/api/backend/agent_domain/v1"
	"jh_app_service/internal/service/backend"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

type Controller struct {
	v1.UnimplementedAgentDomainServer
}

func Register(s *grpcx.GrpcServer) {
	v1.RegisterAgentDomainServer(s.Server, &Controller{})
}

func (*Controller) GetAgentDomainList(ctx context.Context, req *v1.GetAgentDomainListReq) (*v1.GetAgentDomainListRes, error) {
	return backend.AgentDomain().GetAgentDomainList(ctx, req)
}

func (*Controller) CreateAgentDomain(ctx context.Context, req *v1.CreateAgentDomainReq) (*v1.CreateAgentDomainRes, error) {
	return backend.AgentDomain().CreateAgentDomain(ctx, req)
}

func (*Controller) UpdateAgentDomain(ctx context.Context, req *v1.UpdateAgentDomainReq) (*v1.UpdateAgentDomainRes, error) {
	return backend.AgentDomain().UpdateAgentDomain(ctx, req)
}

func (*Controller) DeleteAgentDomain(ctx context.Context, req *v1.DeleteAgentDomainReq) (*v1.DeleteAgentDomainRes, error) {
	return backend.AgentDomain().DeleteAgentDomain(ctx, req)
}
