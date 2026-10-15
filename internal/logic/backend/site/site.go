package site

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"jh_app_service/api/backend/site/v1"
	"jh_app_service/internal/dao"
	"jh_app_service/internal/middleware"
	"jh_app_service/internal/model/do"
	"jh_app_service/internal/model/entity"
	"jh_app_service/internal/registry"
	"jh_app_service/internal/service/backend"
	"jh_app_service/internal/util"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
)

type (
	sSite struct{}
)

func init() {
	backend.RegisterSite(&sSite{})
}

type activityParamOption struct {
	Label string
	Value string
}

type activityParamField struct {
	Name    string
	Label   string
	Type    string
	Options []activityParamOption
}

func buildSigninParamFields() []activityParamField {
	fields := make([]activityParamField, 0, 33)
	fields = append(fields,
		activityParamField{
			Name:  "min_recharge_amount",
			Label: "领取门槛-当日最小充值",
			Type:  "number",
		},
		activityParamField{
			Name:  "min_valid_bet_amount",
			Label: "领取门槛-当日最小有效投注",
			Type:  "number",
		},
	)
	for day := 1; day <= 31; day++ {
		fields = append(fields, activityParamField{
			Name:  fmt.Sprintf("day_%d", day),
			Label: fmt.Sprintf("第%d天金额", day),
			Type:  "number",
		})
	}
	return fields
}

const luckyWheelSegmentCount = 8

func buildLuckyWheelParamFields() []activityParamField {
	fields := make([]activityParamField, 0, luckyWheelSegmentCount*6)
	for segment := 1; segment <= luckyWheelSegmentCount; segment++ {
		prefix := fmt.Sprintf("segment_%d_", segment)
		fields = append(fields,
			activityParamField{
				Name:  prefix + "prize_name",
				Label: fmt.Sprintf("扇区%d-奖项名称", segment),
				Type:  "text",
			},
			activityParamField{
				Name:  prefix + "reward_amount",
				Label: fmt.Sprintf("扇区%d-奖金额", segment),
				Type:  "number",
			},
			activityParamField{
				Name:  prefix + "weight",
				Label: fmt.Sprintf("扇区%d-权重", segment),
				Type:  "int",
			},
			activityParamField{
				Name:  prefix + "user_daily_win_limit",
				Label: fmt.Sprintf("扇区%d-单用户日中奖上限", segment),
				Type:  "int",
			},
			activityParamField{
				Name:  prefix + "min_vip_level",
				Label: fmt.Sprintf("扇区%d-最低VIP等级", segment),
				Type:  "int",
			},
			activityParamField{
				Name:  prefix + "icon",
				Label: fmt.Sprintf("扇区%d-图标", segment),
				Type:  "text",
			},
		)
	}
	return fields
}

func parseSigninDayField(name string) (int, bool) {
	if !strings.HasPrefix(name, "day_") {
		return 0, false
	}
	day, err := strconv.Atoi(strings.TrimPrefix(name, "day_"))
	if err != nil || day < 1 || day > 31 {
		return 0, false
	}
	return day, true
}

func parseLuckySegmentField(name string) (int, string, bool) {
	if !strings.HasPrefix(name, "segment_") {
		return 0, "", false
	}

	raw := strings.TrimPrefix(name, "segment_")
	parts := strings.SplitN(raw, "_", 2)
	if len(parts) != 2 {
		return 0, "", false
	}

	segment, err := strconv.Atoi(parts[0])
	if err != nil || segment < 1 || segment > luckyWheelSegmentCount {
		return 0, "", false
	}

	field := parts[1]
	switch field {
	case "prize_name", "reward_amount", "weight", "user_daily_win_limit", "min_vip_level", "icon":
		return segment, field, true
	default:
		return 0, "", false
	}
}

var activityParamMappings = map[string][]activityParamField{
	"activity_invite_bonus": {
		{
			Name:  "trigger_event",
			Label: "发放触发",
			Type:  "select",
			Options: []activityParamOption{
				{Label: "充值总和", Value: "1"},
				{Label: "首充", Value: "2"},
			},
		},
		{
			Name:  "reward_type",
			Label: "奖励类型",
			Type:  "select",
			Options: []activityParamOption{
				{Label: "固定", Value: "1"},
				{Label: "比例", Value: "2"},
			},
		},
		{
			Name:  "min_recharge_amount",
			Label: "触发最小计算金额",
			Type:  "number",
		},
		{
			Name:  "inviter_reward_amount",
			Label: "邀请人奖励金额",
			Type:  "number",
		},
	},
	"activity_register_bonus": {
		{
			Name:  "reward_amount",
			Label: "注册奖励金额",
			Type:  "number",
		},
		{
			Name:  "limit_ip",
			Label: "同IP注册限制次数",
			Type:  "int",
		},
		{
			Name:  "limit_device",
			Label: "同设备注册限制次数",
			Type:  "int",
		},
	},
	"activity_deposit_bonus": {
		{
			Name:  "trigger_event",
			Label: "发放触发",
			Type:  "select",
			Options: []activityParamOption{
				{Label: "充值总和", Value: "1"},
				{Label: "首充", Value: "2"},
			},
		},
		{
			Name:  "reward_type",
			Label: "奖励类型",
			Type:  "select",
			Options: []activityParamOption{
				{Label: "固定", Value: "1"},
				{Label: "比例", Value: "2"},
			},
		},
		{
			Name:  "min_recharge_amount",
			Label: "触发最小充值金额",
			Type:  "number",
		},
		{
			Name:  "reward_amount",
			Label: "奖励金额/比例值",
			Type:  "number",
		},
	},
	"activity_signin_bonus": buildSigninParamFields(),
	"activity_wheel_bonus":  buildLuckyWheelParamFields(),
}

