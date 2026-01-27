# MQ Outbox 实现总结

## 已完成的工作

### 1. 数据库层

✅ **创建表结构**

- `database/migrations/create_mq_outbox_table.sql` - 数据库表定义

✅ **Model 层**

- `internal/model/entity/mq_outbox.go` - Entity 实体
- `internal/model/do/mq_outbox.go` - DO 对象

✅ **DAO 层**

- `internal/dao/internal/mq_outbox.go` - Internal DAO
- `internal/dao/mq_outbox.go` - 外部 DAO（支持动态数据库路由）

### 2. 业务逻辑层

✅ **Outbox 服务**

- `internal/logic/mq_outbox/mq_outbox.go`
  - `CreateInitBalanceOutbox()` - 创建 Outbox 记录
  - `ProcessPendingMessages()` - 处理待发送消息
  - `processMessage()` - 处理单条消息
  - `processUserCreatedMessage()` - 处理用户创建事件
  - `markAsFailed()` - 标记失败并设置重试

### 3. 注册逻辑修改

✅ **QuickRegister 方法**

- 使用事务同时插入 user 和 mq_outbox
- 移除了 goroutine 异步发送
- 保证数据一致性

✅ **Register 方法**

- 使用事务同时插入 user 和 mq_outbox
- 移除了 goroutine 异步发送
- 保证数据一致性

### 4. 命令行工具

✅ **Worker 命令**

- `internal/cmd/mq_outbox_worker.go`
  - `mq-outbox-worker` - 定时任务（每30秒）
  - `mq-outbox-once` - 手动执行一次

### 5. 文档

✅ **使用指南**

- `MQ_OUTBOX_GUIDE.md` - 完整的使用文档
- `MQ_OUTBOX_IMPLEMENTATION_SUMMARY.md` - 实现总结

## 使用流程

### 开发环境测试

```bash
# 1. 创建数据库表
cd jh_app_service
mysql -u root -p your_database < database/migrations/create_mq_outbox_table.sql

# 2. 启动主服务
./jh_app_service

# 3. 启动 Outbox Worker（新终端）
./jh_app_service mq-outbox-worker

# 4. 测试注册接口
# 注册成功后，会在 mq_outbox 表中看到 status=0 的记录
# Worker 会在30秒内处理并发送到 RabbitMQ
# 发送成功后 status 变为 1
```

### 生产环境部署

```bash
# 使用 Supervisor 或 Systemd 管理 Worker 进程
# 详见 MQ_OUTBOX_GUIDE.md
```

## 核心特性

### 1. 可靠性保证

- ✅ 事务保证：用户创建和消息记录在同一事务中
- ✅ 持久化：消息存储在数据库中，不会丢失
- ✅ 重试机制：失败后自动重试，指数退避
- ✅ 幂等性：消息带 MessageId，支持去重

### 2. 重试策略

```
第1次: 1分钟后
第2次: 5分钟后
第3次: 15分钟后
第4次: 30分钟后
第5次+: 1小时后
```

### 3. 状态管理

```
0 = 待发送 (Pending)
1 = 已发送 (Sent)
2 = 发送失败 (Failed)
```

## 监控要点

### 1. 待发送消息数量

```sql
SELECT COUNT(*) FROM mq_outbox WHERE status = 0;
```

如果数量持续增长，说明：

- RabbitMQ 可能不可用
- Worker 可能未启动
- 发送速度跟不上产生速度

### 2. 失败消息

```sql
SELECT * FROM mq_outbox
WHERE status = 2
ORDER BY retry_count DESC
LIMIT 10;
```

检查 `last_error` 字段了解失败原因。

### 3. 消息延迟

```sql
SELECT
    MIN(TIMESTAMPDIFF(SECOND, created_at, updated_at)) as min_delay,
    AVG(TIMESTAMPDIFF(SECOND, created_at, updated_at)) as avg_delay,
    MAX(TIMESTAMPDIFF(SECOND, created_at, updated_at)) as max_delay
FROM mq_outbox
WHERE status = 1
  AND created_at > DATE_SUB(NOW(), INTERVAL 1 HOUR);
```

## 维护建议

### 1. 定期清理历史数据

```sql
-- 每天执行，清理30天前已发送的消息
DELETE FROM mq_outbox
WHERE status = 1
  AND created_at < DATE_SUB(NOW(), INTERVAL 30 DAY)
LIMIT 10000;
```

### 2. 监控告警

建议设置以下告警：

- 待发送消息数 > 1000
- 失败消息数 > 100
- Worker 进程停止

### 3. 日志查看

```bash
# 查看 Worker 日志
tail -f logs/mq_outbox_worker.log

# 或使用 journalctl（如果用 systemd）
journalctl -u jh_app_mq_outbox -f
```

## 扩展建议

### 1. 支持更多事件类型

在 `internal/logic/mq_outbox/mq_outbox.go` 的 `processMessage()` 中添加新的 case：

```go
switch msg.EventType {
case "user_created":
    return s.processUserCreatedMessage(ctx, msg)
case "balance_transfer":  // 新增
    return s.processBalanceTransferMessage(ctx, msg)
default:
    return fmt.Errorf("未知的事件类型: %s", msg.EventType)
}
```

### 2. 支持优先级

在表中添加 `priority` 字段，处理时优先处理高优先级消息。

### 3. 支持分片

如果消息量很大，可以启动多个 Worker，每个处理不同的 ID 范围：

```sql
WHERE status = 0 AND id % 4 = 0  -- Worker 1
WHERE status = 0 AND id % 4 = 1  -- Worker 2
WHERE status = 0 AND id % 4 = 2  -- Worker 3
WHERE status = 0 AND id % 4 = 3  -- Worker 4
```

## 对比原方案

### 原方案（goroutine 异步发送）

❌ 问题：

- 如果 MQ 不可用，消息直接丢失
- 无法追踪消息状态
- 无法重试
- 进程重启时未发送的消息丢失

### 新方案（Outbox 模式）

✅ 优势：

- 消息持久化，不会丢失
- 可追踪每条消息的状态
- 自动重试机制
- 进程重启后继续处理
- 支持监控和告警

## 性能影响

- **写入性能**: 每次注册多一次 INSERT（可忽略）
- **读取性能**: 无影响
- **额外资源**: Worker 进程（CPU 和内存占用很小）

## 总结

MQ Outbox 模式已完整实现，提供了可靠的消息发送保证。建议在生产环境中：

1. ✅ 使用 Supervisor 或 Systemd 管理 Worker
2. ✅ 设置监控告警
3. ✅ 定期清理历史数据
4. ✅ 查看日志排查问题

所有代码已经修改完成，可以直接使用！
