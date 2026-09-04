package main

import (
	"fmt"

	"github.com/Yogdunana/StarByte/backend/pkg/utils"
	"gorm.io/gorm"
)

func seedUsers(db *gorm.DB) error {
	adminHash, err := utils.HashPassword("admin123")
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}
	testHash, err := utils.HashPassword("test123")
	if err != nil {
		return fmt.Errorf("hash test password: %w", err)
	}

	if err := db.Exec(`
		INSERT INTO users (id, username, password_hash, real_name, email, status)
		VALUES (uuid_generate_v4(), 'admin', ?, '管理员', 'admin@starbyte.local', 0)
		ON CONFLICT (username) DO NOTHING
	`, adminHash).Error; err != nil {
		return err
	}
	if err := db.Exec(`
		INSERT INTO users (id, username, password_hash, real_name, email, status)
		VALUES (uuid_generate_v4(), 'test', ?, '测试会员', 'test@starbyte.local', 0)
		ON CONFLICT (username) DO NOTHING
	`, testHash).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		INSERT INTO user_roles (id, user_id, role_id)
		SELECT uuid_generate_v4(), u.id, r.id
		FROM users u
		CROSS JOIN roles r
		WHERE u.username = 'admin' AND r.code IN ('president', 'super_admin')
		ON CONFLICT (user_id, role_id) DO NOTHING
	`).Error; err != nil {
		return err
	}
	if err := db.Exec(`
		INSERT INTO user_roles (id, user_id, role_id)
		SELECT uuid_generate_v4(), u.id, r.id
		FROM users u
		JOIN roles r ON r.code = 'member'
		WHERE u.username = 'test'
		ON CONFLICT (user_id, role_id) DO NOTHING
	`).Error; err != nil {
		return err
	}
	return nil
}
