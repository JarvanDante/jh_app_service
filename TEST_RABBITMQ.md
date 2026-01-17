# RabbitMQ 异步调用 Balance 服务测试文档

## 实现说明

已完成快速注册接口通过 RabbitMQ 异步调用 balance 服务的功能。

### 实现内容

1. **RabbitMQ 工具类** (`internal/mq/rabbitmq.go`)

   - 初始化 RabbitMQ 连接
   - 声明队列：`init_balance_user`
   - 发布消息函数：`PublishInitBalanceUser`

2. **RabbitMQ 消费者** (`internal/mq/consumer.go`)

   - 启动消费者监听队列
   - 处理消息并调用 balance 服务的 gRPC 接口
   - 支持消息重试机制

3. **配置更新** (`manifest/config/config.yaml`)

   - 添加 RabbitMQ 连接配置
   - 添加 balance 服务地址配置

4. **服务初始化** (`internal/cmd/cmd.go`)

   - 启动时初始化 RabbitMQ 连接
   - 启动消费者
   - 优雅关闭

5. **快速注册接口更新** (`internal/logic/frontend/public/public.go`)
   - 用户注册成功后异步发送消息
   - 从 Consul KV 获取站点配置的 currency 字段
   - 构造消息并发送到 RabbitMQ 队列

### 消息流程

```
用户注册 -> QuickRegister接口
    ↓
创建用户成功
    ↓
异步goroutine:
  1. 获取站点配置(从Consul KV)
  2. 获取currency字段
  3. 构造InitBalanceUserMessage
  4. 发送到RabbitMQ队列
    ↓
RabbitMQ消费者:
  1. 接收消息
  2. 连接balance服务gRPC
  3. 调用InitBalanceUser接口
  4. 确认消息
```

### 消息结构

```go
type InitBalanceUserMessage struct {
    SiteID   int32  `json:"site_id"`   // 站点ID
    UserID   int32  `json:"user_id"`   // 用户ID
    Username string `json:"username"`  // 用户名
    Currency string `json:"currency"`  // 货币类型(从KV获取)
}
```

### 配置说明

**RabbitMQ 配置**:

```yaml
rabbitmq:
  url: "amqp://admin:admin123@172.19.0.36:5672/"
```

**Balance 服务地址**:

```yaml
grpc:
  balanceService: "jh_balance_service:8003"
```

### 测试步骤

1. 确保 RabbitMQ 服务运行正常
2. 确保 balance 服务运行正常
3. 启动 jh_app_service
4. 调用快速注册接口
5. 查看日志确认：
   - 消息发送成功
   - 消费者接收消息
   - 调用 balance 服务成功
   - 余额初始化成功

### 日志关键字

- `发布初始化余额消息`
- `RabbitMQ消费者启动成功`
- `收到初始化余额消息`
- `初始化余额成功`

### 错误处理

- 消息发送失败：记录错误日志，不影响注册流程
- 消费失败：支持重试一次，失败后拒绝消息
- Balance 服务调用失败：记录错误日志

### 依赖

- `github.com/rabbitmq/amqp091-go v1.10.0`
- `jh_balance_service` (本地模块依赖)
