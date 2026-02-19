package backend

import (
	"context"
	v1 "jh_app_service/api/backend/agent_settlement/v1"
)

type IAgentSettlement interface {
	GetAgentSettlementDailyList(ctx context.Context, req *v1.GetAgentSettlementDailyListReq) (*v1.GetAgentSettlementDailyListRes, error)
	CalcAgentSettlementDaily(ctx context.Context, req *v1.CalcAgentSettlementDailyReq) (*v1.CalcAgentSettlementDailyRes, error)
	SettleAgentSettlementDaily(ctx context.Context, req *v1.SettleAgentSettlementDailyReq) (*v1.SettleAgentSettlementDailyRes, error)
	SettleAgentSettlementBatch(ctx context.Context, req *v1.SettleAgentSettlementBatchReq) (*v1.SettleAgentSettlementBatchRes, error)
}

var localAgentSettlement IAgentSettlement

func AgentSettlement() IAgentSettlement {
	if localAgentSettlement == nil {
		panic("implement not found for interface IAgentSettlement, forgot register?")
	}
	return localAgentSettlement
}

func RegisterAgentSettlement(i IAgentSettlement) {
	localAgentSettlement = i
}
