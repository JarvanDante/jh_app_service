CREATE DATABASE IF NOT EXISTS game_analytics;

CREATE TABLE IF NOT EXISTS game_analytics.user
(
    event_id UInt64,
    site_id UInt32,
    user_id UInt64,
    inviter_id UInt64,
    username String,
    status UInt8,
    register_ip String,
    register_time DateTime,
    created_at DateTime,
    updated_at DateTime,
    pay_times UInt32
)
    ENGINE = ReplacingMergeTree(event_id)
    PARTITION BY toYYYYMM(created_at)
    ORDER BY (site_id, user_id);