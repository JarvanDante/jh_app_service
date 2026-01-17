package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	conn    *amqp.Connection
	channel *amqp.Channel
)

// InitRabbitMQ 初始化RabbitMQ连接
func InitRabbitMQ(ctx context.Context) error {
	url := g.Cfg().MustGet(ctx, "rabbitmq.url").String()
	if url == "" {
		return fmt.Errorf("RabbitMQ URL未配置")
	}

	var err error
	conn, err = amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("连接RabbitMQ失败: %v", err)
	}

	channel, err = conn.Channel()
	if err != nil {
		return fmt.Errorf("创建RabbitMQ channel失败: %v", err)
	}

	// 声明队列
	_, err = channel.QueueDeclare(
		"init_balance_user", // 队列名称
		true,                // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return fmt.Errorf("声明队列失败: %v", err)
	}

	g.Log().Info(ctx, "RabbitMQ初始化成功")
	return nil
}

// CloseRabbitMQ 关闭RabbitMQ连接
func CloseRabbitMQ() {
	if channel != nil {
		channel.Close()
	}
	if conn != nil {
		conn.Close()
	}
}

// InitBalanceUserMessage 初始化余额用户消息
type InitBalanceUserMessage struct {
	SiteID   int32  `json:"site_id"`
	UserID   int32  `json:"user_id"`
	Username string `json:"username"`
	Currency string `json:"currency"`
}

// PublishInitBalanceUser 发布初始化余额用户消息
func PublishInitBalanceUser(ctx context.Context, msg *InitBalanceUserMessage) error {
	if channel == nil {
		return fmt.Errorf("RabbitMQ channel未初始化")
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	err = channel.PublishWithContext(
		ctx,
		"",                  // exchange
		"init_balance_user", // routing key
		false,               // mandatory
		false,               // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("发布消息失败: %v", err)
	}

	g.Log().Infof(ctx, "发布初始化余额消息: site_id=%d, user_id=%d, username=%s, currency=%s",
		msg.SiteID, msg.UserID, msg.Username, msg.Currency)

	return nil
}