// GetBasicSetting 获取站点基本设置
func (s *sSite) GetBasicSetting(ctx context.Context, req *v1.GetBasicSettingReq) (*v1.GetBasicSettingRes, error) {
	middleware.LogWithTrace(ctx, "info", "获取站点基本设置请求 - SiteId: %d", req.SiteId)

	// 安全修复(A5)：站点ID 取自管理员 token 的权威站点ID，忽略请求体里的 site_id
	authSiteId, ok := middleware.GetAdminSiteIdFromContext(ctx)
	if !ok || authSiteId == 0 {
		middleware.LogWithTrace(ctx, "error", "未获取管理员站点ID")
		return nil, fmt.Errorf("未获取站点ID")
	}
	siteId := int32(authSiteId)

	// 查询站点基本信息
	var siteInfo *entity.Site
	err := dao.Site.Ctx(ctx).Where(do.Site{
		Id: siteId,
	}).Scan(&siteInfo)

	if err != nil {
		middleware.LogWithTrace(ctx, "error", "查询站点信息失败: %v", err)
		return nil, fmt.Errorf("查询站点信息失败: %v", err)
	}

	if siteInfo == nil {
		middleware.LogWithTrace(ctx, "error", "站点信息不存在 - SiteId: %d", siteId)
		return nil, fmt.Errorf("站点信息不存在")
	}

	// 查询站点配置
	var siteConfig *entity.SiteConfig
	err = dao.SiteConfig.Ctx(ctx).Where(do.SiteConfig{
		SiteId: siteId,
	}).Scan(&siteConfig)

	if err != nil {
		middleware.LogWithTrace(ctx, "error", "查询站点配置失败: %v", err)
		return nil, fmt.Errorf("查询站点配置失败: %v", err)
	}

	if siteConfig == nil {
		middleware.LogWithTrace(ctx, "error", "站点配置不存在 - SiteId: %d", siteId)
		return nil, fmt.Errorf("站点配置不存在")
	}

	// 构建响应数据
	res := &v1.GetBasicSettingRes{
		Code:                 siteInfo.Code, // 从site表获取站点代码
		Name:                 siteInfo.Name, // 从site表获取站点名称
		RegisterTimeInterval: int32(siteConfig.RegisterTimeInterval),
		SwitchRegister:       siteConfig.SwitchRegister == 1,
		IsClose:              siteConfig.IsClose == 1,
		CloseReason:          siteConfig.CloseReason,
		UrlService:           siteConfig.UrlService,
		UrlAgentPc:           siteConfig.UrlTg,
		UrlMobile:            siteConfig.UrlMobile,
		WebUrl:               siteConfig.WebUrl,
		MinWithdraw:          int32(siteConfig.MinWithdraw),
		MaxWithdraw:          int32(siteConfig.MaxWithdraw),
		MobileLogo:           siteConfig.MobileLogo,
		DefaultAgentId:       int32(siteConfig.DefaultAgentId),
		DefaultAgentName:     "", // 这里可以根据 DefaultAgentId 查询代理名称
		Balance:              0,  // 这里可以查询站点余额
		BalanceReset:         0,  // 这里可以查询剩余额度
	}

	// 如果关闭原因为空，设置默认消息
	if res.CloseReason == "" {
		res.CloseReason = "您好，系统正在维护中，估计需要1个小时左右，给您造成不便，敬请谅解！"
	}

	middleware.LogWithTrace(ctx, "info", "获取站点基本设置成功 - SiteId: %d", siteId)
	return res, nil
}

