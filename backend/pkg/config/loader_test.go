package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// helperWriteConfig writes a YAML config file to a temp dir and returns its path.
func helperWriteConfig(t *testing.T, dir, filename, content string) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", filename, err)
	}
	return path
}

const baseConfigYAML = `server:
  port: 8080
  mode: debug
  read_timeout: 30
  write_timeout: 30
database:
  host: localhost
  port: 5432
  user: starbyte
  password: starbyte
  dbname: starbyte
  sslmode: disable
  max_open: 100
  max_idle: 10
redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
jwt:
  secret: "test-secret-key"
  access_token_exp: 7200
  refresh_token_exp: 604800
  issuer: starbyte
logger:
  level: info
  format: json
  filename: logs/starbyte.log
  max_size: 100
  max_backups: 5
  max_age: 30
  compress: true
minio:
  endpoint: localhost:9000
  access_key: minioadmin
  secret_key: minioadmin
  bucket: starbyte
  use_ssl: false
email:
  smtp_host: smtp.example.com
  smtp_port: 587
  username: noreply@example.com
  password: emailpass
  from: noreply@example.com
cors:
  allowed_origins:
    - "http://localhost:3000"
  allowed_methods:
    - "GET"
    - "POST"
  allowed_headers:
    - "Content-Type"
  allow_credentials: true
`

