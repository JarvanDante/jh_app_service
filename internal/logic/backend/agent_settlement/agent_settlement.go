package agent_settlement

import (
	"context"
	"fmt"
	v1 "jh_app_service/api/backend/agent_settlement/v1"
	gamegrpc "jh_app_service/api/grpc/game/v1"
	"jh_app_service/internal/dao"
	"jh_app_service/internal/middleware"
	"jh_app_service/internal/model/do"
	"jh_app_service/internal/model/entity"
	"jh_app_service/internal/registry"
	"jh_app_service/internal/service/backend"
	"jh_app_service/internal/util"
	"strings"

	"github.com/gogf/gf/v2/util/gconv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/gogf/gf/v2/os/gtime"
)

type sAgentSettlement struct{}

func init() { backend.RegisterAgentSettlement(&sAgentSettlement{}) }

func getSiteID(ctx context.Context) int64 {
	siteId, ok := middleware.GetSiteIdFromContext(ctx)
	if !ok || siteId == 0 {
		return 1
	}
	return int64(siteId)
}

func normalizeDateRange(startDate, endDate string) (string, string, error) {
	startDate = strings.TrimSpace(startDate)
	endDate = strings.TrimSpace(endDate)
	if startDate == "" || endDate == "" {
		return "", "", fmt.Errorf("开始日期和结束日期不能为空")
	}
	if startDate > endDate {
		return "", "", fmt.Errorf("开始日期不能大于结束日期")
	}
	return startDate, endDate, nil
}

func formatDate(t *gtime.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format("Y-m-d")
}

func (s *sAgentSettlement) GetAgentSettlementDailyList(ctx context.Context, req *v1.GetAgentSettlementDailyListReq) (*v1.GetAgentSettlementDailyListRes, error) {
	siteId := getSiteID(ctx)
	page := req.Page
	if page <= 0 {
		page = 1
	}
	size := req.Size
	if size <= 0 {
		size = 20
	}

	query := dao.AgentSettlementDaily.Ctx(ctx).As("s").
		LeftJoin("agent a", "a.id=s.agent_id AND a.site_id=s.site_id").
		Where("s.site_id", siteId)

	if req.AgentId > 0 {
		query = query.Where("s.agent_id", req.AgentId)
	}
	if req.SettleStatus >= 0 {
		query = query.Where("s.status", req.SettleStatus)
	}
	if req.StartDate != "" {
		query = query.WhereGTE("s.stat_date", req.StartDate)
	}
	if req.EndDate != "" {
		query = query.WhereLTE("s.stat_date", req.EndDate)
	}

	total, err := query.Count("s.id")
	if err != nil {
		return nil, err
	}

	offset := int((page - 1) * size)
	var list []struct {
		Id               int64       `json:"id"`
		StatDate         *gtime.Time `json:"stat_date"`
		AgentId          int64       `json:"agent_id"`
		AgentName        string      `json:"agent_name"`
		BetAmount        float64     `json:"bet_amount"`
		ValidBetAmount   float64     `json:"valid_bet_amount"`
		WinLoss          float64     `json:"win_loss"`
		CommissionAmount float64     `json:"commission_amount"`
		Status           int         `json:"status"`
		SettledAt        *gtime.Time `json:"settled_at"`
		CreatedAt        *gtime.Time `json:"created_at"`
		UpdatedAt        *gtime.Time `json:"updated_at"`
		Rate             float64     `json:"rate"`
	}

	err = query.Fields(`
		s.id,
		s.stat_date,
		s.agent_id,
		COALESCE(a.agent_name, '') AS agent_name,
		s.bet_amount,
		s.valid_bet_amount,
		s.win_loss,
		s.commission_amount,
		s.status,
		s.settled_at,
		s.created_at,
		s.updated_at,
		COALESCE(a.rate, 0) AS rate
	`).Order("s.stat_date DESC").Order("s.id DESC").Limit(offset, int(size)).Scan(&list)
	if err != nil {
		return nil, err
	}

	var summary struct {
		MemberCount      int64   `json:"member_count"`
		BetAmount        float64 `json:"bet_amount"`
		ValidAmount      float64 `json:"valid_amount"`
		NetAmount        float64 `json:"net_amount"`
		CommissionAmount float64 `json:"commission_amount"`
	}

	summaryQuery := dao.AgentSettlementDaily.Ctx(ctx).As("s").
		Where("s.site_id", siteId)
	if req.AgentId > 0 {
		summaryQuery = summaryQuery.Where("s.agent_id", req.AgentId)
	}
	if req.SettleStatus >= 0 {
		summaryQuery = summaryQuery.Where("s.status", req.SettleStatus)
	}
	if req.StartDate != "" {
		summaryQuery = summaryQuery.WhereGTE("s.stat_date", req.StartDate)
	}
	if req.EndDate != "" {
		summaryQuery = summaryQuery.WhereLTE("s.stat_date", req.EndDate)
	}
	_ = summaryQuery.Fields(`
		COALESCE(SUM(s.bet_amount),0) AS bet_amount,
		COALESCE(SUM(s.valid_bet_amount),0) AS valid_amount,
		COALESCE(SUM(s.win_loss),0) AS net_amount,
		COALESCE(SUM(s.commission_amount),0) AS commission_amount
	`).Scan(&summary)

	statusMap := map[int]string{0: "待结算", 1: "已结算", 2: "已作废"}
	items := make([]*v1.AgentSettlementDailyItem, 0, len(list))
	for _, row := range list {
		items = append(items, &v1.AgentSettlementDailyItem{
			Id:               row.Id,
			StatDate:         formatDate(row.StatDate),
			AgentId:          row.AgentId,
			AgentName:        row.AgentName,
			MemberCount:      0,
			BetAmount:        row.BetAmount,
			ValidAmount:      row.ValidBetAmount,
			NetAmount:        row.WinLoss,
			Rate:             row.Rate,
			CommissionAmount: row.CommissionAmount,
			SettleStatus:     int32(row.Status),
			SettleStatusName: statusMap[row.Status],
			SettledAt:        util.FormatTime(row.SettledAt),
			CreatedAt:        util.FormatTime(row.CreatedAt),
			UpdatedAt:        util.FormatTime(row.UpdatedAt),
		})
	}

	return &v1.GetAgentSettlementDailyListRes{
		List:        items,
		Total:       int32(total),
		CurrentPage: page,
		PerPage:     size,
		StatusList: []*v1.OptionItem{{Value: 0, Name: "待结算"}, {Value: 1, Name: "已结算"}, {Value: 2, Name: "已作废"}},
		Summary: &v1.SettlementSummary{
			MemberCount:      summary.MemberCount,
			BetAmount:        summary.BetAmount,
			ValidAmount:      summary.ValidAmount,
			NetAmount:        summary.NetAmount,
			CommissionAmount: summary.CommissionAmount,
		},
	}, nil
}

