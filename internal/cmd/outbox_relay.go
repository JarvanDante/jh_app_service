package cmd

import (
	"context"
	"time"

	"jh_app_service/internal/logic/mq_outbox"
	"jh_app_service/internal/mq"
	"jh_app_service/internal/registry"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
)

var (
	OutboxRelay = &gcmd.Command{
		Name:  "outbox-relay",
		Usage: "outbox-relay [--site=site_code]",
		Brief: "启动Outbox中继，将mq_outbox消息投递到MQ",
		Description: `
启动Outbox中继，轮询各站点数据库的 mq_outbox 表，
将待发送消息投递到MQ，并更新发送状态。

参数：
  --site  指定站点代码，如果不指定则处理所有站点

示例：
  ./jh_app_service outbox-relay                    # 处理所有站点
  ./jh_app_service outbox-relay --site=jh          # 只处理 jh 站点
`,
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			siteCode := parser.GetOpt("site").String()

			if siteCode != "" {
				g.Log().Infof(ctx, "Outbox Relay 启动（站点: %s）...", siteCode)
			} else {
				g.Log().Info(ctx, "Outbox Relay 启动（所有站点）...")
			}

			// 初始化 Consul 客户端
			consulClient := registry.InitConsul()

			// 初始化数据库配置管理器
			dbManager := registry.GetDBConfigManagerV2()
			if err := dbManager.Init(consulClient); err != nil {
				g.Log().Fatalf(ctx, "Failed to init db config manager: %v", err)
			}
			defer dbManager.Stop()

			// 初始化 RabbitMQ
			if err := mq.InitRabbitMQ(ctx); err != nil {
				g.Log().Fatalf(ctx, "Failed to init RabbitMQ: %v", err)
			}
			defer mq.CloseRabbitMQ()

			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()

			// 立即执行一次
			if err := mq_outbox.MqOutbox().ProcessPendingMessagesForSite(ctx, siteCode); err != nil {
				g.Log().Errorf(ctx, "处理待发送消息失败: %v", err)
			}

			for {
				select {
				case <-ctx.Done():
					g.Log().Info(ctx, "Outbox Relay 停止")
					return nil
				case <-ticker.C:
					if err := mq_outbox.MqOutbox().ProcessPendingMessagesForSite(ctx, siteCode); err != nil {
						g.Log().Errorf(ctx, "处理待发送消息失败: %v", err)
					}
				}
			}
		},
	}
)

func init() {
	Main.AddCommand(OutboxRelay)
}
