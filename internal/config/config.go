package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	// Server Configuration
	Server struct {
		Host           string `yaml:"host"`
		Port           int    `yaml:"port"`
		ReadTimeout    int    `yaml:"read_timeout"`    // seconds
		WriteTimeout   int    `yaml:"write_timeout"`   // seconds
		MaxHeaderBytes int    `yaml:"max_header_bytes"`
		Env            string `yaml:"env"` // development, production
	} `yaml:"server"`

	// WeChat Configuration
	WeChat struct {
		AppID          string `yaml:"app_id"`
		AppSecret      string `yaml:"app_secret"`
		Token          string `yaml:"token"`
		EncodingAESKey string `yaml:"encoding_aes_key"` // optional
		AppKey         string `yaml:"app_key"`          // for stable access token
	} `yaml:"wechat"`

	// AccessToken Configuration
	AccessToken struct {
		CacheDuration     int  `yaml:"cache_duration"` // seconds, default 7000 (100 min before expire)
		RefreshInterval   int  `yaml:"refresh_interval"` // seconds
		EnableProactive   bool `yaml:"enable_proactive"`
		EnableReactive    bool `yaml:"enable_reactive"`
		UseStableAPI      bool `yaml:"use_stable_api"` // use stable access token API
	} `yaml:"access_token"`

	// Rate Limiting Configuration
	RateLimit struct {
		Enabled    bool   `yaml:"enabled"`
		Storage    string `yaml:"storage"` // memory, redis
		RedisAddr  string `yaml:"redis_addr"`
		Prefix     string `yaml:"prefix"`
		APIQuotas  map[string]int `yaml:"api_quotas"` // daily quotas per API
	} `yaml:"rate_limit"`

	// Monitoring Configuration
	Monitoring struct {
		Enabled       bool     `yaml:"enabled"`
		MetricsPath   string   `yaml:"metrics_path"`
		HealthPath    string   `yaml:"health_path"`
		AlertEnabled  bool     `yaml:"alert_enabled"`
		AlertWebhook  string   `yaml:"alert_webhook"`
		LogLevel      string   `yaml:"log_level"` // debug, info, warn, error
	} `yaml:"monitoring"`

	// Async Processing Configuration
	Async struct {
		Enabled       bool `yaml:"enabled"`
		Workers       int  `yaml:"workers"`
		QueueSize     int  `yaml:"queue_size"`
		RetryCount    int  `yaml:"retry_count"`
		RetryDelay    int  `yaml:"retry_delay"` // milliseconds
	} `yaml:"async"`

	// Database Configuration (optional)
	Database struct {
		Type     string `yaml:"type"` // mysql, postgres, sqlite
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Name     string `yaml:"name"`
		MaxOpen  int    `yaml:"max_open"`
		MaxIdle  int    `yaml:"max_idle"`
	} `yaml:"database"`

	// Cache Configuration
	Cache struct {
		Type  string `yaml:"type"` // memory, redis
		Redis struct {
			Addr     string `yaml:"addr"`
			Password string `yaml:"password"`
			DB       int    `yaml:"db"`
		} `yaml:"redis"`
	} `yaml:"cache"`

	// API Domains Configuration
	APIDomain struct {
		Primary       string `yaml:"primary"`         // api.weixin.qq.com
		Backup        string `yaml:"backup"`          // api2.weixin.qq.com
		Shanghai      string `yaml:"shanghai"`        // sh.api.weixin.qq.com
		Shenzhen      string `yaml:"shenzhen"`        // sz.api.weixin.qq.com
		HongKong      string `yaml:"hong_kong"`       // hk.api.weixin.qq.com
		CurrentDomain string `yaml:"current_domain"`  // which domain to use
	} `yaml:"api_domain"`
}

// Load reads configuration from file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults
	cfg.applyDefaults()

	// Apply environment overrides
	cfg.applyEnvOverrides()

	return &cfg, nil
}

