// Package config 全局配置定义
// 日程管理模块 (Schedule Module) - 对应 GitHub Issue #78
//
// 配置来源: 环境变量
//   启动时从环境变量读取，缺失项使用默认值
//   生产环境建议配合 .env 文件或容器编排注入
package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config 全局配置
type Config struct {
	// Server HTTP 服务配置
	Server ServerConfig
	// Database 数据库配置
	Database DatabaseConfig
	// JWT 认证配置
	JWT JWTConfig
	// Reminder 提醒调度配置
	Reminder ReminderConfig
}

// ServerConfig HTTP 服务配置
type ServerConfig struct {
	// Port 监听端口，默认 8080
	Port int
	// Mode Gin 运行模式: debug / release / test
	Mode string
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	// Host 主机地址
	Host string
	// Port 端口，默认 3306
	Port int
	// User 用户名
	User string
	// Password 密码
	Password string
	// Name 数据库名称
	Name string
	// MaxOpenConns 最大打开连接数，默认 50
	MaxOpenConns int
	// MaxIdleConns 最大空闲连接数，默认 10
	MaxIdleConns int
	// ConnMaxLifetime 连接最大生命周期（分钟），默认 30
	ConnMaxLifetimeMin int
}

// DSN 构造 MySQL 连接字符串
// 格式: user:password@tcp(host:port)/dbname?parseTime=true&charset=utf8mb4
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.Name,
	)
}

// JWTConfig JWT 认证配置
type JWTConfig struct {
	// Secret 签名密钥，必填
	Secret string
	// ExpireHours token 有效期（小时），默认 24
	ExpireHours int
}

// ReminderConfig 提醒调度配置
type ReminderConfig struct {
	// Enabled 是否启用调度器
	Enabled bool
	// IntervalSeconds 扫描间隔（秒），默认 60
	IntervalSeconds int
	// BatchSize 单次扫描最大处理量，默认 100
	BatchSize int
}

// Load 从环境变量加载配置
//   缺失的非关键字段使用默认值
//   JWT.Secret 为空则 panic（安全要求）
func Load() *Config {
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnvInt("SERVER_PORT", 8080),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:               getEnv("DB_HOST", "127.0.0.1"),
			Port:               getEnvInt("DB_PORT", 3306),
			User:               getEnv("DB_USER", "root"),
			Password:           getEnv("DB_PASSWORD", ""),
			Name:               getEnv("DB_NAME", "schedule_service"),
			MaxOpenConns:       getEnvInt("DB_MAX_OPEN_CONNS", 50),
			MaxIdleConns:       getEnvInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetimeMin: getEnvInt("DB_CONN_MAX_LIFETIME", 30),
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", ""),
			ExpireHours: getEnvInt("JWT_EXPIRE_HOURS", 24),
		},
		Reminder: ReminderConfig{
			Enabled:          getEnvBool("REMINDER_ENABLED", true),
			IntervalSeconds:  getEnvInt("REMINDER_INTERVAL_SECONDS", 60),
			BatchSize:        getEnvInt("REMINDER_BATCH_SIZE", 100),
		},
	}

	// JWT Secret 强制校验
	if cfg.JWT.Secret == "" {
		panic("JWT_SECRET environment variable is required")
	}

	return cfg
}

// ----------------------------------------------------------------------------
// 环境变量读取辅助
// ----------------------------------------------------------------------------

// getEnv 读取字符串环境变量，缺失返回 fallback
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getEnvInt 读取整型环境变量，缺失或非法返回 fallback
func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

// getEnvBool 读取布尔环境变量，缺失或非法返回 fallback
//   接受: 1/0, true/false, TRUE/FALSE
func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	switch v {
	case "1", "true", "TRUE", "True":
		return true
	case "0", "false", "FALSE", "False":
		return false
	}
	return fallback
}
