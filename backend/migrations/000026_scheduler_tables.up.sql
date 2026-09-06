-- 000026_scheduler_tables.up.sql
-- Issue #73 定时任务调度

CREATE TABLE IF NOT EXISTS scheduler_tasks (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(100) NOT NULL,
    code            VARCHAR(64)  NOT NULL,
    cron_expr       VARCHAR(64)  NOT NULL DEFAULT '',
    run_at          TIMESTAMPTZ,
    timezone        VARCHAR(64)  NOT NULL DEFAULT 'Asia/Shanghai',
    handler_key     VARCHAR(64)  NOT NULL,
    payload         TEXT         NOT NULL DEFAULT '',
    depends_on      TEXT         NOT NULL DEFAULT '[]',
    shard_key       VARCHAR(64)  NOT NULL DEFAULT '',
    status          SMALLINT     NOT NULL DEFAULT 0,
    max_retries     INT          NOT NULL DEFAULT 3,
    timeout_sec     INT          NOT NULL DEFAULT 60,
    retry_count     INT          NOT NULL DEFAULT 0,
    next_run_at     TIMESTAMPTZ,
    last_run_at     TIMESTAMPTZ,
    last_status     VARCHAR(20)  NOT NULL DEFAULT '',
    created_by      UUID,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_scheduler_tasks_code ON scheduler_tasks(code);
CREATE INDEX IF NOT EXISTS idx_scheduler_tasks_status_next ON scheduler_tasks(status, next_run_at);
CREATE INDEX IF NOT EXISTS idx_scheduler_tasks_handler ON scheduler_tasks(handler_key);

CREATE TABLE IF NOT EXISTS scheduler_runs (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id       UUID         NOT NULL REFERENCES scheduler_tasks(id) ON DELETE CASCADE,
    scheduled_at  TIMESTAMPTZ  NOT NULL,
    started_at    TIMESTAMPTZ,
    finished_at   TIMESTAMPTZ,
    status        VARCHAR(20)  NOT NULL DEFAULT 'pending',
    attempt       INT          NOT NULL DEFAULT 1,
    worker_id     VARCHAR(128) NOT NULL DEFAULT '',
    error_text    TEXT         NOT NULL DEFAULT '',
    output        TEXT         NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_scheduler_runs_task ON scheduler_runs(task_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_scheduler_runs_status ON scheduler_runs(status);

CREATE TABLE IF NOT EXISTS scheduler_run_logs (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    run_id      UUID         NOT NULL REFERENCES scheduler_runs(id) ON DELETE CASCADE,
    level       VARCHAR(16)  NOT NULL DEFAULT 'info',
    line        TEXT         NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_scheduler_run_logs_run ON scheduler_run_logs(run_id, created_at);