func (s *sAgentSettlement) CalcAgentSettlementDaily(ctx context.Context, req *v1.CalcAgentSettlementDailyReq) (*v1.CalcAgentSettlementDailyRes, error) {
	siteId := getSiteID(ctx)
	startDate, endDate, err := normalizeDateRange(req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	gameAddr, err := registry.GetServiceAddr("jh_game_service")
	if err != nil {
		return nil, fmt.Errorf("获取游戏服务地址失败: %v", err)
	}
	conn, err := grpc.NewClient(gameAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("连接游戏服务失败: %v", err)
	}
	defer conn.Close()

	gameClient := gamegrpc.NewAgentSettlementClient(conn)
	md := metadata.New(map[string]string{"site_id": gconv.String(siteId)})
	grpcCtx := metadata.NewOutgoingContext(ctx, md)
	statRes, err := gameClient.GetAgentDailyBetStats(grpcCtx, &gamegrpc.GetAgentDailyBetStatsReq{
		StartDate: startDate,
		EndDate:   endDate,
		AgentId:   req.AgentId,
	})
	if err != nil {
		return nil, fmt.Errorf("调用游戏服务统计失败: %v", err)
	}

	var agentRows []struct {
		Id         int64   `json:"id"`
		SettleType int32   `json:"settle_type"`
		Rate       float64 `json:"rate"`
	}
	agentQuery := dao.Agent.Ctx(ctx).Where("site_id", siteId)
	if req.AgentId > 0 {
		agentQuery = agentQuery.Where("id", req.AgentId)
	}
	if err = agentQuery.Fields("id, settle_type, rate").Scan(&agentRows); err != nil {
		return nil, err
	}
	agentMap := make(map[int64]struct {
		SettleType int32
		Rate       float64
	}, len(agentRows))
	for _, a := range agentRows {
		agentMap[a.Id] = struct {
			SettleType int32
			Rate       float64
		}{SettleType: a.SettleType, Rate: a.Rate}
	}

	affected := 0
	now := gtime.Now()
	for _, row := range statRes.List {
		agentCfg, ok := agentMap[row.AgentId]
		if !ok {
			continue
		}

		commission := 0.0
		switch agentCfg.SettleType {
		case 2:
			commission = row.ValidBetAmount * agentCfg.Rate / 100
		default:
			if row.WinLoss < 0 {
				commission = -row.WinLoss * agentCfg.Rate / 100
			}
		}

		var exist entity.AgentSettlementDaily
		err = dao.AgentSettlementDaily.Ctx(ctx).
			Where("site_id", siteId).
			Where("agent_id", row.AgentId).
			Where("stat_date", row.StatDate).
			Scan(&exist)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "no rows") {
				return nil, err
			}
			err = nil
		}

		if exist.Id > 0 {
			if exist.Status == 1 {
				continue
			}
			_, err = dao.AgentSettlementDaily.Ctx(ctx).
				Where("id", exist.Id).
				Where("site_id", siteId).
				Data(do.AgentSettlementDaily{
					BetAmount:        row.BetAmount,
					ValidBetAmount:   row.ValidBetAmount,
					WinLoss:          row.WinLoss,
					CommissionAmount: commission,
					Status:           0,
					SettledAt:        nil,
					UpdatedAt:        now,
				}).Update()
			if err != nil {
				return nil, err
			}
			affected++
			continue
		}

		_, err = dao.AgentSettlementDaily.Ctx(ctx).Data(do.AgentSettlementDaily{
			SiteId:           siteId,
			AgentId:          row.AgentId,
			StatDate:         gtime.NewFromStr(row.StatDate + " 00:00:00"),
			BetAmount:        row.BetAmount,
			ValidBetAmount:   row.ValidBetAmount,
			WinLoss:          row.WinLoss,
			CommissionAmount: commission,
			Status:           0,
			CreatedAt:        now,
			UpdatedAt:        now,
		}).Insert()
		if err != nil {
			return nil, err
		}
		affected++
	}

	return &v1.CalcAgentSettlementDailyRes{Affected: int32(affected)}, nil
}

