package main

import (
	"fmt"

	"gorm.io/gorm"
)

type seedConfig struct {
	Key         string
	Value       string
	Type        string
	Category    string
	Description string
}

var seedConfigsData = []seedConfig{
	{Key: "site.name", Value: "深圳北理莫斯科大学计算机协会", Type: "string", Category: "system", Description: "协会显示名称"},
	{Key: "site.allow_register", Value: "true", Type: "boolean", Category: "security", Description: "是否开放注册"},
	{Key: "notification.email_enabled", Value: "false", Type: "boolean", Category: "notification", Description: "是否启用邮件通知"},
	{Key: "smtp_settings", Value: `{"host":"smtp.exmail.qq.com","port":465,"ssl_mode":"implicit","from":"computerassociation@smbu.edu.cn","from_name":"StarByte-SMTP","username":"computerassociation@smbu.edu.cn"}`, Type: "json", Category: "notification", Description: "SMTP 发信设置（不含密码；密码用 STARBYTE_SMTP_PASSWORD）"},
}

func seedRuntimeConfigs(db *gorm.DB) error {
	for _, c := range seedConfigsData {
		if err := db.Exec(`
			INSERT INTO configs (id, config_key, config_value, config_type, description, category, is_public)
			VALUES (uuid_generate_v4(), ?, ?, ?, ?, ?, false)
			ON CONFLICT (config_key) DO NOTHING`,
			c.Key, c.Value, c.Type, c.Description, c.Category,
		).Error; err != nil {
			return fmt.Errorf("seed config %s: %w", c.Key, err)
		}
	}
	return nil
}
