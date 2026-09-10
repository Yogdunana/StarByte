CREATE TABLE admission_permission_refreshes (
    user_id UUID PRIMARY KEY REFERENCES users(id),
    revision UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE admission_reminders (
    application_id UUID NOT NULL REFERENCES member_applications(id),
    revision INTEGER NOT NULL,
    stage VARCHAR(40) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY(application_id, revision, stage)
);
INSERT INTO scheduler_tasks (name,code,cron_expr,timezone,handler_key,next_run_at)
SELECT '入会审批提醒与候补期检查','admission_maintenance','0 * * * * *','Asia/Shanghai','admission_maintenance',NOW()
WHERE NOT EXISTS (SELECT 1 FROM scheduler_tasks WHERE code='admission_maintenance' AND status<>2);
