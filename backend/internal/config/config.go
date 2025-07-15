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
	Frontend   FrontendConfig
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
	URL          string // 外部访问URL，用于OAuth授权
	InternalURL  string // 内部访问URL，用于API调用
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Token        string
	Scopes       string // OAuth权限范围
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

// FrontendConfig 前端配置
type FrontendConfig struct {
	URL string // 前端应用URL
}

func LoadConfig() (*Config, error) {
	// 1. 加载应用基础配置
	configPaths := []string{
		"config/app.env",
		"../config/app.env",
	}

	configLoaded := false
	for _, path := range configPaths {
		if err := godotenv.Load(path); err == nil {
			fmt.Printf("Loaded application config from: %s\n", path)
			configLoaded = true
			break
		}
	}

	if !configLoaded {
		fmt.Println("Warning: No application config file found")
	}

	// 2. 加载OAuth配置
	oauthPaths := []string{
		"config/oauth.env",
		"../config/oauth.env",
	}

	oauthLoaded := false
	for _, path := range oauthPaths {
		if err := godotenv.Load(path); err == nil {
			fmt.Printf("Loaded OAuth config from: %s\n", path)
			oauthLoaded = true
			break
		}
	}

	if !oauthLoaded {
		fmt.Println("Warning: No OAuth config file found")
	}

	// 3. 创建配置实例
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
			URL:          getEnv("GITLAB_EXTERNAL_URL", "http://localhost:8081"),
			InternalURL:  getEnv("GITLAB_INTERNAL_URL", "http://localhost:8081"),
			ClientID:     getEnv("GITLAB_CLIENT_ID", ""),
			ClientSecret: getEnv("GITLAB_CLIENT_SECRET", ""),
			RedirectURI:  getEnv("GITLAB_REDIRECT_URI", "http://localhost:8080/api/auth/gitlab/callback"),
			Token:        getEnv("GITLAB_TOKEN", ""),
			Scopes:       getEnv("GITLAB_SCOPES", "api read_user email"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-jwt-secret-key"),
		},
		Frontend: FrontendConfig{
			URL: getEnv("FRONTEND_URL", "http://localhost:3000"),
		},
	}

	// 4. 验证必要的配置
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
