CREATE TABLE IF NOT EXISTS mq_outbox (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  
  -- 业务维度
  event_type VARCHAR(64) NOT NULL COMMENT '事件类型，如 user_created / balance_transfer',
  aggregate_type VARCHAR(64) NOT NULL COMMENT '聚合根类型，如 user / balance',
  aggregate_id VARCHAR(64) NOT NULL COMMENT '聚合根ID，如 user_id / trade_no',
  
  -- MQ 相关
  exchange VARCHAR(128) NOT NULL COMMENT 'MQ exchange',
  routing_key VARCHAR(128) NOT NULL COMMENT 'MQ routing key',
  
  -- 消息体
  payload JSON NOT NULL COMMENT '发送到 MQ 的消息体',
  
  -- 状态控制
  status TINYINT NOT NULL DEFAULT 0 COMMENT '状态: 0=待发送, 1=已发送, 2=发送失败',
  retry_count INT NOT NULL DEFAULT 0 COMMENT '重试次数',
  next_retry_at DATETIME DEFAULT NULL COMMENT '下次重试时间',
  
  -- 错误记录
  last_error VARCHAR(255) DEFAULT NULL COMMENT '最后一次发送错误信息',
  
  -- 时间
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (id),
  KEY idx_status_retry (status, next_retry_at),
  KEY idx_event (event_type),
  KEY idx_aggregate (aggregate_type, aggregate_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='MQ Outbox 表（可靠消息）';
