package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

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
//	CORS_ALLOW_CREDENTIALS — cors.allow_credentials (true/false)
func applyEnvOverrides(cfg *Config) {
	// Server
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = n
		}
	}
	if v := os.Getenv("SERVER_MODE"); v != "" {
		cfg.Server.Mode = v
	}

	// Database
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Database.Port = n
		}
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.Database.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.Database.DBName = v
	}
	if v := os.Getenv("DB_SSLMODE"); v != "" {
		cfg.Database.SSLMode = v
	}
	if v := os.Getenv("DB_MAX_OPEN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Database.MaxOpen = n
		}
	}
	if v := os.Getenv("DB_MAX_IDLE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Database.MaxIdle = n
		}
	}

	// Redis
	if v := os.Getenv("REDIS_HOST"); v != "" {
		cfg.Redis.Host = v
	}
	if v := os.Getenv("REDIS_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Redis.Port = n
		}
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		cfg.Redis.Password = v
	}
	if v := os.Getenv("REDIS_DB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Redis.DB = n
		}
	}

	// JWT
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.JWT.Secret = v
	}
	if v := os.Getenv("JWT_ACCESS_TOKEN_EXP"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.JWT.AccessTokenExp = n
		}
	}
	if v := os.Getenv("JWT_REFRESH_TOKEN_EXP"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.JWT.RefreshTokenExp = n
		}
	}

	// MinIO
	if v := os.Getenv("MINIO_ENDPOINT"); v != "" {
		cfg.MinIO.Endpoint = v
	}
	if v := os.Getenv("MINIO_ACCESS_KEY"); v != "" {
		cfg.MinIO.AccessKey = v
	}
	if v := os.Getenv("MINIO_SECRET_KEY"); v != "" {
		cfg.MinIO.SecretKey = v
	}
	if v := os.Getenv("MINIO_BUCKET"); v != "" {
		cfg.MinIO.Bucket = v
	}
	if v := os.Getenv("MINIO_USE_SSL"); v != "" {
		cfg.MinIO.UseSSL = v == "true" || v == "1"
	}

	// Email
	if v := os.Getenv("SMTP_HOST"); v != "" {
		cfg.Email.SMTPHost = v
	}
	if v := os.Getenv("SMTP_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Email.SMTPPort = n
		}
	}
	if v := os.Getenv("SMTP_USER"); v != "" {
		cfg.Email.Username = v
	}
	if v := os.Getenv("SMTP_PASSWORD"); v != "" {
		cfg.Email.Password = v
	}
	if v := os.Getenv("SMTP_FROM"); v != "" {
		cfg.Email.From = v
	}

	// Logger
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.Logger.Level = v
	}
	if v := os.Getenv("LOG_FORMAT"); v != "" {
		cfg.Logger.Format = v
	}

	// CORS (comma-separated values)
	if v := os.Getenv("CORS_ALLOWED_ORIGINS"); v != "" {
		cfg.CORS.AllowedOrigins = splitCSV(v)
	}
	if v := os.Getenv("CORS_ALLOWED_METHODS"); v != "" {
		cfg.CORS.AllowedMethods = splitCSV(v)
	}
	if v := os.Getenv("CORS_ALLOWED_HEADERS"); v != "" {
		cfg.CORS.AllowedHeaders = splitCSV(v)
	}
	if v := os.Getenv("CORS_ALLOW_CREDENTIALS"); v != "" {
		cfg.CORS.AllowCredentials = v == "true" || v == "1"
	}
}

// getEnv returns the value of an environment variable or a fallback.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getEnvInt returns the integer value of an environment variable or a fallback.
func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
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
