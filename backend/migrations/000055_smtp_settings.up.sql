-- ============================================================
-- 000055_smtp_settings.up.sql
-- SMTP 非密钥设置写入 configs，密码只走环境变量 STARBYTE_SMTP_PASSWORD。
-- ============================================================

INSERT INTO configs (id, config_key, config_value, config_type, description, category, is_public)
VALUES (
    uuid_generate_v4(),
    'smtp_settings',
    '{"host":"smtp.exmail.qq.com","port":465,"ssl_mode":"implicit","from":"computerassociation@smbu.edu.cn","from_name":"StarByte-SMTP","username":"computerassociation@smbu.edu.cn"}',
    'json',
    'SMTP 发信设置（不含密码；密码用 STARBYTE_SMTP_PASSWORD）',
    'notification',
    false
)
ON CONFLICT (config_key) DO NOTHING;
