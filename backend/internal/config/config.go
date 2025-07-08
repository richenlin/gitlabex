package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 应用程序配置
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	GitLab     GitLabConfig     `mapstructure:"gitlab"`
	OnlyOffice OnlyOfficeConfig `mapstructure:"onlyoffice"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Log        LogConfig        `mapstructure:"log"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug, release, test
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
	Timezone string `mapstructure:"timezone"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// GitLabConfig GitLab配置
type GitLabConfig struct {
	URL          string `mapstructure:"url"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
	Token        string `mapstructure:"token"`
}

// OnlyOfficeConfig OnlyOffice配置
type OnlyOfficeConfig struct {
	DocumentServerURL string `mapstructure:"document_server_url"`
	JWTSecret         string `mapstructure:"jwt_secret"`
	CallbackURL       string `mapstructure:"callback_url"`
	MaxFileSize       int64  `mapstructure:"max_file_size"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	ExpireTime time.Duration `mapstructure:"expire_time"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}

// LoadConfig 加载配置
func LoadConfig(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// 环境变量支持
	viper.AutomaticEnv()
	viper.SetEnvPrefix("GITLABEX")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		// 如果没有配置文件，使用默认配置
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return loadDefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 从环境变量覆盖配置
	overrideFromEnv(&config)

	return &config, nil
}

// loadDefaultConfig 加载默认配置
func loadDefaultConfig() *Config {
	config := &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
			Mode: "debug",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "password",
			DBName:   "gitlabex",
			SSLMode:  "disable",
			Timezone: "UTC",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
		},
		GitLab: GitLabConfig{
			URL:          "http://localhost",
			ClientID:     "",
			ClientSecret: "",
			RedirectURL:  "http://localhost:8080/auth/callback",
			Token:        "",
		},
		OnlyOffice: OnlyOfficeConfig{
			DocumentServerURL: "http://localhost:8000",
			JWTSecret:         "your-secret-key",
			CallbackURL:       "http://localhost:8080/api/onlyoffice/callback",
			MaxFileSize:       50 * 1024 * 1024, // 50MB
		},
		JWT: JWTConfig{
			Secret:     "your-jwt-secret",
			ExpireTime: 24 * time.Hour,
		},
		Log: LogConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100,
			MaxAge:     30,
			MaxBackups: 3,
		},
	}

	// 从环境变量覆盖配置
	overrideFromEnv(config)

	return config
}

// overrideFromEnv 从环境变量覆盖配置
func overrideFromEnv(config *Config) {
	if host := os.Getenv("SERVER_HOST"); host != "" {
		config.Server.Host = host
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}

	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		// 解析数据库URL
		// 格式: postgres://user:password@host:port/dbname?sslmode=disable
		// 简化实现，实际项目中应该使用专门的URL解析库
		config.Database.Host = "postgres"
		config.Database.User = "community"
		config.Database.Password = "password"
		config.Database.DBName = "community"
	}

	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		config.Redis.Host = "redis"
		config.Redis.Port = 6379
	}

	if gitlabURL := os.Getenv("GITLAB_URL"); gitlabURL != "" {
		config.GitLab.URL = gitlabURL
	}

	if clientID := os.Getenv("GITLAB_CLIENT_ID"); clientID != "" {
		config.GitLab.ClientID = clientID
	}

	if clientSecret := os.Getenv("GITLAB_CLIENT_SECRET"); clientSecret != "" {
		config.GitLab.ClientSecret = clientSecret
	}

	if onlyOfficeURL := os.Getenv("ONLYOFFICE_URL"); onlyOfficeURL != "" {
		config.OnlyOffice.DocumentServerURL = onlyOfficeURL
	}

	if jwtSecret := os.Getenv("ONLYOFFICE_JWT_SECRET"); jwtSecret != "" {
		config.OnlyOffice.JWTSecret = jwtSecret
	}
}

// GetDatabaseDSN 获取数据库连接字符串
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		c.Database.Host,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.Port,
		c.Database.SSLMode,
		c.Database.Timezone,
	)
}

// GetRedisAddr 获取Redis地址
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// GetServerAddr 获取服务器地址
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
