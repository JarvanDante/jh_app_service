package backend

import (
	"context"
	v1 "jh_app_service/api/backend/agent/v1"
)

type (
	IAgent interface {
		GetAgentList(ctx context.Context, req *v1.GetAgentListReq) (*v1.GetAgentListRes, error)
		CreateAgent(ctx context.Context, req *v1.CreateAgentReq) (*v1.CreateAgentRes, error)
		UpdateAgent(ctx context.Context, req *v1.UpdateAgentReq) (*v1.UpdateAgentRes, error)
		DeleteAgent(ctx context.Context, req *v1.DeleteAgentReq) (*v1.DeleteAgentRes, error)
	}
)

var localAgent IAgent

func Agent() IAgent {
	if localAgent == nil {
		panic("implement not found for interface IAgent, forgot register?")
	}
	return localAgent
}

func RegisterAgent(i IAgent) {
	localAgent = i
}
