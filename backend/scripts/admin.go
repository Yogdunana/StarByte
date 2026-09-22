package main

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	authrepo "github.com/Yogdunana/StarByte/backend/internal/auth/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	pkgredis "github.com/Yogdunana/StarByte/backend/pkg/redis"
	"github.com/Yogdunana/StarByte/backend/pkg/utils"
	"gorm.io/gorm"
)

// 管理员维护模式。
//
// 全量种子（make seed / 容器里默认执行）只负责「首次把 admin 建出来」；日常改账号
// 或改口令由本文件的模式承担，避免为了改一个口令把整批种子重跑一遍。
//
// 调用方是运维 CLI（starbyte admin），走一次性容器：
//
//	docker compose run --rm --entrypoint /usr/local/bin/starbyte-seed backend -mode=set-admin-password
//
// 输出刻意用 ASCII 的 KEY=VALUE 行，CLI 用 grep 就能解析，不依赖 jq。

const (
	adminModeSeed        = "seed"
	adminModeShow        = "show-admin"
	adminModeSetPassword = "set-admin-password"
	adminModeSetUsername = "set-admin-username"

	// 管理员身份由角色判定，不靠用户名：用户名可以被运维改掉，
	// 而 super_admin 的授予才是「这个人能进后台」的事实来源。
	adminRoleCode = "super_admin"

	adminGeneratedPasswordLen = 24
	adminUsernameMaxLen       = 50

	// 新值的传入通道（优先级 -value > 环境变量 > stdin > 自动生成）。
	// 走环境变量而不是命令行参数，是为了让口令不出现在容器 command 里
	// （command 会进 docker inspect / compose ps 的输出）。
	adminPasswordEnvVar = "STARBYTE_ADMIN_NEW_PASSWORD"
	adminUsernameEnvVar = "STARBYTE_ADMIN_NEW_USERNAME"

	// AuthRepo 只有 SetLockout，没有 Clear。用极短 TTL 让锁自然过期，
	// 不改 pkg 接口、也不在脚本里复刻 Redis key 前缀。
	adminLockoutClearTTL = 2 * time.Second
)

var errAdminNotFound = errors.New("未找到该管理员账号")

// adminAccount 是 starbyte admin show 的一行。
type adminAccount struct {
	ID          string
	Username    string
	RealName    string
	Status      int
	LastLoginAt *time.Time
	CreatedAt   time.Time
}

// runAdminMode 分发非种子模式。种子模式（SeedAll）不经过这里，行为完全不变。
func runAdminMode(db *gorm.DB, cfg *config.Config, mode, username, value string) error {
	switch mode {
	case adminModeShow:
		return showAdmin(db, username)
	case adminModeSetPassword:
		return setAdminPassword(db, cfg, username, value)
	case adminModeSetUsername:
		return setAdminUsername(db, cfg, username, value)
	default:
		return fmt.Errorf("未知的 -mode: %s（可选 %s / %s / %s / %s）",
			mode, adminModeSeed, adminModeShow, adminModeSetPassword, adminModeSetUsername)
	}
}

// showAdmin 列出全部 super_admin。指定 username 时只看这一个账号（不校验角色，
// 便于运维查看"改过名之后"的账号）。
func showAdmin(db *gorm.DB, username string) error {
	accounts, err := listAdmins(db, username)
	if err != nil {
		return err
	}
	if len(accounts) == 0 {
		return errAdminNotFound
	}
	for _, a := range accounts {
		fmt.Printf("ADMIN_USERNAME=%s\n", a.Username)
		fmt.Printf("ADMIN_ID=%s\n", a.ID)
		fmt.Printf("ADMIN_STATUS=%d\n", a.Status)
		fmt.Printf("ADMIN_CREATED_AT=%s\n", a.CreatedAt.Format("2006-01-02 15:04:05"))
		if a.LastLoginAt == nil {
			fmt.Printf("ADMIN_LAST_LOGIN=-\n")
		} else {
			fmt.Printf("ADMIN_LAST_LOGIN=%s\n", a.LastLoginAt.Format("2006-01-02 15:04:05"))
		}
	}
	return nil
}

