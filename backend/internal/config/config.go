package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config 应用程序配置结构 - 现代化设计，无向后兼容字段
type Config struct {
	Server   ServerConfig   `yaml:"server" validate:"required"`
	Database DatabaseConfig `yaml:"database" validate:"required"`
	Redis    RedisConfig    `yaml:"redis" validate:"required"`
	GitLab   GitLabConfig   `yaml:"gitlab" validate:"required"`
	JWT      JWTConfig      `yaml:"jwt" validate:"required"`
	MinIO    MinIOConfig    `yaml:"minio" validate:"required"`
	Logging  LoggingConfig  `yaml:"logging" validate:"required"`
	Upload   UploadConfig   `yaml:"upload" validate:"required"`
	Security SecurityConfig `yaml:"security" validate:"required"`
	APIKeys  APIKeysConfig  `yaml:"api_keys" validate:"required"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host        string `yaml:"host" validate:"required" default:"0.0.0.0"`
	Port        string `yaml:"port" validate:"required,numeric" default:"8080"`
	Environment string `yaml:"environment" validate:"required,oneof=development production test" default:"production"`
	Debug       bool   `yaml:"debug" default:"false"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `yaml:"host" validate:"required" default:"localhost"`
	Port     string `yaml:"port" validate:"required,numeric" default:"5432"`
	User     string `yaml:"user" validate:"required" default:"gitlabex"`
	Password string `yaml:"password" validate:"required"`
	Name     string `yaml:"name" validate:"required" default:"gitlabex"`
	SSLMode  string `yaml:"ssl_mode" validate:"oneof=disable require verify-ca verify-full" default:"disable"`
	TimeZone string `yaml:"timezone" default:"Asia/Shanghai"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `yaml:"host" validate:"required" default:"localhost"`
	Port     string `yaml:"port" validate:"required,numeric" default:"6379"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db" validate:"min=0,max=15" default:"0"`
}

// GitLabConfig GitLab配置
type GitLabConfig struct {
	URL          string `yaml:"url" validate:"required,url"`
	ClientID     string `yaml:"client_id" validate:"required"`
	ClientSecret string `yaml:"client_secret" validate:"required"`
	RedirectURI  string `yaml:"redirect_uri" validate:"required,url"`
	Scopes       string `yaml:"scopes" default:"api read_api openid"`
	SystemToken  string `yaml:"system_token" validate:"required"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret          string `yaml:"secret" validate:"required,min=32"`
	ExpirationHours int    `yaml:"expiration_hours" validate:"required,min=1,max=168" default:"24"`
}

// MinIOConfig MinIO配置
type MinIOConfig struct {
	Endpoint  string `yaml:"endpoint" validate:"required"`
	AccessKey string `yaml:"access_key" validate:"required"`
	SecretKey string `yaml:"secret_key" validate:"required"`
	UseSSL    bool   `yaml:"use_ssl" default:"false"`
	Region    string `yaml:"region" default:"us-east-1"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string `yaml:"level" validate:"required,oneof=debug info warn error" default:"info"`
	Format string `yaml:"format" validate:"required,oneof=json text" default:"json"`
}