func TestLoad_BaseConfig(t *testing.T) {
	// 显式声明测试态：基础配置里的弱 JWT 密钥只在 dev/test 下放行，
	// 避免这次 fail-closed 改动把既有测试当成「未设置 APP_ENV」的生产校验而误拒。
	t.Setenv("APP_ENV", "test")
	dir := t.TempDir()
	path := helperWriteConfig(t, dir, "config.yaml", baseConfigYAML)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify server config
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want 8080", cfg.Server.Port)
	}
	if cfg.Server.Mode != "debug" {
		t.Errorf("Server.Mode = %s, want debug", cfg.Server.Mode)
	}

	// Verify database config
	if cfg.Database.Host != "localhost" {
		t.Errorf("Database.Host = %s, want localhost", cfg.Database.Host)
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("Database.Port = %d, want 5432", cfg.Database.Port)
	}

	// Verify JWT
	if cfg.JWT.Secret != "test-secret-key" {
		t.Errorf("JWT.Secret = %s, want test-secret-key", cfg.JWT.Secret)
	}

	// Verify CORS
	if len(cfg.CORS.AllowedOrigins) != 1 || cfg.CORS.AllowedOrigins[0] != "http://localhost:3000" {
		t.Errorf("CORS.AllowedOrigins = %v, want [http://localhost:3000]", cfg.CORS.AllowedOrigins)
	}
	if !cfg.CORS.AllowCredentials {
		t.Error("CORS.AllowCredentials = false, want true")
	}

	// Verify Email
	if cfg.Email.SMTPHost != "smtp.example.com" {
		t.Errorf("Email.SMTPHost = %s, want smtp.example.com", cfg.Email.SMTPHost)
	}
	if cfg.Backup.Prefix != "backups" {
		t.Errorf("Backup.Prefix = %s, want backups", cfg.Backup.Prefix)
	}
	if cfg.Backup.TimeoutSec != 1800 {
		t.Errorf("Backup.TimeoutSec = %d, want 1800", cfg.Backup.TimeoutSec)
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	dir := t.TempDir()
	helperWriteConfig(t, dir, "config.yaml", baseConfigYAML)

	// Create a dev override that changes some values.
	devYAML := `server:
  mode: release
database:
  dbname: starbyte_dev
  max_open: 50
redis:
  db: 1
`
	helperWriteConfig(t, dir, "config.dev.yaml", devYAML)

	t.Setenv("APP_ENV", "dev")
	path := filepath.Join(dir, "config.yaml")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Overridden values
	if cfg.Server.Mode != "release" {
		t.Errorf("Server.Mode = %s, want release (from dev override)", cfg.Server.Mode)
	}
	if cfg.Database.DBName != "starbyte_dev" {
		t.Errorf("Database.DBName = %s, want starbyte_dev", cfg.Database.DBName)
	}
	if cfg.Database.MaxOpen != 50 {
		t.Errorf("Database.MaxOpen = %d, want 50", cfg.Database.MaxOpen)
	}
	if cfg.Redis.DB != 1 {
		t.Errorf("Redis.DB = %d, want 1", cfg.Redis.DB)
	}

	// Non-overridden values should remain from base
	if cfg.Database.Host != "localhost" {
		t.Errorf("Database.Host = %s, want localhost (from base)", cfg.Database.Host)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := helperWriteConfig(t, dir, "config.yaml", "server:\n  port: [invalid")

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	cfg := &Config{
		Server:   ServerConfig{Port: 8080, Mode: "debug"},
		Database: DatabaseConfig{Host: "localhost", Port: 5432, User: "starbyte", DBName: "starbyte"},
		Redis:    RedisConfig{Host: "localhost", Port: 6379},
		JWT:      JWTConfig{Secret: "base-secret"},
		MinIO:    MinIOConfig{Endpoint: "localhost:9000"},
		Email:    EmailConfig{SMTPHost: "smtp.example.com", SMTPPort: 587},
	}

	t.Setenv("SERVER_PORT", "9090")
	t.Setenv("SERVER_MODE", "release")
	t.Setenv("DB_HOST", "db.prod.internal")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_PASSWORD", "secret-pass")
	t.Setenv("REDIS_HOST", "redis.prod.internal")
	t.Setenv("REDIS_PORT", "6380")
	t.Setenv("JWT_SECRET", "env-override-secret")
	t.Setenv("MINIO_ACCESS_KEY", "prod-access-key")
	t.Setenv("SMTP_HOST", "smtp.prod.com")
	t.Setenv("STARBYTE_SMTP_PASSWORD", "from-starbyte-secret")
	t.Setenv("SMTP_PASSWORD", "legacy-should-lose")
	t.Setenv("STARBYTE_SMTP_FROM_NAME", "StarByte-Prod")
	t.Setenv("SMTP_SSL_MODE", "implicit")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://a.com,https://b.com")
	t.Setenv("CORS_ALLOWED_METHODS", "GET,POST,PUT")
	t.Setenv("CORS_ALLOWED_HEADERS", "Content-Type,Authorization")
	t.Setenv("CORS_ALLOW_CREDENTIALS", "true")
	t.Setenv("CORS_EXPOSE_HEADERS", "X-Request-Id,X-Custom-Header")

	applyEnvOverrides(cfg)

	// Verify overrides
	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port = %d, want 9090", cfg.Server.Port)
	}
	if cfg.Server.Mode != "release" {
		t.Errorf("Server.Mode = %s, want release", cfg.Server.Mode)
	}
	if cfg.Database.Host != "db.prod.internal" {
		t.Errorf("Database.Host = %s, want db.prod.internal", cfg.Database.Host)
	}
	if cfg.Database.Port != 5433 {
		t.Errorf("Database.Port = %d, want 5433", cfg.Database.Port)
	}
	if cfg.Database.Password != "secret-pass" {
		t.Errorf("Database.Password = %s, want secret-pass", cfg.Database.Password)
	}
	if cfg.Redis.Host != "redis.prod.internal" {
		t.Errorf("Redis.Host = %s, want redis.prod.internal", cfg.Redis.Host)
	}
	if cfg.Redis.Port != 6380 {
		t.Errorf("Redis.Port = %d, want 6380", cfg.Redis.Port)
	}
	if cfg.JWT.Secret != "env-override-secret" {
		t.Errorf("JWT.Secret = %s, want env-override-secret", cfg.JWT.Secret)
	}
	if cfg.MinIO.AccessKey != "prod-access-key" {
		t.Errorf("MinIO.AccessKey = %s, want prod-access-key", cfg.MinIO.AccessKey)
	}
	if cfg.Email.SMTPHost != "smtp.prod.com" {
		t.Errorf("Email.SMTPHost = %s, want smtp.prod.com", cfg.Email.SMTPHost)
	}
	if cfg.Email.Password != "from-starbyte-secret" {
		t.Errorf("Email.Password = %s, want from-starbyte-secret", cfg.Email.Password)
	}
	if cfg.Email.FromName != "StarByte-Prod" {
		t.Errorf("Email.FromName = %s, want StarByte-Prod", cfg.Email.FromName)
	}
	if cfg.Email.SSLMode != "implicit" {
		t.Errorf("Email.SSLMode = %s, want implicit", cfg.Email.SSLMode)
	}
	if cfg.Logger.Level != "debug" {
		t.Errorf("Logger.Level = %s, want debug", cfg.Logger.Level)
	}
	if len(cfg.CORS.AllowedOrigins) != 2 {
		t.Errorf("CORS.AllowedOrigins len = %d, want 2", len(cfg.CORS.AllowedOrigins))
	}
	if len(cfg.CORS.AllowedMethods) != 3 {
		t.Errorf("CORS.AllowedMethods len = %d, want 3", len(cfg.CORS.AllowedMethods))
	}
	if cfg.CORS.AllowedMethods[0] != "GET" {
		t.Errorf("CORS.AllowedMethods[0] = %s, want GET", cfg.CORS.AllowedMethods[0])
	}
	if len(cfg.CORS.AllowedHeaders) != 2 {
		t.Errorf("CORS.AllowedHeaders len = %d, want 2", len(cfg.CORS.AllowedHeaders))
	}
	if cfg.CORS.AllowedHeaders[1] != "Authorization" {
		t.Errorf("CORS.AllowedHeaders[1] = %s, want Authorization", cfg.CORS.AllowedHeaders[1])
	}
	if len(cfg.CORS.ExposeHeaders) != 2 {
		t.Errorf("CORS.ExposeHeaders len = %d, want 2", len(cfg.CORS.ExposeHeaders))
	}
	if !cfg.CORS.AllowCredentials {
		t.Error("CORS.AllowCredentials = false, want true")
	}

	t.Setenv("STARBYTE_BACKUP_PREFIX", "ops-backups")
	t.Setenv("STARBYTE_BACKUP_TIMEOUT_SEC", "900")
	t.Setenv("STARBYTE_BACKUP_ENCRYPTION_KEY", "env-only-backup-key")
	applyEnvOverrides(cfg)
	if cfg.Backup.Prefix != "ops-backups" {
		t.Errorf("Backup.Prefix = %s, want ops-backups", cfg.Backup.Prefix)
	}
	if cfg.Backup.TimeoutSec != 900 {
		t.Errorf("Backup.TimeoutSec = %d, want 900", cfg.Backup.TimeoutSec)
	}
	if cfg.Backup.EncryptionKey != "env-only-backup-key" {
		t.Errorf("Backup.EncryptionKey not loaded from env")
	}
}

