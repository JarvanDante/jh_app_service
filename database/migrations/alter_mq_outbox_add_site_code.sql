-- 添加 site_code 字段到 mq_outbox 表
ALTER TABLE mq_outbox 
ADD COLUMN site_code VARCHAR(32) NOT NULL DEFAULT '' COMMENT '站点代码' AFTER aggregate_id,
ADD INDEX idx_site_code (site_code);
