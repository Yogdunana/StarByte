package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the root configuration object loaded from the YAML file.
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	Logger   LoggerConfig   `yaml:"logger"`
	MinIO    MinIOConfig    `yaml:"minio"`
}

// ServerConfig holds the HTTP server settings.
type ServerConfig struct {
	Port         int    `yaml:"port"`
	Mode         string `yaml:"mode"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
}

// DatabaseConfig holds the PostgreSQL connection settings.
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
	MaxOpen  int    `yaml:"max_open"`
	MaxIdle  int    `yaml:"max_idle"`
}

// RedisConfig holds the Redis connection settings.
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// JWTConfig holds the JSON Web Token settings.
type JWTConfig struct {
	Secret          string `yaml:"secret"`
	AccessTokenExp  int    `yaml:"access_token_exp"`
	RefreshTokenExp int    `yaml:"refresh_token_exp"`
	Issuer          string `yaml:"issuer"`
}

// LoggerConfig holds the logger and log-rotation settings.
type LoggerConfig struct {
	Level      string `yaml:"level"`
	Filename   string `yaml:"filename"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

// MinIOConfig holds the object storage settings.
type MinIOConfig struct {
	Endpoint  string `yaml:"endpoint"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	Bucket    string `yaml:"bucket"`
	UseSSL    bool   `yaml:"use_ssl"`
}

// Load reads the YAML configuration file at the given path, parses it and
// fills in sensible defaults for any missing values.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	setDefaults(cfg)
	return cfg, nil
}

// setDefaults populates zero-valued fields with sensible development defaults.
func setDefaults(cfg *Config) {
	// Server
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

	// Database
	if cfg.Database.Host == "" {
		cfg.Database.Host = "localhost"
	}
	if cfg.Database.Port == 0 {
		cfg.Database.Port = 5432
	}
	if cfg.Database.User == "" {
		cfg.Database.User = "starbyte"
	}
	if cfg.Database.DBName == "" {
		cfg.Database.DBName = "starbyte"
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

	// Redis
	if cfg.Redis.Host == "" {
		cfg.Redis.Host = "localhost"
	}
	if cfg.Redis.Port == 0 {
		cfg.Redis.Port = 6379
	}

	// JWT
	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "starbyte-secret-key-change-in-production"
	}
	if cfg.JWT.AccessTokenExp == 0 {
		cfg.JWT.AccessTokenExp = 7200
	}
	if cfg.JWT.RefreshTokenExp == 0 {
		cfg.JWT.RefreshTokenExp = 604800
	}
	if cfg.JWT.Issuer == "" {
		cfg.JWT.Issuer = "starbyte"
	}

	// Logger
	if cfg.Logger.Level == "" {
		cfg.Logger.Level = "info"
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

	// MinIO
	if cfg.MinIO.Endpoint == "" {
		cfg.MinIO.Endpoint = "localhost:9000"
	}
	if cfg.MinIO.AccessKey == "" {
		cfg.MinIO.AccessKey = "minioadmin"
	}
	if cfg.MinIO.SecretKey == "" {
		cfg.MinIO.SecretKey = "minioadmin"
	}
	if cfg.MinIO.Bucket == "" {
		cfg.MinIO.Bucket = "starbyte"
	}
}
