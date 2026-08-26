package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// validEnvs defines the set of accepted APP_ENV values.
var validEnvs = map[string]bool{
	"dev":  true,
	"test": true,
	"prod": true,
}

// Load reads the base YAML config, applies environment-specific overrides,
// injects environment variables for sensitive fields, sets defaults, and
// validates the result.
//
// The path argument points to the base config file (e.g. "configs/config.yaml").
// If the APP_ENV environment variable is set (dev/test/prod), a matching
// config.{APP_ENV}.yaml in the same directory is loaded and merged on top.
//
// Environment variables override individual fields after YAML merging.
// See applyEnvOverrides() for the full list of supported variables.
func Load(path string) (*Config, error) {
	// 1. Read and parse base config.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("unmarshal base config: %w", err)
	}

	// 2. Load environment-specific override (if APP_ENV is set).
	env := os.Getenv("APP_ENV")
	if env != "" {
		if !validEnvs[env] {
			return nil, fmt.Errorf("invalid APP_ENV %q, must be one of: dev, test, prod", env)
		}
		overridePath := filepath.Join(filepath.Dir(path), "config."+env+".yaml")
		if overrideData, err := os.ReadFile(overridePath); err == nil {
			if err := yaml.Unmarshal(overrideData, cfg); err != nil {
				return nil, fmt.Errorf("unmarshal %s config: %w", env, err)
			}
		}
		// If the override file doesn't exist, silently fall back to base config.
	}

	// 3. Inject environment variables (highest priority).
	applyEnvOverrides(cfg)

	// 4. Set defaults for any remaining zero-value fields.
	setDefaults(cfg)

	// 5. Validate critical fields.
	if err := validate(cfg, env); err != nil {
		return nil, err
	}

	return cfg, nil
}