// UpdateBasicSetting 更新站点基本设置
func (s *sSite) UpdateBasicSetting(ctx context.Context, req *v1.UpdateBasicSettingReq) (*v1.UpdateBasicSettingRes, error) {
	middleware.LogWithTrace(ctx, "info", "更新站点基本设置请求 - SiteId: %d", req.SiteId)

	// 安全修复(A5)：站点ID 取自管理员 token 的权威站点ID，忽略请求体里的 site_id
	authSiteId, ok := middleware.GetAdminSiteIdFromContext(ctx)
	if !ok || authSiteId == 0 {
		middleware.LogWithTrace(ctx, "error", "未获取管理员站点ID")
		return nil, fmt.Errorf("未获取站点ID")
	}
	siteId := int32(authSiteId)

	// 参数验证
	if err := s.validateUpdateBasicSetting(req); err != nil {
		middleware.LogWithTrace(ctx, "error", "参数验证失败: %v", err)
		return nil, fmt.Errorf("参数验证失败: %v", err)
	}

	// 查询现有配置
	var existingConfig *entity.SiteConfig
	err := dao.SiteConfig.Ctx(ctx).Where(do.SiteConfig{
		SiteId: siteId,
	}).Scan(&existingConfig)

	if err != nil {
		middleware.LogWithTrace(ctx, "error", "查询现有配置失败: %v", err)
		return nil, fmt.Errorf("查询现有配置失败: %v", err)
	}

	middleware.LogWithTrace(ctx, "info", "更新请求参数 - RegisterTimeInterval: %d, SwitchRegister: %t, IsClose: %t", req.RegisterTimeInterval, req.SwitchRegister, req.IsClose)
	middleware.LogWithTrace(ctx, "info", "更新请求参数 - CloseReason: '%s', UrlTg: '%s', UrlService: '%s'", req.CloseReason, req.UrlAgentPc, req.UrlService)
	middleware.LogWithTrace(ctx, "info", "更新请求参数 - MinWithdraw: %d, MaxWithdraw: %d", req.MinWithdraw, req.MaxWithdraw)

	// 使用map来精确控制要更新的字段
	updateFields := g.Map{}

	// 注册时间间隔 - 只有在请求中明确指定时才更新
	if req.RegisterTimeInterval > 0 {
		updateFields["register_time_interval"] = req.RegisterTimeInterval
		middleware.LogWithTrace(ctx, "info", "添加更新字段: register_time_interval = %d", req.RegisterTimeInterval)
	}

	// 布尔值字段 - 只有在请求中明确指定时才更新
	// 注意：protobuf中布尔值的默认值是false，我们需要区分"未设置"和"设置为false"
	// 这里我们假设如果传递了请求，就更新这些布尔值
	switchRegisterValue := g.Map{"true": 1, "false": 0}[fmt.Sprintf("%t", req.SwitchRegister)]
	updateFields["switch_register"] = switchRegisterValue
	middleware.LogWithTrace(ctx, "info", "添加更新字段: switch_register = %v (原值: %t)", switchRegisterValue, req.SwitchRegister)

	isCloseValue := g.Map{"true": 1, "false": 0}[fmt.Sprintf("%t", req.IsClose)]
	updateFields["is_close"] = isCloseValue
	middleware.LogWithTrace(ctx, "info", "添加更新字段: is_close = %v (原值: %t)", isCloseValue, req.IsClose)

	// 字符串字段 - 只有非空时才更新
	if req.CloseReason != "" {
		updateFields["close_reason"] = req.CloseReason
		middleware.LogWithTrace(ctx, "info", "添加更新字段: close_reason = '%s'", req.CloseReason)
	}
	if req.UrlAgentPc != "" {
		updateFields["url_tg"] = req.UrlAgentPc
		middleware.LogWithTrace(ctx, "info", "添加更新字段: url_tg = '%s'", req.UrlAgentPc)
	}
	if req.UrlMobile != "" {
		updateFields["url_mobile"] = req.UrlMobile
		middleware.LogWithTrace(ctx, "info", "添加更新字段: url_mobile = '%s'", req.UrlMobile)
	}
	if req.WebUrl != "" {
		updateFields["web_url"] = req.WebUrl
		middleware.LogWithTrace(ctx, "info", "添加更新字段: web_url = '%s'", req.WebUrl)
	}
	if req.UrlService != "" {
		updateFields["url_service"] = req.UrlService
		middleware.LogWithTrace(ctx, "info", "添加更新字段: url_service = '%s'", req.UrlService)
	}
	if req.MobileLogo != "" {
		updateFields["mobile_logo"] = req.MobileLogo
		middleware.LogWithTrace(ctx, "info", "添加更新字段: mobile_logo = '%s'", req.MobileLogo)
	}

	// 提现金额 - 只有大于0时才更新
	middleware.LogWithTrace(ctx, "info", "检查提现金额 - MinWithdraw: %d, MaxWithdraw: %d", req.MinWithdraw, req.MaxWithdraw)

	if req.MinWithdraw > 0 {
		if req.MinWithdraw < 1 {
			updateFields["min_withdraw"] = 1
			middleware.LogWithTrace(ctx, "info", "添加更新字段: min_withdraw = 1 (原值太小)")
		} else {
			updateFields["min_withdraw"] = req.MinWithdraw
			middleware.LogWithTrace(ctx, "info", "添加更新字段: min_withdraw = %d", req.MinWithdraw)
		}
	} else {
		middleware.LogWithTrace(ctx, "info", "跳过 min_withdraw 更新，值为: %d", req.MinWithdraw)
	}

	if req.MaxWithdraw > 0 {
		if req.MaxWithdraw < 1 {
			updateFields["max_withdraw"] = 9999999
			middleware.LogWithTrace(ctx, "info", "添加更新字段: max_withdraw = 9999999 (原值太小)")
		} else {
			updateFields["max_withdraw"] = req.MaxWithdraw
			middleware.LogWithTrace(ctx, "info", "添加更新字段: max_withdraw = %d", req.MaxWithdraw)
		}
	} else {
		middleware.LogWithTrace(ctx, "info", "跳过 max_withdraw 更新，值为: %d", req.MaxWithdraw)
	}

	middleware.LogWithTrace(ctx, "info", "准备更新的字段总数: %d, 字段: %+v", len(updateFields), updateFields)

	// 如果配置不存在，创建新配置
	if existingConfig == nil {
		updateFields["site_id"] = siteId
		middleware.LogWithTrace(ctx, "info", "配置不存在，创建新配置，字段: %+v", updateFields)
		result, err := dao.SiteConfig.Ctx(ctx).Data(updateFields).Insert()
		if err != nil {
			middleware.LogWithTrace(ctx, "error", "创建站点配置失败: %v", err)
			return nil, fmt.Errorf("创建站点配置失败: %v", err)
		}
		middleware.LogWithTrace(ctx, "info", "创建站点配置成功，影响行数: %v", result)
	} else {
		// 更新现有配置 - 只更新指定的字段
		middleware.LogWithTrace(ctx, "info", "配置存在，更新现有配置，SiteId: %d, 字段: %+v", siteId, updateFields)

		// 检查是否有字段需要更新
		if len(updateFields) == 0 {
			middleware.LogWithTrace(ctx, "warning", "没有字段需要更新，跳过数据库操作")
			return &v1.UpdateBasicSettingRes{
				Message: "没有字段需要更新",
			}, nil
		}

		result, err := dao.SiteConfig.Ctx(ctx).Where(do.SiteConfig{
			SiteId: siteId,
		}).Data(updateFields).Update()

		if err != nil {
			middleware.LogWithTrace(ctx, "error", "更新站点配置失败: %v", err)
			return nil, fmt.Errorf("更新站点配置失败: %v", err)
		}

		rowsAffected, _ := result.RowsAffected()
		middleware.LogWithTrace(ctx, "info", "数据库更新完成，影响行数: %d", rowsAffected)

		if rowsAffected == 0 {
			middleware.LogWithTrace(ctx, "warning", "数据库更新影响行数为0，可能是条件不匹配或数据未变化")
		}
	}

	middleware.LogWithTrace(ctx, "info", "更新站点基本设置成功 - SiteId: %d", siteId)
	return &v1.UpdateBasicSettingRes{
		Message: "设置成功",
	}, nil
}

// GetInviteBonusConfig 获取邀请活动配置
func (s *sSite) GetInviteBonusConfig(ctx context.Context, req *v1.GetInviteBonusConfigReq) (*v1.GetInviteBonusConfigRes, error) {
	// 后台用 admin_site_id，前台/资金服务用 site_id；都不信任请求体 site_id
	authSiteId, ok := middleware.ResolveTenantSiteId(ctx)
	if !ok || authSiteId == 0 {
		return nil, fmt.Errorf("未获取站点ID")
	}
	siteId := int32(authSiteId)

	code := strings.TrimSpace(req.Code)
	if code == "" {
		code = "INVITE"
	}

	siteCode := registry.GetDBConfigManagerV2().GetSiteCodeBySiteID(int64(siteId))
	if siteCode == "" {
		return nil, fmt.Errorf("站点代码不存在: site_id=%d", siteId)
	}

	// 确保使用站点数据库
	ctx = context.WithValue(ctx, "dynamic_db_group", "site_"+siteCode)

	// 查询活动定义
	var activity *entity.SiteActivity
	if err := dao.SiteActivity.Ctx(ctx).
		Where("site_code", siteCode).
		Where("code", code).
		Scan(&activity); err != nil {
		return nil, fmt.Errorf("查询活动失败: %v", err)
	}
	if activity == nil || activity.ActivityId == 0 {
		return nil, fmt.Errorf("活动不存在: code=%s", code)
	}

	if activity.Table == "" {
		return nil, fmt.Errorf("活动参数表为空: code=%s", code)
	}

	// 查询活动参数
	var config *entity.ActivityInviteBonus
	if err := dao.ActivityInviteBonus.Ctx(ctx).
		Where("site_code", siteCode).
		Where("activity_id", activity.ActivityId).
		Scan(&config); err != nil {
		return nil, fmt.Errorf("查询活动参数失败: %v", err)
	}
	if config == nil || config.Id == 0 {
		return nil, fmt.Errorf("活动参数不存在: activity_id=%d", activity.ActivityId)
	}

	return &v1.GetInviteBonusConfigRes{
		ActivityId:          int32(activity.ActivityId),
		Code:                activity.Code,
		Status:              int32(activity.Status),
		Table:               activity.Table,
		TriggerEvent:        int32(config.TriggerEvent),
		RewardType:          int32(config.RewardType),
		MinRechargeAmount:   config.MinRechargeAmount,
		InviterRewardAmount: config.InviterRewardAmount,
		FixedParams:         activity.FixedParams,
	}, nil
}