// applyDefaults sets default values for missing configuration
func (c *Config) applyDefaults() {
	// Server defaults
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = 10
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = 10
	}
	if c.Server.MaxHeaderBytes == 0 {
		c.Server.MaxHeaderBytes = 1 << 20 // 1MB
	}
	if c.Server.Env == "" {
		c.Server.Env = "development"
	}

	// AccessToken defaults
	if c.AccessToken.CacheDuration == 0 {
		c.AccessToken.CacheDuration = 7000 // 100 min before 2-hour expiry
	}
	if c.AccessToken.RefreshInterval == 0 {
		c.AccessToken.RefreshInterval = 3600 // 1 hour
	}
	c.AccessToken.EnableProactive = true
	c.AccessToken.EnableReactive = true
	c.AccessToken.UseStableAPI = true

	// Monitoring defaults
	if c.Monitoring.MetricsPath == "" {
		c.Monitoring.MetricsPath = "/metrics"
	}
	if c.Monitoring.HealthPath == "" {
		c.Monitoring.HealthPath = "/health"
	}
	if c.Monitoring.LogLevel == "" {
		c.Monitoring.LogLevel = "info"
	}

	// Async defaults
	if c.Async.Workers == 0 {
		c.Async.Workers = 10
	}
	if c.Async.QueueSize == 0 {
		c.Async.QueueSize = 1000
	}
	if c.Async.RetryCount == 0 {
		c.Async.RetryCount = 3
	}
	if c.Async.RetryDelay == 0 {
		c.Async.RetryDelay = 1000 // 1 second
	}

	// API Domain defaults
	if c.APIDomain.CurrentDomain == "" {
		c.APIDomain.CurrentDomain = "primary"
	}
	if c.APIDomain.Primary == "" {
		c.APIDomain.Primary = "api.weixin.qq.com"
	}
	if c.APIDomain.Backup == "" {
		c.APIDomain.Backup = "api2.weixin.qq.com"
	}
	if c.APIDomain.Shanghai == "" {
		c.APIDomain.Shanghai = "sh.api.weixin.qq.com"
	}
	if c.APIDomain.Shenzhen == "" {
		c.APIDomain.Shenzhen = "sz.api.weixin.qq.com"
	}
	if c.APIDomain.HongKong == "" {
		c.APIDomain.HongKong = "hk.api.weixin.qq.com"
	}

	// Rate limit default quotas
	if c.RateLimit.APIQuotas == nil {
		c.RateLimit.APIQuotas = map[string]int{
			"access_token":     2000,
			"menu_create":      1000,
			"menu_query":       10000,
			"menu_delete":      1000,
			"tag_create":       1000,
			"tag_query":        1000,
			"tag_update":       1000,
			"tag_move_user":    100000,
			"media_upload":     100000,
			"media_download":   200000,
			"customer_message": 500000,
			"mass_send":        100,
			"qrcode_create":    100000,
			"user_list":        500,
			"user_info":        5000000,
		}
	}
}

// applyEnvOverrides applies environment variable overrides
func (c *Config) applyEnvOverrides() {
	// Server overrides
	if v := os.Getenv("SERVER_HOST"); v != "" {
		c.Server.Host = v
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &c.Server.Port)
	}
	if v := os.Getenv("SERVER_ENV"); v != "" {
		c.Server.Env = v
	}

	// WeChat overrides
	if v := os.Getenv("WECHAT_APPID"); v != "" {
		c.WeChat.AppID = v
	}
	if v := os.Getenv("WECHAT_APPSECRET"); v != "" {
		c.WeChat.AppSecret = v
	}
	if v := os.Getenv("WECHAT_TOKEN"); v != "" {
		c.WeChat.Token = v
	}
	if v := os.Getenv("WECHAT_ENCODING_AES_KEY"); v != "" {
		c.WeChat.EncodingAESKey = v
	}

	// Redis overrides
	if v := os.Getenv("REDIS_ADDR"); v != "" {
		c.Cache.Redis.Addr = v
	}
}

// GetAPIEndpoint returns the current API domain
func (c *Config) GetAPIEndpoint() string {
	switch c.APIDomain.CurrentDomain {
	case "backup":
		return c.APIDomain.Backup
	case "shanghai":
		return c.APIDomain.Shanghai
	case "shenzhen":
		return c.APIDomain.Shenzhen
	case "hong_kong":
		return c.APIDomain.HongKong
	default:
		return c.APIDomain.Primary
	}
}

// GetReadTimeout returns read timeout as duration
func (c *Config) GetReadTimeout() time.Duration {
	return time.Duration(c.Server.ReadTimeout) * time.Second
}

// GetWriteTimeout returns write timeout as duration
func (c *Config) GetWriteTimeout() time.Duration {
	return time.Duration(c.Server.WriteTimeout) * time.Second
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.WeChat.AppID == "" {
		return fmt.Errorf("app_id is required")
	}
	if c.WeChat.AppSecret == "" {
		return fmt.Errorf("app_secret is required")
	}
	if c.WeChat.Token == "" {
		return fmt.Errorf("token is required")
	}
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}
	return nil
}
