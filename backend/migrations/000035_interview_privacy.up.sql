-- Internal interview assessments require a separate, scoped permission.
INSERT INTO permissions (id, name, code, resource, action, description, status)
VALUES (uuid_generate_v4(), '查看内部面试评估', 'interview_private:read',
        'interview_private', 'read', '仅相关审批负责人及明确授权的审计人员可查看；申请人不可查看本人评估', 0)
ON CONFLICT (code) DO NOTHING;

INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT uuid_generate_v4(), r.id, p.id,
       CASE WHEN r.code = 'president' THEN 'all'
            WHEN r.code = 'vice_president' THEN 'department_and_sub'
            ELSE 'department' END
FROM roles r CROSS JOIN permissions p
WHERE p.code = 'interview_private:read'
  AND r.code IN ('president', 'vice_president', 'minister')
ON CONFLICT (role_id, permission_id) DO NOTHING;

UPDATE notification_templates
SET body_template = '{{.real_name}}，你的面试结果为 {{.result}}。请在申请进度页查看后续安排。',
    variables_schema = '{"real_name":"string","result":"string"}'::jsonb,
    updated_at = NOW()
WHERE code = 'interview_result';

-- Preserve exact historical copies before removing internal assessments from
-- candidate-visible channels. No application API exposes this archive.
CREATE TABLE IF NOT EXISTS interview_privacy_archive (
 source_table TEXT NOT NULL,
 source_id UUID NOT NULL,
 original_data JSONB NOT NULL,
 archived_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 PRIMARY KEY (source_table, source_id)
);
INSERT INTO interview_privacy_archive (source_table, source_id, original_data)
SELECT 'notifications', n.id, to_jsonb(n) FROM notifications n
-- Fresh installations still have the original `type` column. Existing servers
-- may also have `category`, added by Notification AutoMigrate at startup.
-- Read optional fields through JSON so neither schema requires starting the app
-- before migrations, and preserve the exact original row before sanitizing it.
WHERE (to_jsonb(n)->>'category' = 'interview' OR to_jsonb(n)->>'type' = 'interview')
  AND n.title LIKE '面试结果：%' ON CONFLICT DO NOTHING;
UPDATE notifications n SET content = '你的面试结果已更新，请在申请进度页查看后续安排。'
WHERE EXISTS (SELECT 1 FROM interview_privacy_archive x
 WHERE x.source_table = 'notifications' AND x.source_id = n.id);

INSERT INTO interview_privacy_archive (source_table, source_id, original_data)
SELECT 'member_applications', a.id, to_jsonb(a) FROM member_applications a
WHERE EXISTS (SELECT 1 FROM interviews i WHERE i.application_id = a.id
 AND i.result_comment <> '' AND i.result_comment = a.review_comment) ON CONFLICT DO NOTHING;
UPDATE member_applications a SET review_comment = '面试结果已记录，请查看申请进度。'
WHERE EXISTS (SELECT 1 FROM interview_privacy_archive x
 WHERE x.source_table = 'member_applications' AND x.source_id = a.id);

INSERT INTO interview_privacy_archive (source_table, source_id, original_data)
SELECT 'member_application_histories', h.id, to_jsonb(h) FROM member_application_histories h
WHERE EXISTS (SELECT 1 FROM interviews i WHERE i.application_id = h.application_id
 AND i.result_comment <> '' AND i.result_comment = h.comment) ON CONFLICT DO NOTHING;
UPDATE member_application_histories h SET comment = '面试结果已记录。'
WHERE EXISTS (SELECT 1 FROM interview_privacy_archive x
 WHERE x.source_table = 'member_application_histories' AND x.source_id = h.id);
