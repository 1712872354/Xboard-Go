package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 全局配置结构体
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Log       LogConfig       `mapstructure:"log"`
	App       AppConfig       `mapstructure:"app"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	GRPC      GRPCConfig      `mapstructure:"grpc"`
	Captcha   CaptchaConfig   `mapstructure:"captcha"`
}

// GRPCConfig gRPC 服务配置
type GRPCConfig struct {
	Enabled bool `mapstructure:"enabled"` // 是否启用 gRPC 服务
	Port    int  `mapstructure:"port"`    // gRPC 监听端口
}

// ServerConfig 服务配置
type ServerConfig struct {
	Host           string   `mapstructure:"host"`
	Port           int      `mapstructure:"port"`
	Mode           string   `mapstructure:"mode"`
	AllowedOrigins []string `mapstructure:"allowed_origins"` // CORS允许的源
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string `mapstructure:"driver"`
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	SSLMode         string `mapstructure:"sslmode"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret          string `mapstructure:"secret"`
	AccessTokenTTL  int    `mapstructure:"access_token_ttl"`
	RefreshTokenTTL int    `mapstructure:"refresh_token_ttl"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level    string `mapstructure:"level"`
	Format   string `mapstructure:"format"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name                 string `mapstructure:"name"`
	Version              string `mapstructure:"version"`
	BcryptCost           int    `mapstructure:"bcrypt_cost"`
	DefaultUserRole      string `mapstructure:"default_user_role"`
	SubscribeTokenLength int    `mapstructure:"subscribe_token_length"`
	NodeAPIKey           string `mapstructure:"node_api_key"` // 节点API密钥，用于XrayR对接
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled       bool     `mapstructure:"enabled"`        // 是否启用限流
	IPLimit       int      `mapstructure:"ip_limit"`       // IP 每分钟最大请求数
	UserLimit     int      `mapstructure:"user_limit"`     // 用户每分钟最大请求数
	IPWhitelist   []string `mapstructure:"ip_whitelist"`   // IP 白名单
	PathWhitelist []string `mapstructure:"path_whitelist"` // 路径白名单
}

// CaptchaConfig 验证码配置
type CaptchaConfig struct {
	Provider  string  `mapstructure:"provider"`   // turnstile, recaptcha_v2, recaptcha_v3
	SiteKey   string  `mapstructure:"site_key"`   // 站点密钥
	SecretKey string  `mapstructure:"secret_key"` // 私密密钥
	MinScore  float64 `mapstructure:"min_score"`  // reCAPTCHA v3 最小分数
}

var globalConfig *Config

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// 支持环境变量覆盖
	v.SetEnvPrefix("XBOARD")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	globalConfig = &cfg
	return &cfg, nil
}

// Get 获取全局配置
func Get() *Config {
	if globalConfig == nil {
		panic("config not loaded, call Load first")
	}
	return globalConfig
}

// DSN 获取数据库连接字符串
func (c *DatabaseConfig) DSN() string {
	switch c.Driver {
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			c.User, c.Password, c.Host, c.Port, c.DBName)
	case "sqlite", "sqlite3":
		return c.DBName
	default:
		return ""
	}
}

// Addr 获取服务监听地址
func (c *ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// RedisAddr 获取 Redis 地址
func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