// applyEnvOverrides overwrites sensitive config fields from environment
// variables. Environment variables take the highest priority — they override
// values from both the base YAML and environment-specific YAML.
//
// Supported environment variables:
//
//	APP_ENV              — selects config.{env}.yaml override
//	SERVER_PORT          — server.port
//	SERVER_MODE          — server.mode
//	DB_HOST              — database.host
//	DB_PORT              — database.port
//	DB_USER              — database.user
//	DB_PASSWORD          — database.password
//	DB_NAME              — database.dbname
//	DB_SSLMODE           — database.sslmode
//	DB_MAX_OPEN          — database.max_open
//	DB_MAX_IDLE          — database.max_idle
//	REDIS_HOST           — redis.host
//	REDIS_PORT           — redis.port
//	REDIS_PASSWORD       — redis.password
//	REDIS_DB             — redis.db
//	JWT_SECRET           — jwt.secret
//	JWT_ACCESS_TOKEN_EXP — jwt.access_token_exp
//	JWT_REFRESH_TOKEN_EXP — jwt.refresh_token_exp
//	MINIO_ENDPOINT       — minio.endpoint
//	MINIO_ACCESS_KEY     — minio.access_key
//	MINIO_SECRET_KEY     — minio.secret_key
//	MINIO_BUCKET         — minio.bucket
//	MINIO_USE_SSL        — minio.use_ssl
//	SMTP_HOST            — email.smtp_host
//	SMTP_PORT            — email.smtp_port
//	SMTP_USER            — email.username
//	SMTP_PASSWORD        — email.password
//	SMTP_FROM            — email.from
//	LOG_LEVEL            — logger.level
//	LOG_FORMAT           — logger.format
//	CORS_ALLOWED_ORIGINS  — cors.allowed_origins (comma-separated)
//	CORS_ALLOWED_METHODS  — cors.allowed_methods (comma-separated)
//	CORS_ALLOWED_HEADERS  — cors.allowed_headers (comma-separated)
//	CORS_EXPOSE_HEADERS   — cors.expose_headers (comma-separated)
//	CORS_ALLOW_CREDENTIALS — cors.allow_credentials (true/false)
func applyEnvOverrides(cfg *Config) {
	// Server
	cfg.Server.Port = getEnvInt("SERVER_PORT", cfg.Server.Port)
	cfg.Server.Mode = getEnv("SERVER_MODE", cfg.Server.Mode)

	// Database
	cfg.Database.Host = getEnv("DB_HOST", cfg.Database.Host)
	cfg.Database.Port = getEnvInt("DB_PORT", cfg.Database.Port)
	cfg.Database.User = getEnv("DB_USER", cfg.Database.User)
	cfg.Database.Password = getEnv("DB_PASSWORD", cfg.Database.Password)
	cfg.Database.DBName = getEnv("DB_NAME", cfg.Database.DBName)
	cfg.Database.SSLMode = getEnv("DB_SSLMODE", cfg.Database.SSLMode)
	cfg.Database.MaxOpen = getEnvInt("DB_MAX_OPEN", cfg.Database.MaxOpen)
	cfg.Database.MaxIdle = getEnvInt("DB_MAX_IDLE", cfg.Database.MaxIdle)

	// Redis
	cfg.Redis.Host = getEnv("REDIS_HOST", cfg.Redis.Host)
	cfg.Redis.Port = getEnvInt("REDIS_PORT", cfg.Redis.Port)
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", cfg.Redis.Password)
	cfg.Redis.DB = getEnvInt("REDIS_DB", cfg.Redis.DB)

	// JWT
	cfg.JWT.Secret = getEnv("JWT_SECRET", cfg.JWT.Secret)
	cfg.JWT.AccessTokenExp = getEnvInt("JWT_ACCESS_TOKEN_EXP", cfg.JWT.AccessTokenExp)
	cfg.JWT.RefreshTokenExp = getEnvInt("JWT_REFRESH_TOKEN_EXP", cfg.JWT.RefreshTokenExp)

	// MinIO
	cfg.MinIO.Endpoint = getEnv("MINIO_ENDPOINT", cfg.MinIO.Endpoint)
	cfg.MinIO.AccessKey = getEnv("MINIO_ACCESS_KEY", cfg.MinIO.AccessKey)
	cfg.MinIO.SecretKey = getEnv("MINIO_SECRET_KEY", cfg.MinIO.SecretKey)
	cfg.MinIO.Bucket = getEnv("MINIO_BUCKET", cfg.MinIO.Bucket)
	cfg.MinIO.UseSSL = getEnvBool("MINIO_USE_SSL", cfg.MinIO.UseSSL)

	// Email
	cfg.Email.SMTPHost = getEnv("SMTP_HOST", cfg.Email.SMTPHost)
	cfg.Email.SMTPPort = getEnvInt("SMTP_PORT", cfg.Email.SMTPPort)
	cfg.Email.Username = getEnv("SMTP_USER", cfg.Email.Username)
	cfg.Email.Password = getEnv("SMTP_PASSWORD", cfg.Email.Password)
	cfg.Email.From = getEnv("SMTP_FROM", cfg.Email.From)

	// Logger
	cfg.Logger.Level = getEnv("LOG_LEVEL", cfg.Logger.Level)
	cfg.Logger.Format = getEnv("LOG_FORMAT", cfg.Logger.Format)

	// CORS (comma-separated values)
	cfg.CORS.AllowedOrigins = getEnvCSV("CORS_ALLOWED_ORIGINS", cfg.CORS.AllowedOrigins)
	cfg.CORS.AllowedMethods = getEnvCSV("CORS_ALLOWED_METHODS", cfg.CORS.AllowedMethods)
	cfg.CORS.AllowedHeaders = getEnvCSV("CORS_ALLOWED_HEADERS", cfg.CORS.AllowedHeaders)
	cfg.CORS.ExposeHeaders = getEnvCSV("CORS_EXPOSE_HEADERS", cfg.CORS.ExposeHeaders)
	cfg.CORS.AllowCredentials = getEnvBool("CORS_ALLOW_CREDENTIALS", cfg.CORS.AllowCredentials)
}

// getEnv returns the value of an environment variable or a fallback.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getEnvInt returns the integer value of an environment variable or a fallback.
// If the environment variable is set but cannot be parsed as an integer, a
// warning is logged and the fallback is returned.
func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		log.Printf("config: ignoring invalid integer for %s=%q: %v", key, v, err)
		return fallback
	}
	return n
}

// getEnvBool returns the boolean value of an environment variable or a fallback.
// The value is considered true if it is "true" or "1" (case-sensitive).
func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v == "true" || v == "1"
}

// getEnvCSV returns the comma-separated value of an environment variable as a
// slice, or a fallback if the variable is not set.
func getEnvCSV(key string, fallback []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return splitCSV(v)
}

