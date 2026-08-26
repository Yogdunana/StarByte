package config

import (
	"os"
	"path/filepath"
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
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://a.com,https://b.com")
	t.Setenv("CORS_ALLOW_CREDENTIALS", "true")

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
	if cfg.Logger.Level != "debug" {
		t.Errorf("Logger.Level = %s, want debug", cfg.Logger.Level)
	}
	if len(cfg.CORS.AllowedOrigins) != 2 {
		t.Errorf("CORS.AllowedOrigins len = %d, want 2", len(cfg.CORS.AllowedOrigins))
	}
	if !cfg.CORS.AllowCredentials {
		t.Error("CORS.AllowCredentials = false, want true")
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
	if cfg.Email.SMTPPort != 587 {
		t.Errorf("Email.SMTPPort = %d, want 587", cfg.Email.SMTPPort)
	}

	// CORS
	if len(cfg.CORS.AllowedOrigins) != 1 || cfg.CORS.AllowedOrigins[0] != "*" {
		t.Errorf("CORS.AllowedOrigins = %v, want [*]", cfg.CORS.AllowedOrigins)
	}
	if len(cfg.CORS.AllowedMethods) != 6 {
		t.Errorf("CORS.AllowedMethods len = %d, want 6", len(cfg.CORS.AllowedMethods))
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
	if err := validate(cfg, ""); err != nil {
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
		JWT:      JWTConfig{Secret: "real-prod-secret"},
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
	err := validate(cfg, "")
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
	err := validate(cfg, "")
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
	err := validate(cfg, "")
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

func TestLoad_MinimalConfigWithDefaults(t *testing.T) {
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