func TestSetDefaults(t *testing.T) {
	cfg := &Config{}
	setDefaults(cfg)

	// Server
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want 8080", cfg.Server.Port)
	}
	if cfg.Server.Mode != "debug" {
		t.Errorf("Server.Mode = %s, want debug", cfg.Server.Mode)
	}
	if cfg.Server.ReadTimeout != 30 {
		t.Errorf("Server.ReadTimeout = %d, want 30", cfg.Server.ReadTimeout)
	}

	// Database
	if cfg.Database.Port != 5432 {
		t.Errorf("Database.Port = %d, want 5432", cfg.Database.Port)
	}
	if cfg.Database.SSLMode != "disable" {
		t.Errorf("Database.SSLMode = %s, want disable", cfg.Database.SSLMode)
	}
	if cfg.Database.MaxOpen != 100 {
		t.Errorf("Database.MaxOpen = %d, want 100", cfg.Database.MaxOpen)
	}

	// Redis
	if cfg.Redis.Port != 6379 {
		t.Errorf("Redis.Port = %d, want 6379", cfg.Redis.Port)
	}

	// JWT
	if cfg.JWT.AccessTokenExp != 7200 {
		t.Errorf("JWT.AccessTokenExp = %d, want 7200", cfg.JWT.AccessTokenExp)
	}
	if cfg.JWT.Issuer != "starbyte" {
		t.Errorf("JWT.Issuer = %s, want starbyte", cfg.JWT.Issuer)
	}

	// Logger
	if cfg.Logger.Level != "info" {
		t.Errorf("Logger.Level = %s, want info", cfg.Logger.Level)
	}
	if cfg.Logger.Format != "json" {
		t.Errorf("Logger.Format = %s, want json", cfg.Logger.Format)
	}

	// MinIO
	if cfg.MinIO.Endpoint != "localhost:9000" {
		t.Errorf("MinIO.Endpoint = %s, want localhost:9000", cfg.MinIO.Endpoint)
	}

	// Email
	if cfg.Email.SMTPPort != 465 {
		t.Errorf("Email.SMTPPort = %d, want 465", cfg.Email.SMTPPort)
	}
	if cfg.Email.SMTPHost != "smtp.exmail.qq.com" {
		t.Errorf("Email.SMTPHost = %s, want smtp.exmail.qq.com", cfg.Email.SMTPHost)
	}
	if cfg.Email.From != "computerassociation@smbu.edu.cn" {
		t.Errorf("Email.From = %s, want computerassociation@smbu.edu.cn", cfg.Email.From)
	}
	if cfg.Email.FromName != "StarByte-SMTP" {
		t.Errorf("Email.FromName = %s, want StarByte-SMTP", cfg.Email.FromName)
	}
	if cfg.Email.SSLMode != "implicit" {
		t.Errorf("Email.SSLMode = %s, want implicit", cfg.Email.SSLMode)
	}

	// CORS
	if len(cfg.CORS.AllowedOrigins) != 1 || cfg.CORS.AllowedOrigins[0] != "*" {
		t.Errorf("CORS.AllowedOrigins = %v, want [*]", cfg.CORS.AllowedOrigins)
	}
	if len(cfg.CORS.AllowedMethods) != 6 {
		t.Errorf("CORS.AllowedMethods len = %d, want 6", len(cfg.CORS.AllowedMethods))
	}

	if cfg.Backup.Prefix != "backups" {
		t.Errorf("Backup.Prefix = %s, want backups", cfg.Backup.Prefix)
	}
	if cfg.Backup.PgDumpBin != "pg_dump" {
		t.Errorf("Backup.PgDumpBin = %s, want pg_dump", cfg.Backup.PgDumpBin)
	}
	if cfg.Backup.TimeoutSec != 1800 {
		t.Errorf("Backup.TimeoutSec = %d, want 1800", cfg.Backup.TimeoutSec)
	}
	if cfg.CAS.DefaultRole != "user" {
		t.Errorf("CAS.DefaultRole = %s, want user", cfg.CAS.DefaultRole)
	}
}

