package mq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	balanceV1 "jh_app_service/api/grpc/balance/v1"
	"jh_app_service/internal/registry"
)

// StartConsumer 启动消费者
func StartConsumer(ctx context.Context) error {
	if channel == nil {
		return fmt.Errorf("RabbitMQ channel未初始化")
	}

	msgs, err := channel.Consume(
		"init_balance_user", // queue
		"",                  // consumer
		false,               // auto-ack
		false,               // exclusive
		false,               // no-local
		false,               // no-wait
		nil,                 // args
	)
	if err != nil {
		return fmt.Errorf("注册消费者失败: %v", err)
	}

	g.Log().Info(ctx, "RabbitMQ消费者启动成功，等待消息...")

	// 启动消费协程
	go func() {
		for msg := range msgs {
			processMessage(ctx, msg)
		}
	}()

	return nil
}

// processMessage 处理消息
func processMessage(ctx context.Context, delivery amqp.Delivery) {
	var msg InitBalanceUserMessage
	err := json.Unmarshal(delivery.Body, &msg)
	if err != nil {
		g.Log().Errorf(ctx, "解析消息失败: %v", err)
		delivery.Nack(false, false) // 拒绝消息，不重新入队
		return
	}

	g.Log().Infof(ctx, "收到初始化余额消息: site_id=%d, user_id=%d, username=%s, currency=%s",
		msg.SiteID, msg.UserID, msg.Username, msg.Currency)

	// 调用balance服务
	err = callBalanceService(ctx, &msg)
	if err != nil {
		g.Log().Errorf(ctx, "调用balance服务失败: %v", err)
		// 重试3次后放弃
		if delivery.Redelivered {
			g.Log().Errorf(ctx, "消息已重试，放弃处理: site_id=%d, user_id=%d", msg.SiteID, msg.UserID)
			delivery.Nack(false, false) // 拒绝消息，不重新入队
		} else {
			delivery.Nack(false, true) // 拒绝消息，重新入队
		}
		return
	}

	// 确认消息
	delivery.Ack(false)
	g.Log().Infof(ctx, "初始化余额成功: site_id=%d, user_id=%d", msg.SiteID, msg.UserID)
}

// callBalanceService 调用balance服务
func callBalanceService(ctx context.Context, msg *InitBalanceUserMessage) error {
	// 从Consul获取balance服务地址
	balanceAddr, err := registry.GetServiceAddr("jh_balance_service")
	if err != nil {
		return fmt.Errorf("从Consul获取balance服务地址失败: %v", err)
	}

	g.Log().Infof(ctx, "balance服务地址: %s", balanceAddr)

	// 创建gRPC连接
	conn, err := grpc.NewClient(
		balanceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("连接balance服务失败: %v", err)
	}
	defer conn.Close()

	// 创建客户端
	client := balanceV1.NewBalanceClient(conn)

	// 添加metadata
	md := metadata.New(map[string]string{
		"site_id": fmt.Sprintf("%d", msg.SiteID),
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	// 调用InitBalanceUser
	req := &balanceV1.InitBalanceUserReq{
		SiteId:   msg.SiteID,
		UserId:   msg.UserID,
		Username: msg.Username,
		Currency: msg.Currency,
	}

	resp, err := client.InitBalanceUser(ctx, req)
	if err != nil {
		return fmt.Errorf("调用InitBalanceUser gRPC失败: %v", err)
	}

	g.Log().Infof(ctx, "Balance服务响应: success=%v, message=%s", resp.Success, resp.Message)

	if !resp.Success {
		return fmt.Errorf("初始化余额失败: %s", resp.Message)
	}

	return nil
}