// splitCSV is a helper to parse comma-separated string values into a slice.
func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// setDefaults fills in sensible default values for any zero-value fields
// that were not provided by the YAML files or environment variables.
// This ensures the application can start with a minimal config file while
// still enforcing critical values via validate().
func setDefaults(cfg *Config) {
	// Server defaults
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug"
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 30
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 30
	}

	// Database defaults
	if cfg.Database.Port == 0 {
		cfg.Database.Port = 5432
	}
	if cfg.Database.SSLMode == "" {
		cfg.Database.SSLMode = "disable"
	}
	if cfg.Database.MaxOpen == 0 {
		cfg.Database.MaxOpen = 100
	}
	if cfg.Database.MaxIdle == 0 {
		cfg.Database.MaxIdle = 10
	}

	// Redis defaults
	if cfg.Redis.Port == 0 {
		cfg.Redis.Port = 6379
	}

	// JWT defaults
	if cfg.JWT.AccessTokenExp == 0 {
		cfg.JWT.AccessTokenExp = 7200
	}
	if cfg.JWT.RefreshTokenExp == 0 {
		cfg.JWT.RefreshTokenExp = 604800
	}
	if cfg.JWT.Issuer == "" {
		cfg.JWT.Issuer = "starbyte"
	}

	// Logger defaults
	if cfg.Logger.Level == "" {
		cfg.Logger.Level = "info"
	}
	if cfg.Logger.Format == "" {
		cfg.Logger.Format = "json"
	}
	if cfg.Logger.Filename == "" {
		cfg.Logger.Filename = "logs/starbyte.log"
	}
	if cfg.Logger.MaxSize == 0 {
		cfg.Logger.MaxSize = 100
	}
	if cfg.Logger.MaxBackups == 0 {
		cfg.Logger.MaxBackups = 5
	}
	if cfg.Logger.MaxAge == 0 {
		cfg.Logger.MaxAge = 30
	}

	// MinIO defaults
	if cfg.MinIO.Endpoint == "" {
		cfg.MinIO.Endpoint = "localhost:9000"
	}
	if cfg.MinIO.Bucket == "" {
		cfg.MinIO.Bucket = "starbyte"
	}

	// Email defaults
	if cfg.Email.SMTPPort == 0 {
		cfg.Email.SMTPPort = 587
	}

	// CORS defaults
	if len(cfg.CORS.AllowedOrigins) == 0 {
		cfg.CORS.AllowedOrigins = []string{"*"}
	}
	if len(cfg.CORS.AllowedMethods) == 0 {
		cfg.CORS.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
	}
	if len(cfg.CORS.AllowedHeaders) == 0 {
		cfg.CORS.AllowedHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-Id"}
	}
}

// validate checks that critical configuration fields are set to meaningful
// values. It returns an error describing the first problem found, or nil
// if the configuration passes all checks.
//
// The env parameter is the current APP_ENV value (may be empty). In
// production mode the validation rules are stricter — for example, the
// JWT secret must not be the placeholder value.
func validate(cfg *Config, env string) error {
	// JWT secret is always required.
	if cfg.JWT.Secret == "" {
		return fmt.Errorf("config validation: jwt.secret is required")
	}
	// In production, reject the placeholder JWT secret.
	if env == "prod" && (cfg.JWT.Secret == "starbyte-secret-key-change-in-production" ||
		strings.Contains(cfg.JWT.Secret, "change-in-production")) {
		return fmt.Errorf("config validation: jwt.secret must be changed from default in production")
	}

	// Database host is required.
	if cfg.Database.Host == "" {
		return fmt.Errorf("config validation: database.host is required")
	}
	if cfg.Database.User == "" {
		return fmt.Errorf("config validation: database.user is required")
	}
	if cfg.Database.DBName == "" {
		return fmt.Errorf("config validation: database.dbname is required")
	}

	// In production, database password must not be empty.
	if env == "prod" && cfg.Database.Password == "" {
		return fmt.Errorf("config validation: database.password is required in production")
	}

	// In production, Redis host must be set.
	if env == "prod" && cfg.Redis.Host == "" {
		return fmt.Errorf("config validation: redis.host is required in production")
	}
	// In production, Redis password should be set (security best practice).
	if env == "prod" && cfg.Redis.Password == "" {
		return fmt.Errorf("config validation: redis.password should be set in production (use environment variable REDIS_PASSWORD)")
	}

	// Server port range check.
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("config validation: server.port must be between 1 and 65535, got %d", cfg.Server.Port)
	}

	// Server mode must be one of the allowed values.
	switch cfg.Server.Mode {
	case "debug", "release", "test":
	default:
		return fmt.Errorf("config validation: server.mode must be debug, release, or test, got %q", cfg.Server.Mode)
	}

	return nil
}