func TestSetDefaults_DoesNotOverrideExisting(t *testing.T) {
	cfg := &Config{
		Server:   ServerConfig{Port: 3000, Mode: "release"},
		Database: DatabaseConfig{Port: 5433, MaxOpen: 200},
		Redis:    RedisConfig{Port: 6380},
	}
	setDefaults(cfg)

	if cfg.Server.Port != 3000 {
		t.Errorf("Server.Port = %d, want 3000 (existing should not be overridden)", cfg.Server.Port)
	}
	if cfg.Server.Mode != "release" {
		t.Errorf("Server.Mode = %s, want release", cfg.Server.Mode)
	}
	if cfg.Database.Port != 5433 {
		t.Errorf("Database.Port = %d, want 5433", cfg.Database.Port)
	}
	if cfg.Database.MaxOpen != 200 {
		t.Errorf("Database.MaxOpen = %d, want 200", cfg.Database.MaxOpen)
	}
	if cfg.Redis.Port != 6380 {
		t.Errorf("Redis.Port = %d, want 6380", cfg.Redis.Port)
	}
}

func TestValidate_Success(t *testing.T) {
	cfg := &Config{
		Server:   ServerConfig{Port: 8080, Mode: "debug"},
		Database: DatabaseConfig{Host: "localhost", User: "starbyte", DBName: "starbyte", Password: "pass"},
		JWT:      JWTConfig{Secret: "valid-secret"},
	}
	if err := validate(cfg, "test"); err != nil {
		t.Errorf("validate() error = %v, want nil", err)
	}
}

func TestValidate_MissingJWTSecret(t *testing.T) {
	cfg := &Config{
		Server:   ServerConfig{Port: 8080, Mode: "debug"},
		Database: DatabaseConfig{Host: "localhost", User: "starbyte", DBName: "starbyte"},
		JWT:      JWTConfig{Secret: ""},
	}
	err := validate(cfg, "")
	if err == nil {
		t.Fatal("expected error for empty JWT secret, got nil")
	}
}

func TestValidate_ProductionDefaultSecret(t *testing.T) {
	cfg := &Config{
		Server:   ServerConfig{Port: 8080, Mode: "release"},
		Database: DatabaseConfig{Host: "localhost", User: "starbyte", DBName: "starbyte", Password: "pass"},
		JWT:      JWTConfig{Secret: "starbyte-secret-key-change-in-production"},
	}
	err := validate(cfg, "prod")
	if err == nil {
		t.Fatal("expected error for default JWT secret in prod, got nil")
	}
}

