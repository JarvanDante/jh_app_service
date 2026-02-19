package agent

import (
	"context"
	v1 "jh_app_service/api/backend/agent/v1"
	"jh_app_service/internal/service/backend"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

type Controller struct {
	v1.UnimplementedAgentServer
}

func Register(s *grpcx.GrpcServer) {
	v1.RegisterAgentServer(s.Server, &Controller{})
}

func (*Controller) GetAgentList(ctx context.Context, req *v1.GetAgentListReq) (*v1.GetAgentListRes, error) {
	return backend.Agent().GetAgentList(ctx, req)
}

func (*Controller) CreateAgent(ctx context.Context, req *v1.CreateAgentReq) (*v1.CreateAgentRes, error) {
	return backend.Agent().CreateAgent(ctx, req)
}

func (*Controller) UpdateAgent(ctx context.Context, req *v1.UpdateAgentReq) (*v1.UpdateAgentRes, error) {
	return backend.Agent().UpdateAgent(ctx, req)
}

func (*Controller) DeleteAgent(ctx context.Context, req *v1.DeleteAgentReq) (*v1.DeleteAgentRes, error) {
	return backend.Agent().DeleteAgent(ctx, req)
}
