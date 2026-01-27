package public

import (
	"context"
	"fmt"
	"jh_app_service/internal/dao"
	"jh_app_service/internal/middleware"
	"jh_app_service/internal/mq"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// asyncSyncUserToGame 异步同步用户到游戏服务（快速路径）
func (s *sPublic) asyncSyncUserToGame(ctx context.Context, siteId int32, siteCode string, userId int64, username string) {
	go func(siteCode string, userId int64) {
		msg := &mq.SyncUserToGameMessage{
			SiteID:   siteId,
			SiteCode: siteCode,
			UserID:   userId,
			Username: username,
		}

		if err := mq.PublishSyncUserToGame(context.Background(), msg); err != nil {
			g.Log().Errorf(ctx, "异步发送同步用户到游戏服务消息失败（将由Outbox兜底）: %v", err)
		} else {
			// 发送成功，标记 Outbox 为已发送
			dbGroup := fmt.Sprintf("site_%s", siteCode)
			updateCtx := context.WithValue(context.Background(), "dynamic_db_group", dbGroup)
			updateCtx = context.WithValue(updateCtx, middleware.SiteCodeContextKey, siteCode)

			_, err := dao.MqOutbox.Ctx(updateCtx).
				Where("aggregate_type", "user").
				Where("aggregate_id", userId).
				Where("routing_key", "user.sync_to_game").
				Where("status", 0).
				Update(g.Map{"status": 1, "updated_at": gtime.Now()})
			if err != nil {
				g.Log().Errorf(updateCtx, "更新同步用户到游戏服务Outbox状态失败: %v", err)
			} else {
				g.Log().Infof(updateCtx, "快速路径同步用户到游戏服务成功: user_id=%d", userId)
			}
		}
	}(siteCode, userId)
}

// asyncUpdateInitBalanceOutbox 异步更新初始化余额Outbox状态（快速路径）
func (s *sPublic) asyncUpdateInitBalanceOutbox(ctx context.Context, siteCode string, userId int64) {
	dbGroup := fmt.Sprintf("site_%s", siteCode)
	updateCtx := context.WithValue(context.Background(), "dynamic_db_group", dbGroup)
	updateCtx = context.WithValue(updateCtx, middleware.SiteCodeContextKey, siteCode)

	_, err := dao.MqOutbox.Ctx(updateCtx).
		Where("aggregate_type", "user").
		Where("aggregate_id", userId).
		Where("routing_key", "user.init_balance").
		Where("status", 0).
		Update(g.Map{"status": 1, "updated_at": gtime.Now()})
	if err != nil {
		g.Log().Errorf(updateCtx, "更新Outbox状态失败: %v", err)
	} else {
		g.Log().Infof(updateCtx, "快速路径发送成功，已更新Outbox状态: user_id=%d", userId)
	}
}