func listAdmins(db *gorm.DB, username string) ([]adminAccount, error) {
	query := `
		SELECT u.id::text, u.username, COALESCE(u.real_name, ''), u.status,
		       u.last_login_at, u.created_at
		FROM users u
		JOIN user_roles ur ON ur.user_id = u.id
		JOIN roles r ON r.id = ur.role_id
		WHERE r.code = ? AND u.deleted_at IS NULL`
	args := []interface{}{adminRoleCode}
	if strings.TrimSpace(username) != "" {
		query += ` AND u.username = ?`
		args = append(args, username)
	}
	query += ` ORDER BY u.created_at`

	rows, err := db.Raw(query, args...).Rows()
	if err != nil {
		return nil, fmt.Errorf("query admins: %w", err)
	}
	// errcheck 不放过 *sql.Rows.Close（它返回 error，且不匹配配置里
	// 排除的 io.Closer.Close），显式丢弃。
	defer func() { _ = rows.Close() }()

	var out []adminAccount
	for rows.Next() {
		var a adminAccount
		var lastLogin sql.NullTime
		if err := rows.Scan(&a.ID, &a.Username, &a.RealName, &a.Status, &lastLogin, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan admin: %w", err)
		}
		if lastLogin.Valid {
			t := lastLogin.Time
			a.LastLoginAt = &t
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// setAdminPassword 改口令并吊销该账号全部在线会话。
//
// 必须吊销：应急改密最常见的起因就是「怀疑口令已经泄露」，只改库不动会话的话
// 攻击者手里的 refresh token 还能续期（最长 7 天），改密等于没改。
func setAdminPassword(db *gorm.DB, cfg *config.Config, username, explicit string) error {
	id, err := lookupAdminID(db, username)
	if err != nil {
		return err
	}

	password, generated, err := resolveNewPassword(explicit)
	if err != nil {
		return err
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}
	// id 是 UUID 列：参数以字符串传入时必须显式 ::uuid，
	// 否则 Postgres 报 "operator does not exist: uuid = text"。
	res := db.Exec(
		`UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?::uuid`,
		hash, id,
	)
	if res.Error != nil {
		return fmt.Errorf("update admin password: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return errAdminNotFound
	}

	revoked, revokeErr := revokeAdminSessions(cfg, username, id)
	if revokeErr != nil {
		// 口令已经改成功了，不要让调用方以为失败而重试一个已生效的操作。
		// 但旧 token 在剩余 TTL 内仍然有效，必须让运维看得见。
		fmt.Fprintf(os.Stderr, "starbyte-seed: 会话吊销失败，旧令牌在过期前仍然有效: %v\n", revokeErr)
	}

	fmt.Printf("ADMIN_USERNAME=%s\n", username)
	if generated {
		// 只在自动生成时回显；运维自己指定的值不再原样打出来。
		fmt.Printf("ADMIN_PASSWORD=%s\n", password)
	} else {
		fmt.Printf("ADMIN_PASSWORD_SOURCE=explicit\n")
	}
	fmt.Printf("ADMIN_SESSIONS_REVOKED=%d\n", revoked)
	return nil
}

// resolveNewPassword 按「-value > 环境变量 > stdin > 自动生成」确定新口令。
func resolveNewPassword(explicit string) (password string, generated bool, err error) {
	if v := strings.TrimSpace(explicit); v != "" {
		return v, false, nil
	}
	if v := strings.TrimSpace(os.Getenv(adminPasswordEnvVar)); v != "" {
		return v, false, nil
	}
	if v := readLineFromStdin(); v != "" {
		return v, false, nil
	}
	// 最后兜底：生成 24 位随机口令。极低概率全字母或全数字，过不了强度校验就重来。
	for i := 0; i < 5; i++ {
		p, genErr := genRandomPassword(adminGeneratedPasswordLen)
		if genErr != nil {
			return "", false, fmt.Errorf("generate admin password: %w", genErr)
		}
		if utils.ValidatePasswordStrength(p) {
			return p, true, nil
		}
	}
	return "", false, errors.New("连续 5 次生成的随机口令未通过强度校验")
}

// readLineFromStdin 只在 stdin 是管道/文件时读取（避免交互式终端卡住等待输入）。
func readLineFromStdin() string {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return ""
	}
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return ""
	}
	sc := bufio.NewScanner(os.Stdin)
	if sc.Scan() {
		return strings.TrimSpace(sc.Text())
	}
	return ""
}

// setAdminUsername 改管理员账号名。旧名字上的失败计数与锁定一并清掉，
// 否则「因为连续输错被锁 → 改用户名」会带着一把锁过去。
func setAdminUsername(db *gorm.DB, cfg *config.Config, oldName, newName string) error {
	target := strings.TrimSpace(newName)
	if v := strings.TrimSpace(os.Getenv(adminUsernameEnvVar)); target == "" {
		target = v
	}
	if err := validateAdminUsername(target); err != nil {
		return err
	}

	id, err := lookupAdminID(db, oldName)
	if err != nil {
		return err
	}

	var occupied int64
	if err := db.Raw(
		`SELECT count(*) FROM users WHERE username = ? AND id <> ?::uuid AND deleted_at IS NULL`,
		target, id,
	).Scan(&occupied).Error; err != nil {
		return fmt.Errorf("check username: %w", err)
	}
	if occupied > 0 {
		return fmt.Errorf("账号名 %q 已被占用", target)
	}

	res := db.Exec(
		`UPDATE users SET username = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?::uuid AND deleted_at IS NULL`,
		target, id,
	)
	if res.Error != nil {
		return fmt.Errorf("update admin username: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return errAdminNotFound
	}

	// 用户名变了，旧令牌里的身份信息已经对不上，一律吊销。
	revoked, revokeErr := revokeAdminSessions(cfg, oldName, id)
	if revokeErr != nil {
		fmt.Fprintf(os.Stderr, "starbyte-seed: 会话吊销失败，旧令牌在过期前仍然有效: %v\n", revokeErr)
	}

	fmt.Printf("ADMIN_USERNAME=%s\n", target)
	fmt.Printf("ADMIN_USERNAME_OLD=%s\n", oldName)
	fmt.Printf("ADMIN_SESSIONS_REVOKED=%d\n", revoked)
	return nil
}

func validateAdminUsername(name string) error {
	if name == "" {
		return errors.New("新账号名为空")
	}
	if len(name) > adminUsernameMaxLen {
		return fmt.Errorf("新账号名超过 %d 字符", adminUsernameMaxLen)
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case r == '_' || r == '.' || r == '-' || r == '@':
		default:
			return fmt.Errorf("新账号名含非法字符 %q（允许字母、数字、_ . - @）", string(r))
		}
	}
	return nil
}

func lookupAdminID(db *gorm.DB, username string) (string, error) {
	var id string
	err := db.Raw(`SELECT id::text FROM users WHERE username = ? AND deleted_at IS NULL`, username).Scan(&id).Error
	if err != nil {
		return "", fmt.Errorf("find admin: %w", err)
	}
	if id == "" {
		return "", errAdminNotFound
	}
	return id, nil
}

// revokeAdminSessions 黑名单化全部 access token、删除会话与 refresh token，
// 并清掉该账号名上的失败计数与锁定。返回吊销的会话数。
func revokeAdminSessions(cfg *config.Config, username, userID string) (int, error) {
	if err := pkgredis.Init(&cfg.Redis); err != nil {
		return 0, fmt.Errorf("connect redis: %w", err)
	}
	defer func() { _ = pkgredis.Close() }()

	ctx := context.Background()
	ar := authrepo.NewAuthRepo(pkgredis.Client())

	sessions, err := ar.ListSessionsByUser(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("list sessions: %w", err)
	}
	ttl := time.Duration(cfg.JWT.AccessTokenExp) * time.Second
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	revoked := 0
	for _, s := range sessions {
		if err := ar.BlacklistToken(ctx, s.TokenID, ttl); err != nil {
			return revoked, fmt.Errorf("blacklist token: %w", err)
		}
		if err := ar.DeleteSession(ctx, s.TokenID); err != nil {
			return revoked, fmt.Errorf("delete session: %w", err)
		}
		revoked++
	}
	if err := ar.DeleteRefreshTokensByUser(ctx, userID); err != nil {
		return revoked, fmt.Errorf("delete refresh tokens: %w", err)
	}
	if err := ar.ResetLoginAttempts(ctx, username); err != nil {
		return revoked, fmt.Errorf("reset login attempts: %w", err)
	}
	if err := ar.SetLockout(ctx, username, adminLockoutClearTTL); err != nil {
		return revoked, fmt.Errorf("clear lockout: %w", err)
	}
	return revoked, nil
}