// GetSiteActivityList 查询活动配置列表
func (s *sSite) GetSiteActivityList(ctx context.Context, req *v1.GetSiteActivityListReq) (*v1.GetSiteActivityListRes, error) {
	// 后台用 admin_site_id，前台/资金服务用 site_id；都不信任请求体 site_id
	authSiteId, ok := middleware.ResolveTenantSiteId(ctx)
	if !ok || authSiteId == 0 {
		return nil, fmt.Errorf("未获取站点ID")
	}
	siteId := int32(authSiteId)

	siteCode := registry.GetDBConfigManagerV2().GetSiteCodeBySiteID(int64(siteId))
	if siteCode == "" {
		return nil, fmt.Errorf("站点代码不存在: site_id=%d", siteId)
	}

	ctx = context.WithValue(ctx, "dynamic_db_group", "site_"+siteCode)

	model := dao.SiteActivity.Ctx(ctx).Where("site_code", siteCode)
	if req.ActivityId > 0 {
		model = model.Where("activity_id", req.ActivityId)
	}
	if strings.TrimSpace(req.Code) != "" {
		model = model.Where("code", strings.TrimSpace(req.Code))
	}

	var list []*entity.SiteActivity
	if err := model.Order("sort asc, id asc").Scan(&list); err != nil {
		return nil, fmt.Errorf("查询活动配置失败: %v", err)
	}

	res := &v1.GetSiteActivityListRes{List: make([]*v1.SiteActivityInfo, 0, len(list))}
	for _, item := range list {
		res.List = append(res.List, &v1.SiteActivityInfo{
			Id:          int32(item.Id),
			SiteCode:    item.SiteCode,
			ActivityId:  int32(item.ActivityId),
			Name:        item.Name,
			Code:        item.Code,
			Description: item.Description,
			StartTime:   util.FormatTime(item.StartTime),
			EndTime:     util.FormatTime(item.EndTime),
			FixedParams: item.FixedParams,
			Status:      int32(item.Status),
			Sort:        int32(item.Sort),
			Table:       item.Table,
			Uri:         item.Uri,
			CreatedAt:   util.FormatTime(item.CreatedAt),
			UpdatedAt:   util.FormatTime(item.UpdatedAt),
		})
	}

	return res, nil
}

