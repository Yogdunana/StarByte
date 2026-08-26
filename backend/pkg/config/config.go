package config

// Config is the root configuration object loaded from YAML files.
//
// Loading order (each layer overrides the previous):
//  1. configs/config.yaml        — base config (shared across all environments)
//  2. configs/config.{APP_ENV}.yaml — environment-specific overrides (dev/test/prod)
//  3. Environment variables       — sensitive values injected at runtime
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	Logger   LoggerConfig   `yaml:"logger"`
	MinIO    MinIOConfig    `yaml:"minio"`
	Email    EmailConfig    `yaml:"email"`
	CORS     CORSConfig     `yaml:"cors"`
}

// ServerConfig holds the HTTP server settings.
type ServerConfig struct {
	Port         int    `yaml:"port"`
	Mode         string `yaml:"mode"` // debug, release, test
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
	Format     string `yaml:"format"` // json, console
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

// EmailConfig holds the SMTP email settings.
type EmailConfig struct {
	SMTPHost string `yaml:"smtp_host"`
	SMTPPort int    `yaml:"smtp_port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
}

// CORSConfig holds the Cross-Origin Resource Sharing settings.
type CORSConfig struct {
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
}