// UploadConfig 文件上传配置
type UploadConfig struct {
	AllowedFileTypes string `yaml:"allowed_file_types" validate:"required"`
	MaxSize          string `yaml:"max_size" validate:"required" default:"50MB"`
	Path             string `yaml:"path" validate:"required" default:"/app/uploads"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	CORSAllowedOrigins   string `yaml:"cors_allowed_origins" validate:"required"`
	CORSAllowedMethods   string `yaml:"cors_allowed_methods" default:"GET,POST,PUT,DELETE,OPTIONS"`
	CORSAllowedHeaders   string `yaml:"cors_allowed_headers" default:"Content-Type,Authorization,X-Requested-With,X-API-Key"`
	CORSAllowCredentials bool   `yaml:"cors_allow_credentials" default:"true"`
	RateLimitRPM         int    `yaml:"rate_limit_rpm" validate:"min=1" default:"300"`
	RateLimitBurst       int    `yaml:"rate_limit_burst" validate:"min=1" default:"50"`
}

// APIKeysConfig API密钥配置
type APIKeysConfig struct {
	ThirdPartyAPIKey string `yaml:"third_party_api_key" validate:"required,min=32"`
}

// Load 加载配置 - 简化的现代化实现
func Load() *Config {
	cfg := loadFromYAML()
	if cfg == nil {
		log.Fatal("无法加载YAML配置文件，请确保配置文件存在且格式正确")
	}

	// 应用环境变量覆盖
	applyEnvironmentOverrides(cfg)

	// 验证配置
	if err := validateConfig(cfg); err != nil {
		log.Fatalf("配置验证失败: %v", err)
	}

	log.Printf("配置加载成功，环境: %s", cfg.Server.Environment)
	return cfg
}

// loadFromYAML 从YAML配置文件加载配置 - 简化的路径查找
func loadFromYAML() *Config {
	// 简化的配置文件路径，只支持两个位置
	configPaths := []string{
		"/app/config/config.yml", // Docker容器内路径
		"../config/config.yml",   // 本地开发路径（从backend目录运行）
	}

	for _, configPath := range configPaths {
		if data, err := os.ReadFile(configPath); err == nil {
			cfg := &Config{}
			if err := yaml.Unmarshal(data, cfg); err == nil {
				log.Printf("加载配置文件: %s", configPath)
				return cfg
			} else {
				log.Printf("解析配置文件失败 %s: %v", configPath, err)
			}
		}
	}

	return nil
}

// applyEnvironmentOverrides 应用环境变量覆盖 - 完整的覆盖支持
func applyEnvironmentOverrides(cfg *Config) {
	// 服务器配置
	if v := os.Getenv("SERVER_HOST"); v != "" {
		cfg.Server.Host = v
		log.Printf("环境变量覆盖: SERVER_HOST=%s", v)
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		cfg.Server.Port = v
	}
	if v := os.Getenv("APP_ENV"); v != "" {
		cfg.Server.Environment = v
		log.Printf("环境变量覆盖: APP_ENV=%s", v)
	}
	if v := os.Getenv("DEBUG"); v != "" {
		cfg.Server.Debug = getEnvAsBool("DEBUG", false)
	}

	// 数据库配置
	if v := os.Getenv("DATABASE_HOST"); v != "" {
		cfg.Database.Host = v
		log.Printf("环境变量覆盖: DATABASE_HOST=%s", v)
	}
	if v := os.Getenv("DATABASE_PORT"); v != "" {
		cfg.Database.Port = v
	}
	if v := os.Getenv("DATABASE_USER"); v != "" {
		cfg.Database.User = v
	}
	if v := os.Getenv("DATABASE_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("DATABASE_NAME"); v != "" {
		cfg.Database.Name = v
	}

	// Redis配置
	if v := os.Getenv("REDIS_HOST"); v != "" {
		cfg.Redis.Host = v
		log.Printf("环境变量覆盖: REDIS_HOST=%s", v)
	}
	if v := os.Getenv("REDIS_PORT"); v != "" {
		cfg.Redis.Port = v
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		cfg.Redis.Password = v
	}
	if v := os.Getenv("REDIS_DB"); v != "" {
		cfg.Redis.DB = getEnvAsInt("REDIS_DB", 0)
	}

	// GitLab配置
	if v := os.Getenv("GITLAB_URL"); v != "" {
		cfg.GitLab.URL = v
		log.Printf("环境变量覆盖: GITLAB_URL=%s", v)
	}
	if v := os.Getenv("GITLAB_CLIENT_ID"); v != "" {
		cfg.GitLab.ClientID = v
	}
	if v := os.Getenv("GITLAB_CLIENT_SECRET"); v != "" {
		cfg.GitLab.ClientSecret = v
	}
	if v := os.Getenv("GITLAB_REDIRECT_URI"); v != "" {
		cfg.GitLab.RedirectURI = v
		log.Printf("环境变量覆盖: GITLAB_REDIRECT_URI=%s", v)
	}
	if v := os.Getenv("GITLAB_SYSTEM_TOKEN"); v != "" {
		cfg.GitLab.SystemToken = v
	}

	// JWT配置
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.JWT.Secret = v
	}
	if v := os.Getenv("JWT_EXPIRATION_HOURS"); v != "" {
		cfg.JWT.ExpirationHours = getEnvAsInt("JWT_EXPIRATION_HOURS", 24)
	}

	// MinIO配置
	if v := os.Getenv("MINIO_ENDPOINT"); v != "" {
		cfg.MinIO.Endpoint = v
		log.Printf("环境变量覆盖: MINIO_ENDPOINT=%s", v)
	}
	if v := os.Getenv("MINIO_ACCESS_KEY"); v != "" {
		cfg.MinIO.AccessKey = v
	}
	if v := os.Getenv("MINIO_SECRET_KEY"); v != "" {
		cfg.MinIO.SecretKey = v
	}
	if v := os.Getenv("MINIO_USE_SSL"); v != "" {
		cfg.MinIO.UseSSL = getEnvAsBool("MINIO_USE_SSL", false)
	}
	if v := os.Getenv("MINIO_REGION"); v != "" {
		cfg.MinIO.Region = v
	}

	// API密钥配置
	if v := os.Getenv("THIRD_PARTY_API_KEY"); v != "" {
		cfg.APIKeys.ThirdPartyAPIKey = v
	}

	// 日志配置
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.Logging.Level = v
	}
	if v := os.Getenv("LOG_FORMAT"); v != "" {
		cfg.Logging.Format = v
	}
}

// validateConfig 验证配置 - 简单的验证逻辑
func validateConfig(cfg *Config) error {
	// 验证必需的配置项
	if cfg.Database.Password == "" {
		return fmt.Errorf("数据库密码不能为空")
	}
	if cfg.JWT.Secret == "" || len(cfg.JWT.Secret) < 32 {
		return fmt.Errorf("JWT密钥长度必须至少32个字符")
	}
	if cfg.GitLab.ClientID == "" || cfg.GitLab.ClientSecret == "" {
		return fmt.Errorf("GitLab OAuth配置不完整")
	}
	if cfg.APIKeys.ThirdPartyAPIKey == "" || len(cfg.APIKeys.ThirdPartyAPIKey) < 32 {
		return fmt.Errorf("第三方API密钥长度必须至少32个字符")
	}

	return nil
}

// GetDatabaseURL 构建数据库连接URL - 现代化的URL构建
func (cfg *Config) GetDatabaseURL() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
		cfg.Database.TimeZone,
	)
}

// GetRedisAddr 获取Redis地址
func (cfg *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)
}

// GetMinIOEndpoint 获取MinIO端点
func (cfg *Config) GetMinIOEndpoint() string {
	return cfg.MinIO.Endpoint
}

// String 返回配置的字符串表示（隐藏敏感信息）
func (cfg *Config) String() string {
	return fmt.Sprintf("Config{Environment=%s, Database=%s:%s, Redis=%s:%s, GitLab=%s}",
		cfg.Server.Environment,
		cfg.Database.Host, cfg.Database.Port,
		cfg.Redis.Host, cfg.Redis.Port,
		cfg.GitLab.URL,
	)
}

// 工具函数
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