// GetSiteActivityParams 获取活动参数配置
func (s *sSite) GetSiteActivityParams(ctx context.Context, req *v1.GetSiteActivityParamsReq) (*v1.GetSiteActivityParamsRes, error) {
	if req.ActivityId <= 0 {
		return nil, fmt.Errorf("活动ID不能为空")
	}

	// 后台用 admin_site_id，前台/资金服务用 site_id；都不信任请求体 site_id / site_code
	authSiteId, ok := middleware.ResolveTenantSiteId(ctx)
	if !ok || authSiteId == 0 {
		return nil, fmt.Errorf("未获取站点ID")
	}
	siteCode := registry.GetDBConfigManagerV2().GetSiteCodeBySiteID(int64(authSiteId))
	if siteCode == "" {
		return nil, fmt.Errorf("站点代码不存在: site_id=%d", authSiteId)
	}

	ctx = context.WithValue(ctx, "dynamic_db_group", "site_"+siteCode)

	var activity *entity.SiteActivity
	if err := dao.SiteActivity.Ctx(ctx).
		Where("site_code", siteCode).
		Where("activity_id", req.ActivityId).
		Scan(&activity); err != nil {
		return nil, fmt.Errorf("查询活动失败: %v", err)
	}
	if activity == nil || activity.ActivityId == 0 {
		return nil, fmt.Errorf("活动不存在: activity_id=%d", req.ActivityId)
	}

	if strings.TrimSpace(activity.Table) == "" {
		return nil, fmt.Errorf("没有对应参数配置表，请确认")
	}

	fields, ok := activityParamMappings[activity.Table]
	if !ok {
		return nil, fmt.Errorf("没有对应参数配置表，请确认")
	}

	values := map[string]string{}
	switch activity.Table {
	case "activity_invite_bonus":
		var config *entity.ActivityInviteBonus
		if err := dao.ActivityInviteBonus.Ctx(ctx).
			Where("site_code", siteCode).
			Where("activity_id", activity.ActivityId).
			Scan(&config); err != nil {
			return nil, fmt.Errorf("查询活动参数失败: %v", err)
		}
		if config == nil || config.Id == 0 {
			return nil, fmt.Errorf("没有对应参数配置表，请确认")
		}
		values["trigger_event"] = strconv.Itoa(config.TriggerEvent)
		values["reward_type"] = strconv.Itoa(config.RewardType)
		values["min_recharge_amount"] = gconv.String(config.MinRechargeAmount)
		values["inviter_reward_amount"] = gconv.String(config.InviterRewardAmount)
	case "activity_register_bonus":
		var config *entity.ActivityRegisterBonus
		if err := dao.ActivityRegisterBonus.Ctx(ctx).
			Where("site_code", siteCode).
			Where("activity_id", activity.ActivityId).
			Scan(&config); err != nil {
			return nil, fmt.Errorf("查询活动参数失败: %v", err)
		}
		if config == nil || config.Id == 0 {
			return nil, fmt.Errorf("没有对应参数配置表，请确认")
		}
		values["reward_amount"] = gconv.String(config.RewardAmount)
		values["limit_ip"] = strconv.Itoa(config.LimitIp)
		values["limit_device"] = strconv.Itoa(config.LimitDevice)
	case "activity_deposit_bonus":
		var config *entity.ActivityDepositBonus
		if err := dao.ActivityDepositBonus.Ctx(ctx).
			Where("site_code", siteCode).
			Where("activity_id", activity.ActivityId).
			Scan(&config); err != nil {
			return nil, fmt.Errorf("查询活动参数失败: %v", err)
		}
		if config == nil || config.Id == 0 {
			return nil, fmt.Errorf("没有对应参数配置表，请确认")
		}
		values["trigger_event"] = strconv.Itoa(config.TriggerEvent)
		values["reward_type"] = strconv.Itoa(config.RewardType)
		values["min_recharge_amount"] = gconv.String(config.MinRechargeAmount)
		values["reward_amount"] = gconv.String(config.RewardAmount)
	case "activity_signin_bonus":
		for day := 1; day <= 31; day++ {
			values[fmt.Sprintf("day_%d", day)] = "0"
		}
		values["min_recharge_amount"] = "0"
		values["min_valid_bet_amount"] = "0"

		if strings.TrimSpace(activity.FixedParams) != "" {
			fixed := map[string]interface{}{}
			if err := json.Unmarshal([]byte(activity.FixedParams), &fixed); err == nil {
				if v, ok := fixed["min_recharge_amount"]; ok {
					values["min_recharge_amount"] = gconv.String(v)
				}
				if v, ok := fixed["min_valid_bet_amount"]; ok {
					values["min_valid_bet_amount"] = gconv.String(v)
				}
			}
		}

		var list []*entity.ActivitySigninBonus
		if err := dao.ActivitySigninBonus.Ctx(ctx).
			Where("site_code", siteCode).
			Where("activity_id", activity.ActivityId).
			Order("day_no asc").
			Scan(&list); err != nil {
			return nil, fmt.Errorf("查询活动参数失败: %v", err)
		}
		for _, item := range list {
			if item == nil || item.DayNo < 1 || item.DayNo > 31 {
				continue
			}
			values[fmt.Sprintf("day_%d", item.DayNo)] = gconv.String(item.RewardAmount)
		}
	case "activity_wheel_bonus":
		for segment := 1; segment <= luckyWheelSegmentCount; segment++ {
			prefix := fmt.Sprintf("segment_%d_", segment)
			values[prefix+"prize_name"] = ""
			values[prefix+"reward_amount"] = "0"
			values[prefix+"weight"] = "0"
			values[prefix+"user_daily_win_limit"] = "0"
			values[prefix+"min_vip_level"] = "0"
			values[prefix+"icon"] = ""
		}

		var list []*entity.ActivityWheelBonus
		if err := dao.ActivityWheelBonus.Ctx(ctx).
			Where("site_code", siteCode).
			Where("activity_id", activity.ActivityId).
			Order("segment_no asc").
			Scan(&list); err != nil {
			return nil, fmt.Errorf("查询活动参数失败: %v", err)
		}
		for _, item := range list {
			if item == nil || item.SegmentNo < 1 || item.SegmentNo > luckyWheelSegmentCount {
				continue
			}
			prefix := fmt.Sprintf("segment_%d_", item.SegmentNo)
			values[prefix+"prize_name"] = item.PrizeName
			values[prefix+"reward_amount"] = gconv.String(item.RewardAmount)
			values[prefix+"weight"] = gconv.String(item.Weight)
			values[prefix+"user_daily_win_limit"] = gconv.String(item.UserDailyWinLimit)
			values[prefix+"min_vip_level"] = gconv.String(item.MinVipLevel)
			values[prefix+"icon"] = item.Icon
		}
	default:
		return nil, fmt.Errorf("没有对应参数配置表，请确认")
	}

	res := &v1.GetSiteActivityParamsRes{
		Table:  activity.Table,
		Fields: make([]*v1.ActivityParamField, 0, len(fields)),
		Values: values,
	}
	for _, field := range fields {
		pbField := &v1.ActivityParamField{
			Name:  field.Name,
			Label: field.Label,
			Type:  field.Type,
		}
		if len(field.Options) > 0 {
			pbField.Options = make([]*v1.ParamOption, 0, len(field.Options))
			for _, opt := range field.Options {
				pbField.Options = append(pbField.Options, &v1.ParamOption{
					Label: opt.Label,
					Value: opt.Value,
				})
			}
		}
		res.Fields = append(res.Fields, pbField)
	}

	return res, nil
}

// UpdateSiteActivity 更新活动配置
func (s *sSite) UpdateSiteActivity(ctx context.Context, req *v1.UpdateSiteActivityReq) (*v1.UpdateSiteActivityRes, error) {
	// 安全修复(A5)：站点ID 取自管理员 token 的权威站点ID，忽略请求体里的 site_id；物理库切换也基于该权威 siteId
	authSiteId, ok := middleware.GetAdminSiteIdFromContext(ctx)
	if !ok || authSiteId == 0 {
		return nil, fmt.Errorf("未获取站点ID")
	}
	siteId := int32(authSiteId)

	siteCode := registry.GetDBConfigManagerV2().GetSiteCodeBySiteID(int64(siteId))
	if siteCode == "" {
		return nil, fmt.Errorf("站点代码不存在: site_id=%d", siteId)
	}

	ctx = context.WithValue(ctx, "dynamic_db_group", "site_"+siteCode)

	updateFields := g.Map{}
	var activity *entity.SiteActivity
	if len(req.ParamValues) > 0 {
		var err error
		activity, err = getSiteActivityForUpdate(ctx, siteCode, req.ActivityId, req.Code)
		if err != nil {
			return &v1.UpdateSiteActivityRes{Success: false, Message: err.Error()}, nil
		}
		if strings.TrimSpace(activity.Table) == "" {
			return &v1.UpdateSiteActivityRes{Success: false, Message: "没有对应参数配置表，请确认"}, nil
		}
		fixedParams, err := applyActivityParamValues(ctx, siteCode, activity, req.ParamValues)
		if err != nil {
			return &v1.UpdateSiteActivityRes{Success: false, Message: err.Error()}, nil
		}
		if strings.TrimSpace(fixedParams) != "" {
			updateFields["fixed_params"] = fixedParams
		}
	}
	if strings.TrimSpace(req.Description) != "" {
		updateFields["description"] = strings.TrimSpace(req.Description)
	}
	if strings.TrimSpace(req.StartTime) != "" {
		if t, err := gtime.StrToTime(strings.TrimSpace(req.StartTime)); err == nil {
			updateFields["start_time"] = t
		} else {
			updateFields["start_time"] = strings.TrimSpace(req.StartTime)
		}
	}
	if strings.TrimSpace(req.EndTime) != "" {
		if t, err := gtime.StrToTime(strings.TrimSpace(req.EndTime)); err == nil {
			updateFields["end_time"] = t
		} else {
			updateFields["end_time"] = strings.TrimSpace(req.EndTime)
		}
	}
	if len(req.ParamValues) == 0 && strings.TrimSpace(req.FixedParams) != "" {
		updateFields["fixed_params"] = strings.TrimSpace(req.FixedParams)
	}
	if req.StatusSet {
		updateFields["status"] = req.Status
	}
	if req.SortSet {
		updateFields["sort"] = req.Sort
	}

	if len(updateFields) == 0 {
		return &v1.UpdateSiteActivityRes{Success: false, Message: "没有需要更新的字段"}, nil
	}
	updateFields["updated_at"] = gtime.Now()

	model := dao.SiteActivity.Ctx(ctx).Where("site_code", siteCode)
	if req.ActivityId > 0 {
		model = model.Where("activity_id", req.ActivityId)
	} else if strings.TrimSpace(req.Code) != "" {
		model = model.Where("code", strings.TrimSpace(req.Code))
	} else {
		return &v1.UpdateSiteActivityRes{Success: false, Message: "活动ID或编码不能为空"}, nil
	}

	result, err := model.Data(updateFields).Update()
	if err != nil {
		return &v1.UpdateSiteActivityRes{Success: false, Message: err.Error()}, nil
	}
	rows, _ := result.RowsAffected()
	if rows == 0 && len(req.ParamValues) == 0 {
		return &v1.UpdateSiteActivityRes{Success: false, Message: "未找到匹配活动或数据未变化"}, nil
	}

	return &v1.UpdateSiteActivityRes{Success: true, Message: "更新成功"}, nil
}

