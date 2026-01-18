package cmd

import (
	"context"
	"fmt"
	"jh_app_service/internal/controller/backend/ad"
	"jh_app_service/internal/controller/backend/admin"
	"jh_app_service/internal/controller/backend/message"
	"jh_app_service/internal/controller/backend/notice"
	"jh_app_service/internal/controller/backend/option"
	"jh_app_service/internal/controller/backend/role"
	"jh_app_service/internal/controller/backend/site"
	"jh_app_service/internal/controller/backend/upload"
	"jh_app_service/internal/controller/backend/user"
	"jh_app_service/internal/controller/frontend/public"
	frontenduser "jh_app_service/internal/controller/frontend/user"
	"os"
	"time"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"google.golang.org/grpc"

	"jh_app_service/internal/middleware"
	"jh_app_service/internal/mq"
	"jh_app_service/internal/registry"
	"jh_app_service/internal/tracing"
)

var (
	// Main is the main command.
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start grpc server of jinhuang",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			// 强制输出到stdout确保能看到
			fmt.Println("=== jh_app_service 启动 ===")
			fmt.Printf("启动时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
			fmt.Printf("进程ID: %d\n", os.Getpid())
			fmt.Println("==============================")

			// 初始化Jaeger追踪
			cleanup, err := tracing.InitJaeger()
			if err != nil {
				fmt.Printf("初始化Jaeger失败: %v\n", err)
				g.Log().Errorf(ctx, "初始化Jaeger失败: %v", err)
			} else {
				fmt.Println("Jaeger追踪初始化成功")
				g.Log().Info(ctx, "Jaeger追踪初始化成功")
				defer cleanup()
			}

			// 初始化 Consul 客户端
			fmt.Println("初始化Consul客户端...")
			consulClient := registry.InitConsul()

			// 注册服务到 Consul
			fmt.Println("注册服务到Consul...")
			if err := registry.RegisterService(); err != nil {
				fmt.Printf("Consul注册失败: %v\n", err)
				g.Log().Fatalf(ctx, "register service failed: %v", err)
			}

			fmt.Println("Consul服务注册成功")

			// 初始化数据库配置管理器V2
			fmt.Println("初始化App服务数据库配置管理器V2...")
			dbManagerV2 := registry.GetDBConfigManagerV2()
			if err := dbManagerV2.Init(consulClient); err != nil {
				fmt.Printf("App服务数据库配置管理器V2初始化失败: %v\n", err)
				g.Log().Fatalf(ctx, "init app db config manager v2 failed: %v", err)
			}
			fmt.Println("App服务数据库配置管理器V2初始化成功")

			// 初始化RabbitMQ
			fmt.Println("初始化RabbitMQ...")
			if err := mq.InitRabbitMQ(ctx); err != nil {
				fmt.Printf("RabbitMQ初始化失败: %v\n", err)
				g.Log().Errorf(ctx, "init rabbitmq failed: %v", err)
			} else {
				fmt.Println("RabbitMQ初始化成功")
				defer mq.CloseRabbitMQ()

				// 启动消费者
				fmt.Println("启动RabbitMQ消费者...")
				if err := mq.StartConsumer(ctx); err != nil {
					fmt.Printf("启动RabbitMQ消费者失败: %v\n", err)
					g.Log().Errorf(ctx, "start rabbitmq consumer failed: %v", err)
				} else {
					fmt.Println("RabbitMQ消费者启动成功")
				}
			}

			// 确保在退出时停止数据库配置管理器
			defer dbManagerV2.Stop()

			// 创建数据库拦截器
			dbInterceptor := middleware.NewDBInterceptor()

			c := grpcx.Server.NewConfig()
			c.Options = append(c.Options, []grpc.ServerOption{
				// 使用 StatsHandler 替代 Interceptor 进行追踪和统计
				grpc.StatsHandler(middleware.NewTraceStatsHandler()),
				// 添加数据库拦截器和验证拦截器
				grpcx.Server.ChainUnary(
					dbInterceptor.UnaryServerInterceptor(),
					grpcx.Server.UnaryValidate,
				),
				grpcx.Server.ChainStream(
					dbInterceptor.StreamServerInterceptor(),
				),
			}...)
			s := grpcx.Server.New(c)
			/*---------后台---------*/
			admin.Register(s)
			site.Register(s)
			role.Register(s)
			upload.Register(s)
			user.Register(s)
			message.Register(s)
			ad.Register(s)
			notice.Register(s)
			option.Register(s)

			/*---------前台---------*/
			public.Register(s)
			frontenduser.Register(s)

			fmt.Println("gRPC服务器启动中...")

			// 强制刷新输出
			os.Stdout.Sync()

			s.Run()
			return nil
		},
	}
)
