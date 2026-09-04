-- ============================================================
-- 000015_stats_tables.up.sql
-- 统计快照（Issue #19 原 000008）
-- ============================================================

CREATE TABLE stats_snapshots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    snapshot_date DATE NOT NULL,
    metric_key VARCHAR(100) NOT NULL,
    metric_value JSONB NOT NULL DEFAULT '{}'::jsonb,
    dimension VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (snapshot_date, metric_key, dimension)
);

CREATE INDEX idx_stats_snapshots_date ON stats_snapshots(snapshot_date);
CREATE INDEX idx_stats_snapshots_key ON stats_snapshots(metric_key);
CREATE INDEX idx_stats_snapshots_created_at ON stats_snapshots(created_at);