func getSiteActivityForUpdate(ctx context.Context, siteCode string, activityId int32, code string) (*entity.SiteActivity, error) {
	model := dao.SiteActivity.Ctx(ctx).Where("site_code", siteCode)
	if activityId > 0 {
		model = model.Where("activity_id", activityId)
	} else if strings.TrimSpace(code) != "" {
		model = model.Where("code", strings.TrimSpace(code))
	} else {
		return nil, fmt.Errorf("活动ID或编码不能为空")
	}

	var activity *entity.SiteActivity
	if err := model.Scan(&activity); err != nil {
		return nil, fmt.Errorf("查询活动失败: %v", err)
	}
	if activity == nil || activity.ActivityId == 0 {
		return nil, fmt.Errorf("活动不存在")
	}
	return activity, nil
}

func applyActivityParamValues(ctx context.Context, siteCode string, activity *entity.SiteActivity, paramValues map[string]string) (string, error) {
	switch activity.Table {
	case "activity_invite_bonus":
		return applyInviteBonusParamValues(ctx, siteCode, activity.ActivityId, paramValues)
	case "activity_register_bonus":
		return applyRegisterBonusParamValues(ctx, siteCode, activity.ActivityId, paramValues)
	case "activity_deposit_bonus":
		return applyDepositBonusParamValues(ctx, siteCode, activity.ActivityId, paramValues)
	case "activity_signin_bonus":
		return applySigninBonusParamValues(ctx, siteCode, activity.ActivityId, paramValues)
	case "activity_wheel_bonus":
		return applyWheelBonusParamValues(ctx, siteCode, activity.ActivityId, paramValues)
	default:
		return "", fmt.Errorf("没有对应参数配置表，请确认")
	}
}

func applyInviteBonusParamValues(ctx context.Context, siteCode string, activityId int, paramValues map[string]string) (string, error) {
	var config *entity.ActivityInviteBonus
	if err := dao.ActivityInviteBonus.Ctx(ctx).
		Where("site_code", siteCode).
		Where("activity_id", activityId).
		Scan(&config); err != nil {
		return "", fmt.Errorf("查询活动参数失败: %v", err)
	}
	if config == nil || config.Id == 0 {
		return "", fmt.Errorf("没有对应参数配置表，请确认")
	}

	updateFields := g.Map{}
	for key, value := range paramValues {
		val := strings.TrimSpace(value)
		if val == "" {
			continue
		}
		switch key {
		case "trigger_event":
			config.TriggerEvent = gconv.Int(val)
			updateFields["trigger_event"] = config.TriggerEvent
		case "reward_type":
			config.RewardType = gconv.Int(val)
			updateFields["reward_type"] = config.RewardType
		case "min_recharge_amount":
			config.MinRechargeAmount = gconv.Float64(val)
			updateFields["min_recharge_amount"] = config.MinRechargeAmount
		case "inviter_reward_amount":
			config.InviterRewardAmount = gconv.Float64(val)
			updateFields["inviter_reward_amount"] = config.InviterRewardAmount
		}
	}

	if len(updateFields) > 0 {
		updateFields["updated_at"] = gtime.Now()
		if _, err := dao.ActivityInviteBonus.Ctx(ctx).
			Where("site_code", siteCode).
			Where("activity_id", activityId).
			Data(updateFields).
			Update(); err != nil {
			return "", fmt.Errorf("更新活动参数失败: %v", err)
		}
	}

	fixedParams := map[string]interface{}{
		"trigger_event":         config.TriggerEvent,
		"reward_type":           config.RewardType,
		"min_recharge_amount":   config.MinRechargeAmount,
		"inviter_reward_amount": config.InviterRewardAmount,
	}
	data, err := json.Marshal(fixedParams)
	if err != nil {
		return "", fmt.Errorf("构建固定参数失败: %v", err)
	}
	return string(data), nil
}

func applyRegisterBonusParamValues(ctx context.Context, siteCode string, activityId int, paramValues map[string]string) (string, error) {
	var config *entity.ActivityRegisterBonus
	if err := dao.ActivityRegisterBonus.Ctx(ctx).
		Where("site_code", siteCode).
		Where("activity_id", activityId).
		Scan(&config); err != nil {
		return "", fmt.Errorf("查询活动参数失败: %v", err)
	}
	if config == nil || config.Id == 0 {
		return "", fmt.Errorf("没有对应参数配置表，请确认")
	}

	updateFields := g.Map{}
	for key, value := range paramValues {
		val := strings.TrimSpace(value)
		if val == "" {
			continue
		}
		switch key {
		case "reward_amount":
			config.RewardAmount = gconv.Float64(val)
			updateFields["reward_amount"] = config.RewardAmount
		case "limit_ip":
			config.LimitIp = gconv.Int(val)
			updateFields["limit_ip"] = config.LimitIp
		case "limit_device":
			config.LimitDevice = gconv.Int(val)
			updateFields["limit_device"] = config.LimitDevice
		}
	}

	if len(updateFields) > 0 {
		updateFields["updated_at"] = gtime.Now()
		if _, err := dao.ActivityRegisterBonus.Ctx(ctx).
			Where("site_code", siteCode).
			Where("activity_id", activityId).
			Data(updateFields).
			Update(); err != nil {
			return "", fmt.Errorf("更新活动参数失败: %v", err)
		}
	}

	fixedParams := map[string]interface{}{
		"reward_amount": config.RewardAmount,
		"limit_ip":      config.LimitIp,
		"limit_device":  config.LimitDevice,
	}
	data, err := json.Marshal(fixedParams)
	if err != nil {
		return "", fmt.Errorf("构建固定参数失败: %v", err)
	}
	return string(data), nil
}