func TestValidate_ProductionEmptyPassword(t *testing.T) {
	cfg := &Config{
		Server:   ServerConfig{Port: 8080, Mode: "release"},
		Database: DatabaseConfig{Host: "localhost", User: "starbyte", DBName: "starbyte", Password: ""},
		JWT:      JWTConfig{Secret: "aT9kL2mN8qR5vB7wX4yZ1cD6eF3gH0jP9sU2n"},
	}
	err := validate(cfg, "prod")
	if err == nil {
		t.Fatal("expected error for empty DB password in prod, got nil")
	}
}

func TestValidate_InvalidPort(t *testing.T) {
	cfg := &Config{
		Server:   ServerConfig{Port: 0, Mode: "debug"},
		Database: DatabaseConfig{Host: "localhost", User: "starbyte", DBName: "starbyte"},
		JWT:      JWTConfig{Secret: "valid-secret"},
	}
	err := validate(cfg, "test")
	if err == nil {
		t.Fatal("expected error for invalid port, got nil")
	}
}

func TestValidate_InvalidMode(t *testing.T) {
	cfg := &Config{
		Server:   ServerConfig{Port: 8080, Mode: "invalid"},
		Database: DatabaseConfig{Host: "localhost", User: "starbyte", DBName: "starbyte"},
		JWT:      JWTConfig{Secret: "valid-secret"},
	}
	err := validate(cfg, "test")
	if err == nil {
		t.Fatal("expected error for invalid mode, got nil")
	}
}

func TestValidate_MissingDBHost(t *testing.T) {
	cfg := &Config{
		Server:   ServerConfig{Port: 8080, Mode: "debug"},
		Database: DatabaseConfig{User: "starbyte", DBName: "starbyte"},
		JWT:      JWTConfig{Secret: "valid-secret"},
	}
	err := validate(cfg, "test")
	if err == nil {
		t.Fatal("expected error for missing DB host, got nil")
	}
}

