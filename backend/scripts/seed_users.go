package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"

	"github.com/Yogdunana/StarByte/backend/pkg/utils"
	"gorm.io/gorm"
)

// genRandomPassword 生成 length 位、由大小写字母+数字组成的高熵口令。
// 仅使用字母与数字（无 shell 元字符），避免部署脚本回显或复制时被截断/注入。
func genRandomPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[n.Int64()]
	}
	return string(b), nil
}

func seedUsers(db *gorm.DB) error {
	if seedUsersUseDevPasswords() {
		return seedUsersDev(db)
	}
	return seedUsersProd(db)
}

// seedUsersUseDevPasswords 报告本次播种是否允许写入开发便利口令。
//
// 判定必须与 pkg/config 的语义对齐（fail-closed）：loader 在 APP_ENV 为空时会把
// effectiveEnv 补成 "prod"（按生产做密钥强度校验），只是不额外加载 config.prod.yaml。
// 若这里只判 APP_ENV == "prod"，就会出现「按生产校验密钥 + 连生产库 + 却写开发口令」
// 这种自相矛盾的组合：直接 `go run ./scripts`（绕过 Makefile）且只注入
// DB_* / JWT_SECRET、不导出 APP_ENV 时，会把 admin/admin123（super_admin）与
// test/test123 种进生产库，留下可登录的默认超管。
//
// 因此改为：只有显式 dev/test 才用开发口令，其余（含未设置）一律按生产处理。
// Makefile 的 seed 目标已显式带上 `APP_ENV=$(or $(APP_ENV),dev)`，本地开发不受影响。
func seedUsersUseDevPasswords() bool {
	switch os.Getenv("APP_ENV") {
	case "dev", "test":
		return true
	default:
		return false
	}
}

// seedUsersDev 保持原有开发口令（admin/admin123、test/test123），仅打印提示。
func seedUsersDev(db *gorm.DB) error {
	adminHash, err := utils.HashPassword("admin123")
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}
	testHash, err := utils.HashPassword("test123")
	if err != nil {
		return fmt.Errorf("hash test password: %w", err)
	}

	if err := db.Exec(`
		INSERT INTO users (id, username, password_hash, real_name, email, status, email_verified_at)
		VALUES (uuid_generate_v4(), 'admin', ?, '管理员', 'admin@starbyte.local', 0, CURRENT_TIMESTAMP)
		ON CONFLICT (username) WHERE deleted_at IS NULL DO NOTHING
	`, adminHash).Error; err != nil {
		return err
	}
	if err := db.Exec(`
		INSERT INTO users (id, username, password_hash, real_name, email, status, email_verified_at)
		VALUES (uuid_generate_v4(), 'test', ?, '测试会员', 'test@starbyte.local', 0, CURRENT_TIMESTAMP)
		ON CONFLICT (username) WHERE deleted_at IS NULL DO NOTHING
	`, testHash).Error; err != nil {
		return err
	}
	if err := db.Exec(`
		UPDATE users
		SET email_verified_at = COALESCE(email_verified_at, CURRENT_TIMESTAMP)
		WHERE username IN ('admin', 'test') AND deleted_at IS NULL AND email_verified_at IS NULL
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		INSERT INTO user_roles (id, user_id, role_id)
		SELECT uuid_generate_v4(), u.id, r.id
		FROM users u
		CROSS JOIN roles r
		WHERE u.username = 'admin' AND r.code = 'super_admin'
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

	fmt.Println("============================================================")
	fmt.Println("[StarByte] 开发环境已播种 admin/admin123 与 test/test123（仅用于开发，请勿在生产使用）")
	fmt.Println("============================================================")
	return nil
}

// seedUsersProd 生产播种：不写入测试账号；admin 口令优先取 SEED_ADMIN_PASSWORD，
// 否则生成 24 位强随机口令并打印一次（只显示这一次）。
func seedUsersProd(db *gorm.DB) error {
	// 测试账号仅用于开发，生产不创建。

	// 先判断 admin 是否已存在，已存在则跳过创建/不打印口令。
	var adminCount int64
	if err := db.Raw("SELECT count(*) FROM users WHERE username = 'admin' AND deleted_at IS NULL").Scan(&adminCount).Error; err != nil {
		return fmt.Errorf("check admin existence: %w", err)
	}

	// 口令来源：环境变量优先，否则生成强随机口令。
	adminPassword := os.Getenv("SEED_ADMIN_PASSWORD")
	generated := false
	if adminPassword == "" {
		p, err := genRandomPassword(24)
		if err != nil {
			return fmt.Errorf("generate admin password: %w", err)
		}
		adminPassword = p
		generated = true
	}
	adminHash, err := utils.HashPassword(adminPassword)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}

	if err := db.Exec(`
		INSERT INTO users (id, username, password_hash, real_name, email, status, email_verified_at)
		VALUES (uuid_generate_v4(), 'admin', ?, '管理员', 'admin@starbyte.local', 0, CURRENT_TIMESTAMP)
		ON CONFLICT (username) WHERE deleted_at IS NULL DO NOTHING
	`, adminHash).Error; err != nil {
		return err
	}
	if err := db.Exec(`
		UPDATE users
		SET email_verified_at = COALESCE(email_verified_at, CURRENT_TIMESTAMP)
		WHERE username = 'admin' AND deleted_at IS NULL AND email_verified_at IS NULL
	`).Error; err != nil {
		return err
	}
	if err := db.Exec(`
		INSERT INTO user_roles (id, user_id, role_id)
		SELECT uuid_generate_v4(), u.id, r.id
		FROM users u
		CROSS JOIN roles r
		WHERE u.username = 'admin' AND r.code = 'super_admin'
		ON CONFLICT (user_id, role_id) DO NOTHING
	`).Error; err != nil {
		return err
	}

	if adminCount > 0 {
		fmt.Println("============================================================")
		fmt.Println("[StarByte] 生产环境：admin 已存在，跳过创建（未打印口令）")
		fmt.Println("============================================================")
		return nil
	}

	if generated {
		fmt.Println("============================================================")
		fmt.Println("[StarByte] 生产环境首次播种：admin 账号已创建")
		fmt.Println("用户名: admin")
		fmt.Printf("初始密码: %s\n", adminPassword)
		fmt.Println("请立即登录并修改密码；本密码只显示这一次。")
		fmt.Println("============================================================")
	} else {
		fmt.Println("============================================================")
		fmt.Println("[StarByte] 生产环境首次播种：admin 账号已创建（口令来自 SEED_ADMIN_PASSWORD，未打印）")
		fmt.Println("============================================================")
	}
	return nil
}