func applyDepositBonusParamValues(ctx context.Context, siteCode string, activityId int, paramValues map[string]string) (string, error) {
	var config *entity.ActivityDepositBonus
	if err := dao.ActivityDepositBonus.Ctx(ctx).
		Where("site_code", siteCode).
		Where("activity_id", activityId).
		Scan(&config); err != nil {
		return "", fmt.Errorf("查询活动参数失败: %v", err)
	}
	if config == nil || config.Id == 0 {
		return "", fmt.Errorf("没有对应参数配置表，请确认")
	}

	updateFields := g.Map{}
	for key, value := range paramValues {
		val := strings.TrimSpace(value)
		if val == "" {
			continue
		}
		switch key {
		case "trigger_event":
			config.TriggerEvent = gconv.Int(val)
			updateFields["trigger_event"] = config.TriggerEvent
		case "reward_type":
			config.RewardType = gconv.Int(val)
			updateFields["reward_type"] = config.RewardType
		case "min_recharge_amount":
			config.MinRechargeAmount = gconv.Float64(val)
			updateFields["min_recharge_amount"] = config.MinRechargeAmount
		case "reward_amount":
			config.RewardAmount = gconv.Float64(val)
			updateFields["reward_amount"] = config.RewardAmount
		}
	}

	if len(updateFields) > 0 {
		updateFields["updated_at"] = gtime.Now()
		if _, err := dao.ActivityDepositBonus.Ctx(ctx).
			Where("site_code", siteCode).
			Where("activity_id", activityId).
			Data(updateFields).
			Update(); err != nil {
			return "", fmt.Errorf("更新活动参数失败: %v", err)
		}
	}

	fixedParams := map[string]interface{}{
		"trigger_event":       config.TriggerEvent,
		"reward_type":         config.RewardType,
		"min_recharge_amount": config.MinRechargeAmount,
		"reward_amount":       config.RewardAmount,
	}
	data, err := json.Marshal(fixedParams)
	if err != nil {
		return "", fmt.Errorf("构建固定参数失败: %v", err)
	}
	return string(data), nil
}

func applySigninBonusParamValues(ctx context.Context, siteCode string, activityId int, paramValues map[string]string) (string, error) {
	if activityId <= 0 {
		return "", fmt.Errorf("活动ID不能为空")
	}

	var exists []*entity.ActivitySigninBonus
	if err := dao.ActivitySigninBonus.Ctx(ctx).
		Where("site_code", siteCode).
		Where("activity_id", activityId).
		Scan(&exists); err != nil {
		return "", fmt.Errorf("查询活动参数失败: %v", err)
	}

	dayConfigMap := make(map[int]*entity.ActivitySigninBonus, 31)
	for _, item := range exists {
		if item == nil || item.DayNo < 1 || item.DayNo > 31 {
			continue
		}
		dayConfigMap[item.DayNo] = item
	}

	now := gtime.Now()
	minRechargeAmount := 0.0
	minValidBetAmount := 0.0
	if raw, ok := paramValues["min_recharge_amount"]; ok {
		minRechargeAmount = gconv.Float64(strings.TrimSpace(raw))
		if minRechargeAmount < 0 {
			minRechargeAmount = 0
		}
		minRechargeAmount = float64(int(minRechargeAmount*100+0.5)) / 100
	}
	if raw, ok := paramValues["min_valid_bet_amount"]; ok {
		minValidBetAmount = gconv.Float64(strings.TrimSpace(raw))
		if minValidBetAmount < 0 {
			minValidBetAmount = 0
		}
		minValidBetAmount = float64(int(minValidBetAmount*100+0.5)) / 100
	}

	for key, value := range paramValues {
		dayNo, ok := parseSigninDayField(key)
		if !ok {
			continue
		}

		val := strings.TrimSpace(value)
		if val == "" {
			val = "0"
		}
		amount := gconv.Float64(val)
		if amount < 0 {
			amount = 0
		}
		amount = float64(int(amount*100+0.5)) / 100

		if old, ok := dayConfigMap[dayNo]; ok {
			if old.RewardAmount == amount {
				continue
			}
			if _, err := dao.ActivitySigninBonus.Ctx(ctx).
				Where("id", old.Id).
				Data(g.Map{
					"reward_amount": amount,
					"updated_at":    now,
				}).
				Update(); err != nil {
				return "", fmt.Errorf("更新活动参数失败: %v", err)
			}
			continue
		}

		if _, err := dao.ActivitySigninBonus.Ctx(ctx).Data(g.Map{
			"site_code":     siteCode,
			"activity_id":   activityId,
			"day_no":        dayNo,
			"reward_amount": amount,
			"created_at":    now,
			"updated_at":    now,
		}).Insert(); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				return "", fmt.Errorf("签到参数表唯一键冲突，请确认唯一键为(site_code, activity_id, day_no)")
			}
			return "", fmt.Errorf("新增活动参数失败: %v", err)
		}
	}

	// 回查构建 fixed_params
	values := make(map[string]float64, 33)
	for day := 1; day <= 31; day++ {
		values[fmt.Sprintf("day_%d", day)] = 0
	}

	var finalList []*entity.ActivitySigninBonus
	if err := dao.ActivitySigninBonus.Ctx(ctx).
		Where("site_code", siteCode).
		Where("activity_id", activityId).
		Order("day_no asc").
		Scan(&finalList); err != nil {
		return "", fmt.Errorf("查询活动参数失败: %v", err)
	}
	for _, item := range finalList {
		if item == nil || item.DayNo < 1 || item.DayNo > 31 {
			continue
		}
		values[fmt.Sprintf("day_%d", item.DayNo)] = item.RewardAmount
	}
	values["min_recharge_amount"] = minRechargeAmount
	values["min_valid_bet_amount"] = minValidBetAmount

	data, err := json.Marshal(values)
	if err != nil {
		return "", fmt.Errorf("构建固定参数失败: %v", err)
	}
	return string(data), nil
}

