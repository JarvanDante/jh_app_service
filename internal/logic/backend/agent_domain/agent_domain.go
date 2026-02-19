package agent_domain

import (
	"context"
	"fmt"
	"strings"
	v1 "jh_app_service/api/backend/agent_domain/v1"
	"jh_app_service/internal/dao"
	"jh_app_service/internal/middleware"
	"jh_app_service/internal/model/do"
	"jh_app_service/internal/model/entity"
	"jh_app_service/internal/service/backend"
	"jh_app_service/internal/util"

	"github.com/gogf/gf/v2/os/gtime"
)

type sAgentDomain struct{}

func init() { backend.RegisterAgentDomain(&sAgentDomain{}) }

func (s *sAgentDomain) GetAgentDomainList(ctx context.Context, req *v1.GetAgentDomainListReq) (*v1.GetAgentDomainListRes, error) {
	siteId, ok := middleware.GetSiteIdFromContext(ctx)
	if !ok || siteId == 0 {
		siteId = 1
	}
	page := req.Page
	if page <= 0 { page = 1 }
	size := req.Size
	if size <= 0 { size = 20 }

	query := dao.AgentDomain.Ctx(ctx).Where("site_id", siteId)
	if req.Status >= 0 { query = query.Where("status", req.Status) }
	if req.AgentId > 0 { query = query.Where("agent_id", req.AgentId) }
	if strings.TrimSpace(req.Domain) != "" {
		query = query.WhereLike("domain", "%"+strings.TrimSpace(req.Domain)+"%")
	}

	total, err := query.Count()
	if err != nil { return nil, err }

	var list []*entity.AgentDomain
	offset := int((page - 1) * size)
	err = query.Order("id DESC").Limit(offset, int(size)).Scan(&list)
	if err != nil { return nil, err }

	agentNames := map[int64]string{}
	for _, it := range list {
		if it.AgentId == 0 { continue }
		if _, ok := agentNames[it.AgentId]; ok { continue }
		var a entity.Agent
		_ = dao.Agent.Ctx(ctx).Where("site_id", siteId).Where("id", it.AgentId).Scan(&a)
		agentNames[it.AgentId] = a.AgentName
	}

	statusMap := map[int]string{0: "禁用", 1: "启用"}
	items := make([]*v1.AgentDomainItem, 0, len(list))
	for _, d := range list {
		items = append(items, &v1.AgentDomainItem{
			Id:         d.Id,
			AgentId:    d.AgentId,
			AgentName:  agentNames[d.AgentId],
			Domain:     d.Domain,
			Status:     int32(d.Status),
			StatusName: statusMap[d.Status],
			CreatedAt:  util.FormatTime(d.CreatedAt),
			UpdatedAt:  util.FormatTime(d.UpdatedAt),
		})
	}

	return &v1.GetAgentDomainListRes{
		List:        items,
		Total:       int32(total),
		CurrentPage: page,
		PerPage:     size,
		StatusList:  []*v1.OptionItem{{Value: 1, Name: "启用"}, {Value: 0, Name: "禁用"}},
	}, nil
}

func (s *sAgentDomain) CreateAgentDomain(ctx context.Context, req *v1.CreateAgentDomainReq) (*v1.CreateAgentDomainRes, error) {
	siteId, ok := middleware.GetSiteIdFromContext(ctx)
	if !ok || siteId == 0 { siteId = 1 }
	domain := strings.ToLower(strings.TrimSpace(req.Domain))
	if req.AgentId <= 0 || domain == "" {
		return nil, fmt.Errorf("agent_id和domain不能为空")
	}

	count, err := dao.AgentDomain.Ctx(ctx).Where("site_id", siteId).Where("domain", domain).Count()
	if err != nil { return nil, err }
	if count > 0 { return nil, fmt.Errorf("域名已存在") }

	_, err = dao.AgentDomain.Ctx(ctx).Data(do.AgentDomain{
		SiteId:    int64(siteId),
		AgentId:   req.AgentId,
		Domain:    domain,
		Status:    req.Status,
		CreatedAt: gtime.Now(),
		UpdatedAt: gtime.Now(),
	}).Insert()
	if err != nil { return nil, err }
	return &v1.CreateAgentDomainRes{}, nil
}

func (s *sAgentDomain) UpdateAgentDomain(ctx context.Context, req *v1.UpdateAgentDomainReq) (*v1.UpdateAgentDomainRes, error) {
	siteId, ok := middleware.GetSiteIdFromContext(ctx)
	if !ok || siteId == 0 { siteId = 1 }
	if req.Id <= 0 { return nil, fmt.Errorf("id不能为空") }
	domain := strings.ToLower(strings.TrimSpace(req.Domain))
	if req.AgentId <= 0 || domain == "" { return nil, fmt.Errorf("agent_id和domain不能为空") }

	count, err := dao.AgentDomain.Ctx(ctx).Where("site_id", siteId).Where("id", req.Id).Count()
	if err != nil { return nil, err }
	if count == 0 { return nil, fmt.Errorf("代理域名不存在") }

	repeat, err := dao.AgentDomain.Ctx(ctx).Where("site_id", siteId).Where("domain", domain).WhereNot("id", req.Id).Count()
	if err != nil { return nil, err }
	if repeat > 0 { return nil, fmt.Errorf("域名已存在") }

	_, err = dao.AgentDomain.Ctx(ctx).Where("site_id", siteId).Where("id", req.Id).Data(do.AgentDomain{
		AgentId:   req.AgentId,
		Domain:    domain,
		Status:    req.Status,
		UpdatedAt: gtime.Now(),
	}).Update()
	if err != nil { return nil, err }
	return &v1.UpdateAgentDomainRes{}, nil
}

func (s *sAgentDomain) DeleteAgentDomain(ctx context.Context, req *v1.DeleteAgentDomainReq) (*v1.DeleteAgentDomainRes, error) {
	siteId, ok := middleware.GetSiteIdFromContext(ctx)
	if !ok || siteId == 0 { siteId = 1 }
	if req.Id <= 0 { return nil, fmt.Errorf("id不能为空") }
	_, err := dao.AgentDomain.Ctx(ctx).Where("site_id", siteId).Where("id", req.Id).Delete()
	if err != nil { return nil, err }
	return &v1.DeleteAgentDomainRes{}, nil
}
