package config

import (
	"fmt"
	"os"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Debug   bool
	Server  ServerConfig
	WeChat  WeChatConfig
	Redis   RedisConfig
	Database DatabaseConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port int
	Env  string
}

// WeChatConfig holds WeChat API configuration
type WeChatConfig struct {
	AppID           string
	AppSecret       string
	Token           string
	EncodingAESKey  string
	UseStableAK     bool
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	Username        string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int // in seconds
}

// Load loads configuration from environment and config file
func Load() (*Config, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("debug", false)
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.env", "development")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("database.sslmode", "disable")

	// Environment variables
	v.SetEnvPrefix("WECHAT")
	v.AutomaticEnv()

	// Type coercion for environment variables
	v.Set("server.port", cast.ToInt(os.Getenv("SERVER_PORT")))
	v.Set("redis.addr", os.Getenv("REDIS_ADDR"))
	v.Set("redis.password", os.Getenv("REDIS_PASSWORD"))
	v.Set("database.host", os.Getenv("DB_HOST"))
	v.Set("database.port", cast.ToInt(os.Getenv("DB_PORT")))
	v.Set("database.username", os.Getenv("DB_USER"))
	v.Set("database.password", os.Getenv("DB_PASSWORD"))
	v.Set("database.name", os.Getenv("DB_NAME"))
	v.Set("wechat.appid", os.Getenv("WECHAT_APPID"))
	v.Set("wechat.appsecret", os.Getenv("WECHAT_APPSECRET"))
	v.Set("wechat.token", os.Getenv("WECHAT_TOKEN"))
	v.Set("wechat.encodingaeskey", os.Getenv("WECHAT_ENCODING_AES_KEY"))

	// Config file (optional)
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate required fields
	if cfg.WeChat.AppID == "" {
		return nil, fmt.Errorf("WECHAT_APPID is required")
	}
	if cfg.WeChat.AppSecret == "" {
		return nil, fmt.Errorf("WECHAT_APPSECRET is required")
	}
	if cfg.WeChat.Token == "" {
		return nil, fmt.Errorf("WECHAT_TOKEN is required")
	}

	return &cfg, nil
}

// GetDSN returns PostgreSQL connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.Username,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}