func applyWheelBonusParamValues(ctx context.Context, siteCode string, activityId int, paramValues map[string]string) (string, error) {
	if activityId <= 0 {
		return "", fmt.Errorf("活动ID不能为空")
	}

	var exists []*entity.ActivityWheelBonus
	if err := dao.ActivityWheelBonus.Ctx(ctx).
		Where("site_code", siteCode).
		Where("activity_id", activityId).
		Scan(&exists); err != nil {
		return "", fmt.Errorf("查询活动参数失败: %v", err)
	}

	segmentConfigMap := make(map[int]*entity.ActivityWheelBonus, luckyWheelSegmentCount)
	for segment := 1; segment <= luckyWheelSegmentCount; segment++ {
		segmentConfigMap[segment] = &entity.ActivityWheelBonus{
			SiteCode:          siteCode,
			ActivityId:        activityId,
			SegmentNo:         uint(segment),
			PrizeName:         "",
			RewardAmount:      0,
			Weight:            0,
			UserDailyWinLimit: 0,
			MinVipLevel:       0,
			Icon:              "",
		}
	}
	for _, item := range exists {
		if item == nil || item.SegmentNo < 1 || item.SegmentNo > luckyWheelSegmentCount {
			continue
		}
		segmentConfigMap[int(item.SegmentNo)] = item
	}

	for key, value := range paramValues {
		segment, field, ok := parseLuckySegmentField(key)
		if !ok {
			continue
		}

		config := segmentConfigMap[segment]
		if config == nil {
			continue
		}

		val := strings.TrimSpace(value)
		switch field {
		case "prize_name":
			config.PrizeName = val
		case "reward_amount":
			amount := gconv.Float64(val)
			if amount < 0 {
				amount = 0
			}
			config.RewardAmount = float64(int(amount*100+0.5)) / 100
		case "weight":
			weight := gconv.Int(val)
			if weight < 0 {
				weight = 0
			}
			config.Weight = uint(weight)
		case "user_daily_win_limit":
			limit := gconv.Int(val)
			if limit < 0 {
				limit = 0
			}
			config.UserDailyWinLimit = uint(limit)
		case "min_vip_level":
			level := gconv.Int(val)
			if level < 0 {
				level = 0
			}
			config.MinVipLevel = uint(level)
		case "icon":
			config.Icon = val
		}
	}

	now := gtime.Now()
	for segment := 1; segment <= luckyWheelSegmentCount; segment++ {
		config := segmentConfigMap[segment]
		if config == nil {
			continue
		}

		data := g.Map{
			"prize_name":           config.PrizeName,
			"reward_amount":        config.RewardAmount,
			"weight":               config.Weight,
			"user_daily_win_limit": config.UserDailyWinLimit,
			"min_vip_level":        config.MinVipLevel,
			"icon":                 config.Icon,
			"updated_at":           now,
		}

		if config.Id > 0 {
			if _, err := dao.ActivityWheelBonus.Ctx(ctx).
				Where("id", config.Id).
				Data(data).
				Update(); err != nil {
				return "", fmt.Errorf("更新活动参数失败: %v", err)
			}
			continue
		}

		data["site_code"] = siteCode
		data["activity_id"] = activityId
		data["segment_no"] = segment
		data["created_at"] = now

		if _, err := dao.ActivityWheelBonus.Ctx(ctx).Data(data).Insert(); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				return "", fmt.Errorf("大转盘参数表唯一键冲突，请确认唯一键为(site_code, activity_id, segment_no)")
			}
			return "", fmt.Errorf("新增活动参数失败: %v", err)
		}
	}

	values := make(map[string]interface{}, luckyWheelSegmentCount*6)
	for segment := 1; segment <= luckyWheelSegmentCount; segment++ {
		prefix := fmt.Sprintf("segment_%d_", segment)
		values[prefix+"prize_name"] = ""
		values[prefix+"reward_amount"] = 0.0
		values[prefix+"weight"] = 0
		values[prefix+"user_daily_win_limit"] = 0
		values[prefix+"min_vip_level"] = 0
		values[prefix+"icon"] = ""
	}

	var finalList []*entity.ActivityWheelBonus
	if err := dao.ActivityWheelBonus.Ctx(ctx).
		Where("site_code", siteCode).
		Where("activity_id", activityId).
		Order("segment_no asc").
		Scan(&finalList); err != nil {
		return "", fmt.Errorf("查询活动参数失败: %v", err)
	}
	for _, item := range finalList {
		if item == nil || item.SegmentNo < 1 || item.SegmentNo > luckyWheelSegmentCount {
			continue
		}
		prefix := fmt.Sprintf("segment_%d_", item.SegmentNo)
		values[prefix+"prize_name"] = item.PrizeName
		values[prefix+"reward_amount"] = item.RewardAmount
		values[prefix+"weight"] = item.Weight
		values[prefix+"user_daily_win_limit"] = item.UserDailyWinLimit
		values[prefix+"min_vip_level"] = item.MinVipLevel
		values[prefix+"icon"] = item.Icon
	}

	data, err := json.Marshal(values)
	if err != nil {
		return "", fmt.Errorf("构建固定参数失败: %v", err)
	}
	return string(data), nil
}

// validateUpdateBasicSetting 验证更新基本设置的参数
func (s *sSite) validateUpdateBasicSetting(req *v1.UpdateBasicSettingReq) error {
	// 验证URL格式（简单验证）
	if req.UrlAgentPc != "" && len(req.UrlAgentPc) < 7 {
		return fmt.Errorf("Telegram地址格式错误")
	}

	if req.UrlMobile != "" && len(req.UrlMobile) < 7 {
		return fmt.Errorf("手机域名地址格式错误")
	}

	// 验证提现金额
	if req.MinWithdraw < 0 {
		return fmt.Errorf("单笔最低提现金额不能为负数")
	}

	if req.MaxWithdraw < 0 {
		return fmt.Errorf("单笔最高提现金额不能为负数")
	}

	if req.MinWithdraw > req.MaxWithdraw && req.MaxWithdraw > 0 {
		return fmt.Errorf("单笔最低提现金额不能大于最高提现金额")
	}

	// 验证关闭原因长度
	if len(req.CloseReason) > 255 {
		return fmt.Errorf("暂时关闭网站原因不能超过255个字符")
	}

	return nil
}
