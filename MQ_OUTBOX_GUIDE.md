# MQ Outbox 模式使用指南

## 概述

MQ Outbox 模式通过**双重保障机制**确保消息可靠发送：

1. **快速路径**：注册成功后立即异步发送 MQ 消息（原有机制）
2. **兜底保障**：同时在数据库事务中记录 Outbox，如果快速路径失败，由定时任务兜底发送

## 架构流程

```
用户注册
  ↓
[事务开始]
  ├─ 插入 user 表
  ├─ 插入 mq_outbox 表 (status=0)
  └─ [事务提交]
  ↓
  ├─ 快速路径: goroutine 立即发送 MQ
  │   ├─ 成功 → 更新 mq_outbox status=1 ✅
  │   └─ 失败 → 记录日志，由 Outbox Worker 兜底 ⚠️
  │
  └─ 兜底保障: Outbox Worker 定时扫描
      ├─ 查询 status=0 的记录
      ├─ 发送到 RabbitMQ
      ├─ 成功: 更新 status=1
      └─ 失败: 更新 status=2, 设置重试时间
```

## 双重保障的优势

### 快速路径（原有机制）

✅ **低延迟** - 注册后立即发送，用户体验好  
✅ **高吞吐** - 不依赖定时任务，处理速度快  
✅ **简单直接** - 正常情况下直接成功

### 兜底保障（Outbox 模式）

✅ **可靠性** - 消息不会丢失，存储在数据库中  
✅ **自动重试** - 失败后指数退避重试  
✅ **可监控** - 可以查询消息状态和失败原因  
✅ **可恢复** - 进程重启后继续处理未发送的消息

### 组合效果

- **正常情况**：快速路径立即发送，Outbox 记录被标记为已发送，不会重复处理
- **MQ 短暂不可用**：快速路径失败，Outbox Worker 在30秒内重试
- **MQ 长时间不可用**：消息累积在 Outbox，恢复后自动发送
- **进程重启**：未发送的消息不会丢失，重启后继续处理

## 数据库表结构

已创建 `mq_outbox` 表，包含以下字段：

- `id`: 主键
- `event_type`: 事件类型（如 user_created）
- `aggregate_type`: 聚合根类型（如 user）
- `aggregate_id`: 聚合根ID（如 user_id）
- `exchange`: MQ exchange
- `routing_key`: MQ routing key
- `payload`: 消息体（JSON）
- `status`: 状态（0=待发送, 1=已发送, 2=发送失败）
- `retry_count`: 重试次数
- `next_retry_at`: 下次重试时间
- `last_error`: 最后一次错误信息

## 使用方法

### 1. 创建数据库表

```bash
cd jh_app_service
mysql -u root -p your_database < database/migrations/create_mq_outbox_table.sql
```

### 2. 启动 MQ Outbox Worker（定时任务）

```bash
# 方式1: 后台运行（推荐生产环境）
nohup ./jh_app_service mq-outbox-worker > logs/mq_outbox_worker.log 2>&1 &

# 方式2: 前台运行（开发调试）
./jh_app_service mq-outbox-worker
```

Worker 会每 30 秒扫描一次待发送的消息。

### 3. 手动执行一次（测试用）

```bash
./jh_app_service mq-outbox-once
```

## 重试策略

消息发送失败后会自动重试，采用指数退避策略：

- 第1次重试: 1分钟后
- 第2次重试: 5分钟后
- 第3次重试: 15分钟后
- 第4次重试: 30分钟后
- 第5次及以后: 1小时后

## 监控和排查

### 查询待发送的消息

```sql
SELECT * FROM mq_outbox WHERE status = 0 ORDER BY created_at DESC;
```

### 查询发送失败的消息

```sql
SELECT * FROM mq_outbox
WHERE status = 2
ORDER BY retry_count DESC, created_at DESC
LIMIT 20;
```

### 查询某个用户的消息状态

```sql
SELECT * FROM mq_outbox
WHERE aggregate_type = 'user'
  AND aggregate_id = '123'
ORDER BY created_at DESC;
```

### 手动重置失败消息

```sql
-- 重置为待发送状态
UPDATE mq_outbox
SET status = 0,
    retry_count = 0,
    next_retry_at = NOW()
WHERE id = 123;
```

## 代码示例

### 在事务中创建 Outbox 记录

```go
err = dao.User.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
    // 1. 插入用户
    result, err := dao.User.Ctx(ctx).TX(tx).Insert(user)
    if err != nil {
        return err
    }

    userId, _ := result.LastInsertId()

    // 2. 创建 Outbox 记录
    err = logic_mq_outbox.MqOutbox().CreateInitBalanceOutbox(
        ctx, tx, siteId, userId, username, currency,
    )
    if err != nil {
        return err
    }

    return nil
})
```

## 生产环境部署建议

### 1. 使用 Supervisor 管理进程

创建 `/etc/supervisor/conf.d/jh_app_mq_outbox.conf`:

```ini
[program:jh_app_mq_outbox_worker]
command=/path/to/jh_app_service mq-outbox-worker
directory=/path/to/jh_app_service
user=www-data
autostart=true
autorestart=true
redirect_stderr=true
stdout_logfile=/var/log/jh_app_mq_outbox_worker.log
```

启动：

```bash
supervisorctl reread
supervisorctl update
supervisorctl start jh_app_mq_outbox_worker
```

### 2. 使用 Systemd 管理

创建 `/etc/systemd/system/jh_app_mq_outbox.service`:

```ini
[Unit]
Description=JH App MQ Outbox Worker
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/path/to/jh_app_service
ExecStart=/path/to/jh_app_service mq-outbox-worker
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

启动：

```bash
systemctl daemon-reload
systemctl enable jh_app_mq_outbox
systemctl start jh_app_mq_outbox
systemctl status jh_app_mq_outbox
```

## 性能优化

1. **批量处理**: 每次最多处理 100 条消息
2. **索引优化**: 已创建 `idx_status_retry` 索引
3. **定期清理**: 建议定期清理已发送超过 30 天的记录

```sql
-- 清理30天前已发送的消息
DELETE FROM mq_outbox
WHERE status = 1
  AND created_at < DATE_SUB(NOW(), INTERVAL 30 DAY)
LIMIT 1000;
```

## 故障恢复

### 场景1: RabbitMQ 短暂不可用（几秒到几分钟）

- **快速路径**：发送失败，记录日志
- **兜底保障**：Outbox Worker 在30秒内重试
- **用户体验**：注册成功，余额初始化可能延迟几十秒

### 场景2: RabbitMQ 长时间不可用（几小时）

- **快速路径**：持续失败
- **兜底保障**：消息累积在 `mq_outbox` 表中，Worker 持续重试
- **恢复后**：RabbitMQ 恢复后，所有消息自动发送
- **结果**：不会丢失任何消息

### 场景3: 应用进程重启

- **快速路径**：重启时的消息可能丢失
- **兜底保障**：Outbox 中的消息不受影响，重启后继续处理
- **结果**：所有消息最终都会发送

## 注意事项

1. ✅ **双重保障不会重复发送** - 快速路径成功后会标记 Outbox 为已发送
2. ✅ 消息发送使用 `context.Background()`，不受请求超时影响
3. ✅ 消息持久化（`DeliveryMode: amqp.Persistent`）
4. ✅ 消息带 `MessageId` 用于去重
5. ✅ 事务保证数据一致性
6. ⚠️ 需要定期清理已发送的历史记录
7. ⚠️ 正常情况下，大部分消息由快速路径发送，Outbox 只是兜底
