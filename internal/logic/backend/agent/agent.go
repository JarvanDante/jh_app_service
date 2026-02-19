package agent

import (
	"context"
	"fmt"
	v1 "jh_app_service/api/backend/agent/v1"
	"jh_app_service/internal/dao"
	"jh_app_service/internal/middleware"
	"jh_app_service/internal/model/do"
	"jh_app_service/internal/model/entity"
	"jh_app_service/internal/service/backend"
	"jh_app_service/internal/util"

	"github.com/gogf/gf/v2/os/gtime"
)

type sAgent struct{}

func init() { backend.RegisterAgent(&sAgent{}) }

func (s *sAgent) GetAgentList(ctx context.Context, req *v1.GetAgentListReq) (*v1.GetAgentListRes, error) {
	siteId, ok := middleware.GetSiteIdFromContext(ctx)
	if !ok || siteId == 0 {
		siteId = 1
	}
	page := req.Page
	if page <= 0 { page = 1 }
	size := req.Size
	if size <= 0 { size = 20 }

	query := dao.Agent.Ctx(ctx).Where("site_id", siteId)
	if req.Status >= 0 { query = query.Where("status", req.Status) }
	if req.Name != "" { query = query.WhereLike("agent_name", "%"+req.Name+"%") }

	total, err := query.Count()
	if err != nil { return nil, err }

	var list []*entity.Agent
	offset := int((page - 1) * size)
	err = query.Order("id DESC").Limit(offset, int(size)).Scan(&list)
	if err != nil { return nil, err }

	statusMap := map[int]string{0: "禁用", 1: "启用"}
	settleMap := map[int]string{1: "亏损分成", 2: "流水分成"}
	items := make([]*v1.AgentItem, 0, len(list))
	for _, a := range list {
		items = append(items, &v1.AgentItem{
			Id:             a.Id,
			ParentId:       a.ParentId,
			AgentCode:      a.AgentCode,
			AgentName:      a.AgentName,
			Status:         int32(a.Status),
			StatusName:     statusMap[a.Status],
			SettleType:     int32(a.SettleType),
			SettleTypeName: settleMap[a.SettleType],
			Rate:           a.Rate,
			Remark:         a.Remark,
			CreatedAt:      util.FormatTime(a.CreatedAt),
			UpdatedAt:      util.FormatTime(a.UpdatedAt),
		})
	}

	return &v1.GetAgentListRes{
		List:        items,
		Total:       int32(total),
		CurrentPage: page,
		PerPage:     size,
		StatusList: []*v1.OptionItem{{Value: 1, Name: "启用"}, {Value: 0, Name: "禁用"}},
		SettleTypeList: []*v1.OptionItem{{Value: 1, Name: "亏损分成"}, {Value: 2, Name: "流水分成"}},
	}, nil
}

func (s *sAgent) CreateAgent(ctx context.Context, req *v1.CreateAgentReq) (*v1.CreateAgentRes, error) {
	siteId, ok := middleware.GetSiteIdFromContext(ctx)
	if !ok || siteId == 0 { siteId = 1 }
	if req.AgentCode == "" || req.AgentName == "" { return nil, fmt.Errorf("代理编码和名称不能为空") }
	count, err := dao.Agent.Ctx(ctx).Where("site_id", siteId).Where("agent_code", req.AgentCode).Count()
	if err != nil { return nil, err }
	if count > 0 { return nil, fmt.Errorf("代理编码已存在") }

	_, err = dao.Agent.Ctx(ctx).Data(do.Agent{
		SiteId:     int64(siteId),
		ParentId:   req.ParentId,
		AgentCode:  req.AgentCode,
		AgentName:  req.AgentName,
		Status:     req.Status,
		SettleType: req.SettleType,
		Rate:       req.Rate,
		Remark:     req.Remark,
		CreatedAt:  gtime.Now(),
		UpdatedAt:  gtime.Now(),
	}).Insert()
	if err != nil { return nil, err }
	return &v1.CreateAgentRes{}, nil
}

func (s *sAgent) UpdateAgent(ctx context.Context, req *v1.UpdateAgentReq) (*v1.UpdateAgentRes, error) {
	siteId, ok := middleware.GetSiteIdFromContext(ctx)
	if !ok || siteId == 0 { siteId = 1 }
	if req.Id <= 0 { return nil, fmt.Errorf("id不能为空") }
	count, err := dao.Agent.Ctx(ctx).Where("site_id", siteId).Where("id", req.Id).Count()
	if err != nil { return nil, err }
	if count == 0 { return nil, fmt.Errorf("代理不存在") }

	_, err = dao.Agent.Ctx(ctx).Where("site_id", siteId).Where("id", req.Id).Data(do.Agent{
		ParentId:   req.ParentId,
		AgentCode:  req.AgentCode,
		AgentName:  req.AgentName,
		Status:     req.Status,
		SettleType: req.SettleType,
		Rate:       req.Rate,
		Remark:     req.Remark,
		UpdatedAt:  gtime.Now(),
	}).Update()
	if err != nil { return nil, err }
	return &v1.UpdateAgentRes{}, nil
}

func (s *sAgent) DeleteAgent(ctx context.Context, req *v1.DeleteAgentReq) (*v1.DeleteAgentRes, error) {
	siteId, ok := middleware.GetSiteIdFromContext(ctx)
	if !ok || siteId == 0 { siteId = 1 }
	if req.Id <= 0 { return nil, fmt.Errorf("id不能为空") }
	_, err := dao.Agent.Ctx(ctx).Where("site_id", siteId).Where("id", req.Id).Delete()
	if err != nil { return nil, err }
	return &v1.DeleteAgentRes{}, nil
}
