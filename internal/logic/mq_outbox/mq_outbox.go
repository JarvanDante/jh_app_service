package mq_outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"jh_app_service/internal/dao"
	"jh_app_service/internal/middleware"
	"jh_app_service/internal/model/do"
	"jh_app_service/internal/model/entity"
	"jh_app_service/internal/mq"
	"jh_app_service/internal/registry"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

type sMqOutbox struct{}

var insOutbox = sMqOutbox{}
var outboxMissing sync.Map

func MqOutbox() *sMqOutbox {
	return &insOutbox
}

// OutboxStatus 定义 Outbox 状态常量
const (
	OutboxStatusPending = 0 // 待发送
	OutboxStatusSent    = 1 // 已发送
	OutboxStatusFailed  = 2 // 发送失败
)

// CreateInitBalanceOutbox 创建初始化余额的 Outbox 记录
func (s *sMqOutbox) CreateInitBalanceOutbox(ctx context.Context, tx gdb.TX, siteID, userID int64, username, currency, siteCode string) error {
	msg := &mq.InitBalanceUserMessage{
		SiteID:   int32(siteID),
		UserID:   int32(userID),
		Username: username,
		Currency: currency,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	outbox := &do.MqOutbox{
		EventType:     "user_created",
		AggregateType: "user",
		AggregateId:   fmt.Sprintf("%d", userID),
		SiteCode:      siteCode,
		Exchange:      "user.created",
		RoutingKey:    "user.init_balance",
		Payload:       string(payload),
		Status:        OutboxStatusPending,
		RetryCount:    0,
		NextRetryAt:   gtime.Now(),
	}

	_, err = dao.MqOutbox.Ctx(ctx).TX(tx).Insert(outbox)
	return err
}

// CreateSyncUserToGameOutbox 创建同步用户到游戏服务的 Outbox 记录
func (s *sMqOutbox) CreateSyncUserToGameOutbox(ctx context.Context, tx gdb.TX, siteID int64, userID int64, agentID int64, username, siteCode string) error {
	msg := &mq.SyncUserToGameMessage{
		SiteID:   int32(siteID),
		SiteCode: siteCode,
		UserID:   userID,
		AgentID:  agentID,
		Username: username,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	outbox := &do.MqOutbox{
		EventType:     "user_created",
		AggregateType: "user",
		AggregateId:   fmt.Sprintf("%d", userID),
		SiteCode:      siteCode,
		Exchange:      "user.created",
		RoutingKey:    "user.sync_to_game",
		Payload:       string(payload),
		Status:        OutboxStatusPending,
		RetryCount:    0,
		NextRetryAt:   gtime.Now(),
	}

	_, err = dao.MqOutbox.Ctx(ctx).TX(tx).Insert(outbox)
	return err
}

// CreateSyncUserToClickhouseOutbox 创建同步用户到 ClickHouse 的 Outbox 记录
func (s *sMqOutbox) CreateSyncUserToClickhouseOutbox(ctx context.Context, tx gdb.TX, siteID int64, user *entity.User, siteCode string) error {
	if user == nil || user.Id == 0 {
		return fmt.Errorf("用户信息为空")
	}

	msg := &mq.SyncUserToClickHouseMessage{
		EventID:      0, // 先占位，插入Outbox后再更新
		SiteID:       uint32(siteID),
		UserID:       uint64(user.Id),
		InviterID:    uint64(user.InviterId),
		Username:     user.Username,
		Status:       uint8(user.Status),
		RegisterIP:   user.RegisterIp,
		RegisterTime: formatGTime(user.RegisterTime),
		CreatedAt:    formatGTime(user.CreatedAt),
		UpdatedAt:    formatGTime(user.UpdatedAt),
		PayTimes:     uint32(user.PayTimes),
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	outbox := &do.MqOutbox{
		EventType:     "user_sync_clickhouse",
		AggregateType: "user",
		AggregateId:   fmt.Sprintf("%d", user.Id),
		SiteCode:      siteCode,
		Exchange:      "analytics.sync",
		RoutingKey:    "analytics.user",
		Payload:       string(payload), // 先写入占位payload
		Status:        OutboxStatusPending,
		RetryCount:    0,
		NextRetryAt:   gtime.Now(),
	}

	result, err := dao.MqOutbox.Ctx(ctx).TX(tx).Insert(outbox)
	if err != nil {
		return err
	}

	outboxId, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取Outbox ID失败: %w", err)
	}

	// 更新payload中的event_id
	msg.EventID = uint64(outboxId)
	payload, err = json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	_, err = dao.MqOutbox.Ctx(ctx).TX(tx).Where("id", outboxId).Update(do.MqOutbox{
		Payload:   string(payload),
		UpdatedAt: gtime.Now(),
	})
	return err
}

// ProcessPendingMessages 处理待发送的消息（所有站点）
func (s *sMqOutbox) ProcessPendingMessages(ctx context.Context) error {
	return s.ProcessPendingMessagesForSite(ctx, "")
}

// ProcessPendingMessagesForSite 处理指定站点的待发送消息
// 如果 siteCode 为空，则处理所有站点
func (s *sMqOutbox) ProcessPendingMessagesForSite(ctx context.Context, siteCode string) error {
	var sites []string

	if siteCode != "" {
		// 处理指定站点
		sites = []string{siteCode}
	} else {
		// 获取所有站点配置
		dbManager := registry.GetDBConfigManagerV2()
		sites = dbManager.GetAllSiteCodes()

		if len(sites) == 0 {
			g.Log().Warning(ctx, "没有找到任何站点配置")
			return nil
		}
	}

	totalSuccess := 0
	totalFail := 0

	// 遍历每个站点
	for _, site := range sites {
		if _, ok := outboxMissing.Load(site); ok {
			continue
		}

		// 数据库配置组名格式为 site_{code}
		dbGroup := fmt.Sprintf("site_%s", site)

		// 创建带站点信息的 context
		// 设置 dynamic_db_group 用于动态数据库路由
		siteCtx := context.WithValue(ctx, "dynamic_db_group", dbGroup)
		siteCtx = context.WithValue(siteCtx, middleware.SiteCodeContextKey, site)

		// 查询该站点的待发送消息
		var messages []*entity.MqOutbox
		err := dao.MqOutbox.Ctx(siteCtx).
			WhereIn("status", []int{OutboxStatusPending, OutboxStatusFailed}).
			WhereLTE("next_retry_at", gtime.Now()).
			OrderAsc("created_at").
			Limit(100). // 每个站点最多处理100条
			Scan(&messages)

		if err != nil {
			if isTableNotFoundErr(err) {
				if _, loaded := outboxMissing.LoadOrStore(site, struct{}{}); !loaded {
					g.Log().Warningf(siteCtx, "站点 %s 缺少mq_outbox表，已跳过", site)
				}
				continue
			}
			g.Log().Errorf(siteCtx, "查询站点 %s 的待发送消息失败: %v", site, err)
			continue
		}

		if len(messages) == 0 {
			continue
		}

		g.Log().Infof(siteCtx, "站点 %s 找到 %d 条待发送消息", site, len(messages))

		// 处理该站点的消息
		for _, msg := range messages {
			if err := s.processMessage(siteCtx, msg); err != nil {
				g.Log().Errorf(siteCtx, "处理消息失败 [ID=%d, SiteCode=%s]: %v", msg.Id, site, err)
				totalFail++
			} else {
				totalSuccess++
			}
		}
	}

	if totalSuccess > 0 || totalFail > 0 {
		g.Log().Infof(ctx, "消息处理完成: 成功=%d, 失败=%d", totalSuccess, totalFail)
	}

	return nil
}

// processMessage 处理单条消息
func (s *sMqOutbox) processMessage(ctx context.Context, msg *entity.MqOutbox) error {
	// 根据 routing_key 处理不同的消息
	switch msg.RoutingKey {
	case "user.init_balance":
		return s.processInitBalanceMessage(ctx, msg)
	case "user.sync_to_game":
		return s.processSyncUserToGameMessage(ctx, msg)
	case "analytics.user":
		return s.processAnalyticsMessage(ctx, msg)
	default:
		return fmt.Errorf("未知的 routing key: %s", msg.RoutingKey)
	}
}

// processAnalyticsMessage 处理同步到 ClickHouse 的消息（通用发布）
func (s *sMqOutbox) processAnalyticsMessage(ctx context.Context, outboxMsg *entity.MqOutbox) error {
	// 创建带站点信息的 context
	dbGroup := fmt.Sprintf("site_%s", outboxMsg.SiteCode)
	msgCtx := context.WithValue(ctx, "dynamic_db_group", dbGroup)
	msgCtx = context.WithValue(msgCtx, middleware.SiteCodeContextKey, outboxMsg.SiteCode)

	// 直接投递到 MQ（payload 已经是完整 JSON）
	err := mq.PublishRaw(context.Background(), outboxMsg.Exchange, outboxMsg.RoutingKey, []byte(outboxMsg.Payload))
	if err != nil {
		s.markAsFailed(msgCtx, outboxMsg.Id, outboxMsg.SiteCode, err.Error(), true)
		return fmt.Errorf("发送MQ消息失败: %w", err)
	}

	// 发送成功，标记为已发送
	_, err = dao.MqOutbox.Ctx(msgCtx).Where("id", outboxMsg.Id).Update(do.MqOutbox{
		Status:    OutboxStatusSent,
		UpdatedAt: gtime.Now(),
	})
	if err != nil {
		g.Log().Errorf(msgCtx, "更新消息状态失败 [ID=%d]: %v", outboxMsg.Id, err)
	} else {
		g.Log().Infof(msgCtx, "同步用户到ClickHouse消息发送成功 [ID=%d, UserID=%s, SiteCode=%s]",
			outboxMsg.Id, outboxMsg.AggregateId, outboxMsg.SiteCode)
	}

	return err
}

// processInitBalanceMessage 处理初始化余额消息
func (s *sMqOutbox) processInitBalanceMessage(ctx context.Context, outboxMsg *entity.MqOutbox) error {
	// 创建带站点信息的 context
	// 数据库配置组名格式为 site_{code}
	dbGroup := fmt.Sprintf("site_%s", outboxMsg.SiteCode)
	msgCtx := context.WithValue(ctx, "dynamic_db_group", dbGroup)
	msgCtx = context.WithValue(msgCtx, middleware.SiteCodeContextKey, outboxMsg.SiteCode)

	// 解析消息体
	var msg mq.InitBalanceUserMessage
	if err := json.Unmarshal([]byte(outboxMsg.Payload), &msg); err != nil {
		// 解析失败，标记为失败状态（不再重试）
		s.markAsFailed(msgCtx, outboxMsg.Id, outboxMsg.SiteCode, fmt.Sprintf("解析消息失败: %v", err), false)
		return fmt.Errorf("解析消息失败: %w", err)
	}

	// 发送到 RabbitMQ
	err := mq.PublishInitBalanceUser(context.Background(), &msg)
	if err != nil {
		// 发送失败，更新重试信息
		s.markAsFailed(msgCtx, outboxMsg.Id, outboxMsg.SiteCode, err.Error(), true)
		return fmt.Errorf("发送MQ消息失败: %w", err)
	}

	// 发送成功，标记为已发送
	_, err = dao.MqOutbox.Ctx(msgCtx).Where("id", outboxMsg.Id).Update(do.MqOutbox{
		Status:    OutboxStatusSent,
		UpdatedAt: gtime.Now(),
	})

	if err != nil {
		g.Log().Errorf(msgCtx, "更新消息状态失败 [ID=%d]: %v", outboxMsg.Id, err)
	} else {
		g.Log().Infof(msgCtx, "消息发送成功 [ID=%d, EventType=%s, AggregateId=%s, SiteCode=%s]",
			outboxMsg.Id, outboxMsg.EventType, outboxMsg.AggregateId, outboxMsg.SiteCode)
	}

	return err
}

// processSyncUserToGameMessage 处理同步用户到游戏服务消息
func (s *sMqOutbox) processSyncUserToGameMessage(ctx context.Context, outboxMsg *entity.MqOutbox) error {
	// 创建带站点信息的 context
	// 数据库配置组名格式为 site_{code}
	dbGroup := fmt.Sprintf("site_%s", outboxMsg.SiteCode)
	msgCtx := context.WithValue(ctx, "dynamic_db_group", dbGroup)
	msgCtx = context.WithValue(msgCtx, middleware.SiteCodeContextKey, outboxMsg.SiteCode)

	// 解析消息体
	var msg mq.SyncUserToGameMessage
	if err := json.Unmarshal([]byte(outboxMsg.Payload), &msg); err != nil {
		// 解析失败，标记为失败状态（不再重试）
		s.markAsFailed(msgCtx, outboxMsg.Id, outboxMsg.SiteCode, fmt.Sprintf("解析消息失败: %v", err), false)
		return fmt.Errorf("解析消息失败: %w", err)
	}

	// 发送到 RabbitMQ
	err := mq.PublishSyncUserToGame(context.Background(), &msg)
	if err != nil {
		// 发送失败，更新重试信息
		s.markAsFailed(msgCtx, outboxMsg.Id, outboxMsg.SiteCode, err.Error(), true)
		return fmt.Errorf("发送MQ消息失败: %w", err)
	}

	// 发送成功，标记为已发送
	_, err = dao.MqOutbox.Ctx(msgCtx).Where("id", outboxMsg.Id).Update(do.MqOutbox{
		Status:    OutboxStatusSent,
		UpdatedAt: gtime.Now(),
	})

	if err != nil {
		g.Log().Errorf(msgCtx, "更新消息状态失败 [ID=%d]: %v", outboxMsg.Id, err)
	} else {
		g.Log().Infof(msgCtx, "同步用户到游戏服务消息发送成功 [ID=%d, UserID=%d, SiteCode=%s]",
			outboxMsg.Id, msg.UserID, outboxMsg.SiteCode)
	}

	return err
}

// markAsFailed 标记消息为失败状态
func (s *sMqOutbox) markAsFailed(ctx context.Context, id uint64, siteCode, errMsg string, shouldRetry bool) {
	// 创建带站点信息的 context
	// 数据库配置组名格式为 site_{code}
	dbGroup := fmt.Sprintf("site_%s", siteCode)
	updateCtx := context.WithValue(ctx, "dynamic_db_group", dbGroup)
	updateCtx = context.WithValue(updateCtx, middleware.SiteCodeContextKey, siteCode)

	update := do.MqOutbox{
		LastError: errMsg,
		UpdatedAt: gtime.Now(),
	}

	if shouldRetry {
		// 计算下次重试时间（指数退避）
		var retryCount int
		dao.MqOutbox.Ctx(updateCtx).Where("id", id).Value("retry_count", &retryCount)
		retryCount++

		// 指数退避: 1分钟, 5分钟, 15分钟, 30分钟, 1小时
		retryDelays := []int{1, 5, 15, 30, 60}
		delayMinutes := 60 // 默认1小时
		if retryCount <= len(retryDelays) {
			delayMinutes = retryDelays[retryCount-1]
		}

		update.Status = OutboxStatusFailed
		update.RetryCount = retryCount
		update.NextRetryAt = gtime.Now().Add(time.Duration(delayMinutes) * time.Minute)
	} else {
		// 不再重试，标记为失败
		update.Status = OutboxStatusFailed
	}

	dao.MqOutbox.Ctx(updateCtx).Where("id", id).Update(update)
}

func formatGTime(t *gtime.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("Y-m-d H:i:s")
}

func isTableNotFoundErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "doesn't exist") || strings.Contains(msg, "error 1146") || strings.Contains(msg, "no such table")
}