func TestSplitCSV(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"empty", "", nil},
		{"single", "a", []string{"a"}},
		{"multiple", "a,b,c", []string{"a", "b", "c"}},
		{"with spaces", " a , b , c ", []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitCSV(tt.in)
			if len(got) != len(tt.want) {
				t.Errorf("splitCSV(%q) = %v, want %v", tt.in, got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitCSV(%q)[%d] = %q, want %q", tt.in, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	t.Setenv("TEST_GETENV_VAR", "hello")
	if got := getEnv("TEST_GETENV_VAR", "default"); got != "hello" {
		t.Errorf("getEnv = %s, want hello", got)
	}
	if got := getEnv("TEST_GETENV_UNSET", "default"); got != "default" {
		t.Errorf("getEnv = %s, want default", got)
	}
}

func TestGetEnvInt(t *testing.T) {
	t.Setenv("TEST_GETENV_INT", "42")
	if got := getEnvInt("TEST_GETENV_INT", 0); got != 42 {
		t.Errorf("getEnvInt = %d, want 42", got)
	}
	if got := getEnvInt("TEST_GETENV_INT_UNSET", 99); got != 99 {
		t.Errorf("getEnvInt = %d, want 99", got)
	}
}

func TestGetEnvInt_InvalidValue(t *testing.T) {
	t.Setenv("TEST_GETENV_INT_INVALID", "not-a-number")
	got := getEnvInt("TEST_GETENV_INT_INVALID", 77)
	if got != 77 {
		t.Errorf("getEnvInt with invalid value = %d, want fallback 77", got)
	}
}

func TestGetEnvBool(t *testing.T) {
	t.Setenv("TEST_GETENV_BOOL_TRUE", "true")
	t.Setenv("TEST_GETENV_BOOL_ONE", "1")
	t.Setenv("TEST_GETENV_BOOL_FALSE", "false")

	if got := getEnvBool("TEST_GETENV_BOOL_TRUE", false); !got {
		t.Errorf("getEnvBool('true') = false, want true")
	}
	if got := getEnvBool("TEST_GETENV_BOOL_ONE", false); !got {
		t.Errorf("getEnvBool('1') = false, want true")
	}
	if got := getEnvBool("TEST_GETENV_BOOL_FALSE", true); got {
		t.Errorf("getEnvBool('false') = true, want false")
	}
	if got := getEnvBool("TEST_GETENV_BOOL_UNSET", true); !got {
		t.Errorf("getEnvBool(unset) = false, want fallback true")
	}
}

func TestGetEnvCSV(t *testing.T) {
	t.Setenv("TEST_GETENV_CSV", "a, b, c")
	got := getEnvCSV("TEST_GETENV_CSV", nil)
	if len(got) != 3 {
		t.Fatalf("getEnvCSV len = %d, want 3", len(got))
	}
	if got[1] != "b" {
		t.Errorf("getEnvCSV[1] = %q, want 'b'", got[1])
	}

	// Unset should return fallback.
	fallback := []string{"default"}
	got = getEnvCSV("TEST_GETENV_CSV_UNSET", fallback)
	if len(got) != 1 || got[0] != "default" {
		t.Errorf("getEnvCSV(unset) = %v, want [default]", got)
	}
}

func TestLoad_InvalidAppEnv(t *testing.T) {
	dir := t.TempDir()
	path := helperWriteConfig(t, dir, "config.yaml", baseConfigYAML)

	t.Setenv("APP_ENV", "production")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid APP_ENV, got nil")
	}
}

func TestValidate_ProductionMissingRedisHost(t *testing.T) {
	cfg := &Config{
		Server:   ServerConfig{Port: 8080, Mode: "release"},
		Database: DatabaseConfig{Host: "db.prod", User: "starbyte", DBName: "starbyte", Password: "pass"},
		Redis:    RedisConfig{Host: "", Port: 6379, Password: "redis-pass"},
		JWT:      JWTConfig{Secret: "aT9kL2mN8qR5vB7wX4yZ1cD6eF3gH0jP9sU2n"},
	}
	err := validate(cfg, "prod")
	if err == nil {
		t.Fatal("expected error for missing Redis host in prod, got nil")
	}
}

func TestValidate_ProductionEmptyRedisPassword(t *testing.T) {
	cfg := &Config{
		Server:   ServerConfig{Port: 8080, Mode: "release"},
		Database: DatabaseConfig{Host: "db.prod", User: "starbyte", DBName: "starbyte", Password: "pass"},
		Redis:    RedisConfig{Host: "redis.prod", Port: 6379, Password: ""},
		JWT:      JWTConfig{Secret: "aT9kL2mN8qR5vB7wX4yZ1cD6eF3gH0jP9sU2n"},
	}
	err := validate(cfg, "prod")
	if err == nil {
		t.Fatal("expected error for empty Redis password in prod, got nil")
	}
}

func TestValidate_ProductionRedisValid(t *testing.T) {
	cfg := &Config{
		Server:   ServerConfig{Port: 8080, Mode: "release"},
		Database: DatabaseConfig{Host: "db.prod", User: "starbyte", DBName: "starbyte", Password: "pass"},
		Redis:    RedisConfig{Host: "redis.prod", Port: 6379, Password: "redis-pass"},
		JWT:      JWTConfig{Secret: "aT9kL2mN8qR5vB7wX4yZ1cD6eF3gH0jP9sU2n"},
	}
	if err := validate(cfg, "prod"); err != nil {
		t.Errorf("validate() in prod with valid Redis = %v, want nil", err)
	}
}

// 以下为 JWT 强度校验的专项测试（S-01 后端侧）。

func TestValidate_ProdJWTTooShort(t *testing.T) {
	cfg := &Config{
		Server:   ServerConfig{Port: 8080, Mode: "release"},
		Database: DatabaseConfig{Host: "db.prod", User: "starbyte", DBName: "starbyte", Password: "pass"},
		Redis:    RedisConfig{Host: "redis.prod", Port: 6379, Password: "redis-pass"},
		JWT:      JWTConfig{Secret: "short-but-strong-looking"}, // 24 chars < 32
	}
	if err := validate(cfg, "prod"); err == nil {
		t.Fatal("expected error for <32 char JWT secret in prod, got nil")
	}
}

func TestValidate_ProdJWTPlaceholder(t *testing.T) {
	cfg := &Config{
		Server:   ServerConfig{Port: 8080, Mode: "release"},
		Database: DatabaseConfig{Host: "db.prod", User: "starbyte", DBName: "starbyte", Password: "pass"},
		Redis:    RedisConfig{Host: "redis.prod", Port: 6379, Password: "redis-pass"},
		// compose 兜底值 change-me-in-production 不含旧子串，但命中新占位词表。
		JWT: JWTConfig{Secret: "change-me-in-production-abcdefghijklmnop"},
	}
	if err := validate(cfg, "prod"); err == nil {
		t.Fatal("expected error for placeholder JWT secret in prod, got nil")
	}
}

func TestValidate_ProdJWTLongRandomOK(t *testing.T) {
	cfg := &Config{
		Server:   ServerConfig{Port: 8080, Mode: "release"},
		Database: DatabaseConfig{Host: "db.prod", User: "starbyte", DBName: "starbyte", Password: "pass"},
		Redis:    RedisConfig{Host: "redis.prod", Port: 6379, Password: "redis-pass"},
		// 64 字符高熵随机串，即便偶含占位词也不应被拒。
		JWT: JWTConfig{Secret: "xK9mP2qR5vB7wT1zC4dF6gH8jL0sU3yE5aN7cX9mP2qR5vB7wT1zC4dF6gH8jL0sU3yE5aN7cX9mP2qR5vB7"},
	}
	if err := validate(cfg, "prod"); err != nil {
		t.Errorf("validate() with 64-char random secret = %v, want nil", err)
	}
}

func TestValidate_ProdJWTLowEntropy(t *testing.T) {
	cfg := &Config{
		Server:   ServerConfig{Port: 8080, Mode: "release"},
		Database: DatabaseConfig{Host: "db.prod", User: "starbyte", DBName: "starbyte", Password: "pass"},
		Redis:    RedisConfig{Host: "redis.prod", Port: 6379, Password: "redis-pass"},
		JWT:      JWTConfig{Secret: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, // 40 个 a
	}
	if err := validate(cfg, "prod"); err == nil {
		t.Fatal("expected error for low-entropy (repeated) JWT secret in prod, got nil")
	}
}

// TestValidate_ProdJWTSecretBypass 覆盖上一轮修复的两个已确认绕过：
// 长重复串过长度+熵检查，但旧逻辑或因词表缺失、或因 len<64 跳过词表而放行。
func TestValidate_ProdJWTSecretBypass(t *testing.T) {
	bypasses := []string{
		strings.Repeat("password", 5),   // 40 chars, distinct=7, "password" 不在旧词表
		strings.Repeat("starbyte", 8),   // 64 chars, 因 len<64 不成立而跳过词表
		strings.Repeat("change-me-", 7), // 70 chars, 同理跳过词表
	}
	for _, s := range bypasses {
		if err := validateProdJWTSecret(s); err == nil {
			t.Errorf("validateProdJWTSecret(%q) = nil, want rejection (confirmed bypass value)", s)
		}
	}
}

// TestValidate_ProdJWTSecretStrongOK 确保真正的 64 位 [A-Za-z0-9] 随机串不被误伤。
func TestValidate_ProdJWTSecretStrongOK(t *testing.T) {
	// 与 CLI 生成的真实 JWT_SECRET 同形态（64 个 [A-Za-z0-9]），且不含任何占位词子串。
	strong := "AbCdEfGhIjKlMnOpQrStUvWxYz9073528416AbCdEfGhIjKlMnOpQrStUvWxYz90"
	if len(strong) != 64 {
		t.Fatalf("test fixture must be 64 chars, got %d", len(strong))
	}
	if err := validateProdJWTSecret(strong); err != nil {
		t.Errorf("validateProdJWTSecret(strong 64-char) = %v, want nil", err)
	}
}

func TestValidate_NonProdWeakSecretOK(t *testing.T) {
	// 非生产环境不强制 JWT 强度，弱占位词也不应报错。
	cfg := &Config{
		Server:   ServerConfig{Port: 8080, Mode: "debug"},
		Database: DatabaseConfig{Host: "localhost", User: "starbyte", DBName: "starbyte"},
		JWT:      JWTConfig{Secret: "change-me-in-production"},
	}
	if err := validate(cfg, "dev"); err != nil {
		t.Errorf("validate() in non-prod with weak secret = %v, want nil", err)
	}
}

func TestLoad_MinimalConfigWithDefaults(t *testing.T) {
	// 显式声明测试态：minYAML 用的是弱 JWT 密钥，只在 dev/test 下放行。
	t.Setenv("APP_ENV", "test")
	// A minimal config with only critical fields.
	minYAML := `database:
  host: db.local
  user: admin
  dbname: mydb
jwt:
  secret: minimal-secret
`
	dir := t.TempDir()
	path := helperWriteConfig(t, dir, "config.yaml", minYAML)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Check that defaults were applied
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want default 8080", cfg.Server.Port)
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("Database.Port = %d, want default 5432", cfg.Database.Port)
	}
	if cfg.Redis.Port != 6379 {
		t.Errorf("Redis.Port = %d, want default 6379", cfg.Redis.Port)
	}

	// Check that provided values are kept
	if cfg.Database.Host != "db.local" {
		t.Errorf("Database.Host = %s, want db.local", cfg.Database.Host)
	}
	if cfg.JWT.Secret != "minimal-secret" {
		t.Errorf("JWT.Secret = %s, want minimal-secret", cfg.JWT.Secret)
	}
}

// ── fail-closed：未设置 APP_ENV 时按生产严格校验（默认拒绝） ─────────────

// TestLoad_DefaultEnv_PlaceholderJWTRejected 证明：未声明 APP_ENV 时，
// 占位 JWT 密钥（即 configs/config.yaml 里的公开值）会被拒绝，后端将启动
// 失败——这正是我们要的 fail-closed 结果（不能离线伪造 super_admin token）。
func TestLoad_DefaultEnv_PlaceholderJWTRejected(t *testing.T) {
	// 确保 APP_ENV 未设置（空值等价于未设置：Loader 把空值补成 prod）。
	t.Setenv("APP_ENV", "")

	yaml := `server:
  port: 8080
  mode: debug
database:
  host: localhost
  user: starbyte
  dbname: starbyte
  password: strong-db-pass
redis:
  host: localhost
  password: strong-redis-pass
jwt:
  secret: "starbyte-secret-key-change-in-production"
`
	dir := t.TempDir()
	path := helperWriteConfig(t, dir, "config.yaml", yaml)

	if _, err := Load(path); err == nil {
		t.Fatal("expected rejection of placeholder JWT secret when APP_ENV unset, got nil")
	}
}

// TestLoad_TestEnv_PlaceholderJWTAllowed 证明：显式 APP_ENV=test 时，
// 同一弱占位密钥被放行（测试态允许弱口令，不被当成生产误拒）。
func TestLoad_TestEnv_PlaceholderJWTAllowed(t *testing.T) {
	t.Setenv("APP_ENV", "test")

	yaml := `server:
  port: 8080
  mode: debug
database:
  host: localhost
  user: starbyte
  dbname: starbyte
jwt:
  secret: "starbyte-secret-key-change-in-production"
`
	dir := t.TempDir()
	path := helperWriteConfig(t, dir, "config.yaml", yaml)

	if _, err := Load(path); err != nil {
		t.Fatalf("expected placeholder JWT allowed in test env, got %v", err)
	}
}

// TestLoad_DefaultEnv_StrongJWTAllowed 证明：未设置 APP_ENV 时，
// 强随机 64 位 [A-Za-z0-9] 密钥不会被误伤（可以正常启动）。
func TestLoad_DefaultEnv_StrongJWTAllowed(t *testing.T) {
	t.Setenv("APP_ENV", "")

	// 与 CLI 生成的真实 JWT_SECRET 同形态（64 个 [A-Za-z0-9]，不含占位词）。
	strong := "AbCdEfGhIjKlMnOpQrStUvWxYz9073528416AbCdEfGhIjKlMnOpQrStUvWxYz90"
	if len(strong) != 64 {
		t.Fatalf("test fixture must be 64 chars, got %d", len(strong))
	}

	yaml := `server:
  port: 8080
  mode: debug
database:
  host: localhost
  user: starbyte
  dbname: starbyte
  password: strong-db-pass
redis:
  host: localhost
  password: strong-redis-pass
jwt:
  secret: "` + strong + `"
`
	dir := t.TempDir()
	path := helperWriteConfig(t, dir, "config.yaml", yaml)

	if _, err := Load(path); err != nil {
		t.Fatalf("expected strong 64-char JWT allowed when APP_ENV unset, got %v", err)
	}
}
