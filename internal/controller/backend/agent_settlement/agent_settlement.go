package agent_settlement

import (
	"context"
	v1 "jh_app_service/api/backend/agent_settlement/v1"
	"jh_app_service/internal/service/backend"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

type Controller struct {
	v1.UnimplementedAgentSettlementServer
}

func Register(s *grpcx.GrpcServer) {
	v1.RegisterAgentSettlementServer(s.Server, &Controller{})
}

func (*Controller) GetAgentSettlementDailyList(ctx context.Context, req *v1.GetAgentSettlementDailyListReq) (*v1.GetAgentSettlementDailyListRes, error) {
	return backend.AgentSettlement().GetAgentSettlementDailyList(ctx, req)
}

func (*Controller) CalcAgentSettlementDaily(ctx context.Context, req *v1.CalcAgentSettlementDailyReq) (*v1.CalcAgentSettlementDailyRes, error) {
	return backend.AgentSettlement().CalcAgentSettlementDaily(ctx, req)
}

func (*Controller) SettleAgentSettlementDaily(ctx context.Context, req *v1.SettleAgentSettlementDailyReq) (*v1.SettleAgentSettlementDailyRes, error) {
	return backend.AgentSettlement().SettleAgentSettlementDaily(ctx, req)
}

func (*Controller) SettleAgentSettlementBatch(ctx context.Context, req *v1.SettleAgentSettlementBatchReq) (*v1.SettleAgentSettlementBatchRes, error) {
	return backend.AgentSettlement().SettleAgentSettlementBatch(ctx, req)
}
