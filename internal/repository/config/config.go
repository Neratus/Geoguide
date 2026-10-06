package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	CLI      CLIConfig      `mapstructure:"cli"`
	SMTP     SMTPConfig     `mapstructure:"smtp"`
}

type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

type DatabaseConfig struct {
	PrimaryType string          `mapstructure:"primary_type"`
	CacheType   string          `mapstructure:"cache_type"`
	Postgres    PostgresConfig  `mapstructure:"postgres"`
	Cassandra   CassandraConfig `mapstructure:"cassandra"`
	Redis       RedisConfig     `mapstructure:"redis"`
	Tarantool   TarantoolConfig `mapstructure:"tarantool"`
	Minio       MinioConfig     `mapstructure:"minio"`
	Auth        AuthConfig      `mapstructure:"auth"`
}

type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type CassandraConfig struct {
	Hosts       []string `mapstructure:"hosts"`
	Keyspace    string   `mapstructure:"keyspace"`
	Consistency string   `mapstructure:"consistency"`
	Username    string   `mapstructure:"username"`
	Password    string   `mapstructure:"password"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type TarantoolConfig struct {
	Hosts    []string `mapstructure:"hosts"`
	Username string   `mapstructure:"username"`
	Password string   `mapstructure:"password"`
	Space    string   `mapstructure:"space"`
}

type MinioConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl"`
}

type AuthConfig struct {
	JWTSecret string `mapstructure:"jwt_secret"`
}

type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	File       string `mapstructure:"file"`
	MaxSizeMB  int    `mapstructure:"max_size_mb"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAgeDays int    `mapstructure:"max_age_days"`
}

type SMTPConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	From        string `mapstructure:"from"`
	Password    string `mapstructure:"password"`
	TemplateDir string `mapstructure:"template_dir"`
}

type CLIConfig struct {
	Interactive bool `mapstructure:"interactive"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.SetEnvPrefix("GEOGUIDE")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.Database.Postgres.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.Database.Postgres.Port = p
		}
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.Database.Postgres.DBName = v
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.Database.Postgres.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.Database.Postgres.Password = v
	}
	if v := os.Getenv("REDIS_HOST"); v != "" {
		cfg.Database.Redis.Host = v
	}

	if v := os.Getenv("S3_ENDPOINT"); v != "" {
		cfg.Database.Minio.Endpoint = v
	}
	if v := os.Getenv("S3_ACCESS_KEY"); v != "" {
		cfg.Database.Minio.AccessKey = v
	}
	if v := os.Getenv("S3_SECRET_KEY"); v != "" {
		cfg.Database.Minio.SecretKey = v
	}
	if v := os.Getenv("S3_BUCKET"); v != "" {
		cfg.Database.Minio.Bucket = v
	}
	if v := os.Getenv("S3_USE_SSL"); v != "" {
		cfg.Database.Minio.UseSSL = v == "true" || v == "1" || v == "TRUE"
	}

	cfg.Database.Minio.Endpoint = normalizeMinioEndpoint(cfg.Database.Minio.Endpoint)

	return &cfg, nil
}

func normalizeMinioEndpoint(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}

	if strings.HasPrefix(raw, "https://") {
		raw = strings.TrimPrefix(raw, "https://")
	} else if strings.HasPrefix(raw, "http://") {
		raw = strings.TrimPrefix(raw, "http://")
	}

	if i := strings.IndexAny(raw, "/?#"); i != -1 {
		raw = raw[:i]
	}

	return raw
}

func LoadFile(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &cfg, nil
}

func LoadTest() (*Config, error) {
	return Load("../../../config/config.yaml")
}

func (c *PostgresConfig) ConnString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}

func (c *Config) PostgresConnString() string {
	return c.Database.Postgres.ConnString()
}

func (c *CassandraConfig) ContactPoints() []string {
	return c.Hosts
}
