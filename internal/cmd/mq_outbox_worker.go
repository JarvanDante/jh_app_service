package cmd

import (
	"context"
	"jh_app_service/internal/logic/mq_outbox"
	"jh_app_service/internal/mq"
	"jh_app_service/internal/registry"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
)

var (
	MqOutboxWorker = &gcmd.Command{
		Name:  "mq-outbox-worker",
		Usage: "mq-outbox-worker [--site=site_code]",
		Brief: "启动 MQ Outbox 消息处理器，定时扫描并发送待发送的消息",
		Description: `
启动 MQ Outbox Worker，定时扫描并发送待发送的消息。

参数：
  --site  指定站点代码，如果不指定则处理所有站点

示例：
  ./jh_app_service mq-outbox-worker                    # 处理所有站点
  ./jh_app_service mq-outbox-worker --site=jh          # 只处理 jh 站点
`,
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			// 获取站点参数
			siteCode := parser.GetOpt("site").String()

			if siteCode != "" {
				g.Log().Infof(ctx, "MQ Outbox Worker 启动（站点: %s）...", siteCode)
			} else {
				g.Log().Info(ctx, "MQ Outbox Worker 启动（所有站点）...")
			}

			// 初始化 Consul 客户端
			consulClient := registry.InitConsul()

			// 初始化数据库配置管理器
			dbManager := registry.GetDBConfigManagerV2()
			if err := dbManager.Init(consulClient); err != nil {
				g.Log().Fatalf(ctx, "Failed to init db config manager: %v", err)
			}
			defer dbManager.Stop()

			g.Log().Info(ctx, "数据库配置管理器初始化完成")

			// 初始化 RabbitMQ
			if err := mq.InitRabbitMQ(ctx); err != nil {
				g.Log().Fatalf(ctx, "Failed to init RabbitMQ: %v", err)
			}
			defer mq.CloseRabbitMQ()

			g.Log().Info(ctx, "RabbitMQ 初始化完成")

			// 定时任务：每30秒执行一次
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()

			// 立即执行一次
			if err := mq_outbox.MqOutbox().ProcessPendingMessagesForSite(ctx, siteCode); err != nil {
				g.Log().Errorf(ctx, "处理待发送消息失败: %v", err)
			}

			// 定时执行
			for {
				select {
				case <-ctx.Done():
					g.Log().Info(ctx, "MQ Outbox Worker 停止")
					return nil
				case <-ticker.C:
					if err := mq_outbox.MqOutbox().ProcessPendingMessagesForSite(ctx, siteCode); err != nil {
						g.Log().Errorf(ctx, "处理待发送消息失败: %v", err)
					}
				}
			}
		},
	}

	MqOutboxOnce = &gcmd.Command{
		Name:  "mq-outbox-once",
		Usage: "mq-outbox-once [--site=site_code]",
		Brief: "手动执行一次 MQ Outbox 消息处理",
		Description: `
手动执行一次 MQ Outbox 消息处理。

参数：
  --site  指定站点代码，如果不指定则处理所有站点

示例：
  ./jh_app_service mq-outbox-once                    # 处理所有站点
  ./jh_app_service mq-outbox-once --site=jh          # 只处理 jh 站点
`,
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			// 获取站点参数
			siteCode := parser.GetOpt("site").String()

			if siteCode != "" {
				g.Log().Infof(ctx, "开始处理站点 %s 的待发送消息...", siteCode)
			} else {
				g.Log().Info(ctx, "开始处理所有站点的待发送消息...")
			}

			// 初始化 Consul 客户端
			consulClient := registry.InitConsul()

			// 初始化数据库配置管理器
			dbManager := registry.GetDBConfigManagerV2()
			if err := dbManager.Init(consulClient); err != nil {
				g.Log().Fatalf(ctx, "Failed to init db config manager: %v", err)
			}
			defer dbManager.Stop()

			g.Log().Info(ctx, "数据库配置管理器初始化完成")

			// 初始化 RabbitMQ
			if err := mq.InitRabbitMQ(ctx); err != nil {
				g.Log().Fatalf(ctx, "Failed to init RabbitMQ: %v", err)
			}
			defer mq.CloseRabbitMQ()

			g.Log().Info(ctx, "RabbitMQ 初始化完成")

			if err := mq_outbox.MqOutbox().ProcessPendingMessagesForSite(ctx, siteCode); err != nil {
				g.Log().Errorf(ctx, "处理失败: %v", err)
				return err
			}

			g.Log().Info(ctx, "处理完成")
			return nil
		},
	}
)

func init() {
	// 注册命令到主命令
	Main.AddCommand(MqOutboxWorker, MqOutboxOnce)
}
