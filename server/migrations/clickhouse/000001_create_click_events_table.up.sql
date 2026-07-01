CREATE DATABASE IF NOT EXISTS link4it;

CREATE TABLE IF NOT EXISTS link4it.click_events (
    link_id    UUID,
    timestamp  DateTime DEFAULT now(),
    ip         String,
    user_agent String,
    referrer   String
) ENGINE = MergeTree()
ORDER BY (link_id, timestamp);