func (s *sAgentSettlement) SettleAgentSettlementDaily(ctx context.Context, req *v1.SettleAgentSettlementDailyReq) (*v1.SettleAgentSettlementDailyRes, error) {
	siteId := getSiteID(ctx)
	if req.Id <= 0 {
		return nil, fmt.Errorf("id不能为空")
	}
	_, err := dao.AgentSettlementDaily.Ctx(ctx).
		Where("site_id", siteId).
		Where("id", req.Id).
		Where("status", 0).
		Data(do.AgentSettlementDaily{Status: 1, SettledAt: gtime.Now(), UpdatedAt: gtime.Now()}).
		Update()
	if err != nil {
		return nil, err
	}
	return &v1.SettleAgentSettlementDailyRes{}, nil
}

func (s *sAgentSettlement) SettleAgentSettlementBatch(ctx context.Context, req *v1.SettleAgentSettlementBatchReq) (*v1.SettleAgentSettlementBatchRes, error) {
	siteId := getSiteID(ctx)
	startDate, endDate, err := normalizeDateRange(req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	query := dao.AgentSettlementDaily.Ctx(ctx).
		Where("site_id", siteId).
		Where("status", 0).
		WhereGTE("stat_date", startDate).
		WhereLTE("stat_date", endDate)
	if req.AgentId > 0 {
		query = query.Where("agent_id", req.AgentId)
	}

	result, err := query.Data(do.AgentSettlementDaily{
		Status:    1,
		SettledAt: gtime.Now(),
		UpdatedAt: gtime.Now(),
	}).Update()
	if err != nil {
		return nil, err
	}
	affected, _ := result.RowsAffected()
	return &v1.SettleAgentSettlementBatchRes{Affected: int32(affected)}, nil
}
