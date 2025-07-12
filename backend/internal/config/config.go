package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config 应用程序配置
type Config struct {
	Database   DatabaseConfig
	Redis      RedisConfig
	Server     ServerConfig
	OnlyOffice OnlyOfficeConfig
	GitLab     GitLabConfig
	JWT        JWTConfig
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string
	Mode string
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// GitLabConfig GitLab配置
type GitLabConfig struct {
	URL          string
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Token        string
}

// OnlyOfficeConfig OnlyOffice配置
type OnlyOfficeConfig struct {
	BaseURL     string
	JWTSecret   string
	CallbackURL string
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret string
}

func LoadConfig() (*Config, error) {
	// 尝试加载.env文件 - 优先查找config/app.env，然后是.env
	if err := godotenv.Load("config/app.env"); err != nil {
		if err := godotenv.Load("../config/app.env"); err != nil {
			if err := godotenv.Load(".env"); err != nil {
				fmt.Println("No .env file found, using environment variables")
			}
		}
	}

	config := &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "gitlabex"),
			Password: getEnv("DB_PASSWORD", "password123"),
			DBName:   getEnv("DB_NAME", "gitlabex"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", "password123"),
			DB:       0,
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		OnlyOffice: OnlyOfficeConfig{
			BaseURL:     getEnv("ONLYOFFICE_URL", "http://localhost:8000"),
			JWTSecret:   getEnv("ONLYOFFICE_JWT_SECRET", "your-jwt-secret"),
			CallbackURL: getEnv("ONLYOFFICE_CALLBACK_URL", "http://localhost:8080/api/documents/callback"),
		},
		GitLab: GitLabConfig{
			URL:          getEnv("GITLAB_URL", "http://localhost"),
			ClientID:     getEnv("GITLAB_CLIENT_ID", ""),
			ClientSecret: getEnv("GITLAB_CLIENT_SECRET", ""),
			RedirectURI:  getEnv("GITLAB_REDIRECT_URI", "http://localhost:8080/api/auth/gitlab/callback"),
			Token:        getEnv("GITLAB_TOKEN", ""),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-jwt-secret-key"),
		},
	}

	// 验证必要的配置
	if config.GitLab.ClientID == "" || config.GitLab.ClientSecret == "" {
		fmt.Println("Warning: GitLab OAuth configuration is missing. Authentication will not work properly.")
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}

// GetServerAddr 获取服务器地址
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf(":%s", c.Server.Port)
}

// GetBaseURL 获取GitLab基础URL
func (g *GitLabConfig) GetBaseURL() string {
	return g.URL
}
