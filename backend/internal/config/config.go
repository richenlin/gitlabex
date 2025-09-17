package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 应用程序配置结构
type Config struct {
	// 服务器配置
	Server ServerConfig `yaml:"server"`

	// 数据库配置
	Database DatabaseConfig `yaml:"database"`

	// Redis配置
	Redis RedisConfig `yaml:"redis"`

	// GitLab配置
	GitLab GitLabConfig `yaml:"gitlab"`

	// JWT配置
	JWT JWTConfig `yaml:"jwt"`

	// 应用配置
	App AppConfig `yaml:"app"`

	// MinIO对象存储配置
	MinIO MinIOConfig `yaml:"minio"`

	// 日志配置
	Logging LoggingConfig `yaml:"logging"`

	// 文件上传配置
	Upload UploadConfig `yaml:"upload"`

	// 安全配置
	Security SecurityConfig `yaml:"security"`

	// API密钥配置
	APIKeys APIKeysConfig `yaml:"api_keys"`

	// 向后兼容字段
	ServerHost         string
	ServerPort         string
	Environment        string
	Debug              bool
	DatabaseURL        string
	DatabaseHost       string
	DatabasePort       string
	DatabaseUser       string
	DatabasePassword   string
	DatabaseName       string
	RedisHost          string
	RedisPort          string
	RedisPassword      string
	GitLabURL          string
	GitLabClientID     string
	GitLabClientSecret string
	GitLabRedirectURI  string
	GitLabScopes       string
	GitLabSystemToken  string
	JWTSecret          string
	JWTExpirationHours int

	MinIOEndpoint    string
	MinIOAccessKey   string
	MinIOSecretKey   string
	MinIOUseSSL      bool
	MinIORegion      string
	LogLevel         string
	LogFormat        string
	AllowedFileTypes string
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host        string `yaml:"host"`
	Port        string `yaml:"port"`
	Environment string `yaml:"environment"`
	Debug       bool   `yaml:"debug"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	URL      string `yaml:"url"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
}

// GitLabConfig GitLab配置
type GitLabConfig struct {
	URL          string `yaml:"url"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURI  string `yaml:"redirect_uri"`
	Scopes       string `yaml:"scopes"`
	SystemToken  string `yaml:"system_token"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret          string `yaml:"secret"`
	ExpirationHours int    `yaml:"expiration_hours"`
}

// AppConfig 应用配置
type AppConfig struct {
	FrontendURL    string `yaml:"frontend_url"`
	AllowedOrigins string `yaml:"allowed_origins"`
}

// MinIOConfig MinIO配置
type MinIOConfig struct {
	Endpoint  string `yaml:"endpoint"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	UseSSL    bool   `yaml:"use_ssl"`
	Region    string `yaml:"region"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// UploadConfig 文件上传配置
type UploadConfig struct {
	AllowedFileTypes string `yaml:"allowed_file_types"`
	MaxSize          string `yaml:"max_size"`
	Path             string `yaml:"path"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	CORSAllowedOrigins   string `yaml:"cors_allowed_origins"`
	CORSAllowedMethods   string `yaml:"cors_allowed_methods"`
	CORSAllowedHeaders   string `yaml:"cors_allowed_headers"`
	CORSAllowCredentials bool   `yaml:"cors_allow_credentials"`
	RateLimitRPM         int    `yaml:"rate_limit_rpm"`
	RateLimitBurst       int    `yaml:"rate_limit_burst"`
}

// APIKeysConfig API密钥配置
type APIKeysConfig struct {
	SyncAPIKey       string `yaml:"sync_api_key"`
	ThirdPartyAPIKey string `yaml:"third_party_api_key"`
}

// Load 加载配置
func Load() *Config {
	// 从YAML配置文件加载
	cfg := loadFromYAML()
	if cfg == nil {
		log.Fatal("无法加载YAML配置文件，请确保配置文件存在且格式正确")
	}

	log.Println("从YAML配置文件加载成功")

	// 应用环境变量覆盖（支持跨服务器部署）
	applyEnvironmentOverrides(cfg)

	// 设置向后兼容字段
	setBackwardCompatibilityFields(cfg)

	// 构建数据库URL
	cfg.DatabaseURL = buildDatabaseURL(cfg)

	return cfg
}

// loadFromYAML 从YAML配置文件加载配置
func loadFromYAML() *Config {
	// 获取当前环境
	env := getEnv("APP_ENV", "development")

	// 支持cross-server环境
	if env == "cross-server" {
		env = "production" // cross-server使用production配置作为基础
	}

	// 配置文件查找路径
	configPaths := []string{
		"/app/config/config.yml",                       // Docker容器内路径
		"./config/config.yml",                          // 本地开发路径（从项目根目录）
		"../config/config.yml",                         // 从backend目录启动时的路径
		"../../config/config.yml",                      // 更深层目录的路径
		fmt.Sprintf("/app/config/config.%s.yml", env),  // 环境特定配置（Docker）
		fmt.Sprintf("./config/config.%s.yml", env),     // 环境特定配置（本地）
		fmt.Sprintf("../config/config.%s.yml", env),    // 环境特定配置（backend目录）
		fmt.Sprintf("../../config/config.%s.yml", env), // 环境特定配置（更深层目录）
	}

	var cfg *Config
	var baseConfigLoaded bool

	// 首先加载基础配置文件
	for _, configPath := range configPaths {
		if strings.Contains(configPath, fmt.Sprintf(".%s.", env)) {
			continue // 跳过环境特定配置，稍后处理
		}

		if data, err := os.ReadFile(configPath); err == nil {
			cfg = &Config{}
			if err := yaml.Unmarshal(data, cfg); err == nil {
				log.Printf("加载基础配置文件: %s", configPath)
				baseConfigLoaded = true
				break
			} else {
				log.Printf("解析配置文件失败 %s: %v", configPath, err)
			}
		}
	}

	if !baseConfigLoaded {
		return nil
	}

	// 然后加载环境特定配置文件（如果存在）
	for _, configPath := range configPaths {
		if !strings.Contains(configPath, fmt.Sprintf(".%s.", env)) {
			continue // 只处理环境特定配置
		}

		if data, err := os.ReadFile(configPath); err == nil {
			envCfg := &Config{}
			if err := yaml.Unmarshal(data, envCfg); err == nil {
				log.Printf("加载环境特定配置文件: %s", configPath)
				// 合并环境特定配置到基础配置
				mergeConfigs(cfg, envCfg)
				break
			} else {
				log.Printf("解析环境配置文件失败 %s: %v", configPath, err)
			}
		}
	}

	return cfg
}

// mergeConfigs 合并配置
func mergeConfigs(base, override *Config) {
	// 使用反射或手动合并非空字段
	// 这里简化处理，只合并主要配置
	if override.Server.Host != "" {
		base.Server.Host = override.Server.Host
	}
	if override.Server.Port != "" {
		base.Server.Port = override.Server.Port
	}
	if override.Server.Environment != "" {
		base.Server.Environment = override.Server.Environment
	}

	if override.Database.Host != "" {
		base.Database.Host = override.Database.Host
	}
	if override.Database.Port != "" {
		base.Database.Port = override.Database.Port
	}
	if override.Database.User != "" {
		base.Database.User = override.Database.User
	}
	if override.Database.Name != "" {
		base.Database.Name = override.Database.Name
	}

	// 继续合并其他配置...
}

// setBackwardCompatibilityFields 设置向后兼容字段
func setBackwardCompatibilityFields(cfg *Config) {
	cfg.ServerHost = cfg.Server.Host
	cfg.ServerPort = cfg.Server.Port
	cfg.Environment = cfg.Server.Environment
	cfg.Debug = cfg.Server.Debug

	cfg.DatabaseHost = cfg.Database.Host
	cfg.DatabasePort = cfg.Database.Port
	cfg.DatabaseUser = cfg.Database.User
	cfg.DatabasePassword = cfg.Database.Password
	cfg.DatabaseName = cfg.Database.Name

	cfg.RedisHost = cfg.Redis.Host
	cfg.RedisPort = cfg.Redis.Port
	cfg.RedisPassword = cfg.Redis.Password

	cfg.GitLabURL = cfg.GitLab.URL
	cfg.GitLabClientID = cfg.GitLab.ClientID
	cfg.GitLabClientSecret = cfg.GitLab.ClientSecret
	cfg.GitLabRedirectURI = cfg.GitLab.RedirectURI
	cfg.GitLabScopes = cfg.GitLab.Scopes
	cfg.GitLabSystemToken = cfg.GitLab.SystemToken

	cfg.JWTSecret = cfg.JWT.Secret
	cfg.JWTExpirationHours = cfg.JWT.ExpirationHours

	cfg.MinIOEndpoint = cfg.MinIO.Endpoint
	cfg.MinIOAccessKey = cfg.MinIO.AccessKey
	cfg.MinIOSecretKey = cfg.MinIO.SecretKey
	cfg.MinIOUseSSL = cfg.MinIO.UseSSL
	cfg.MinIORegion = cfg.MinIO.Region

	cfg.LogLevel = cfg.Logging.Level
	cfg.LogFormat = cfg.Logging.Format

	cfg.AllowedFileTypes = cfg.Upload.AllowedFileTypes
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为整数
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool 获取环境变量并转换为布尔值
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// applyEnvironmentOverrides 应用环境变量覆盖（支持跨服务器部署）
func applyEnvironmentOverrides(cfg *Config) {
	// GitLab相关配置（最关键的网络地址配置）
	if gitlabURL := os.Getenv("GITLAB_URL"); gitlabURL != "" {
		cfg.GitLab.URL = gitlabURL
		log.Printf("使用环境变量覆盖GitLab URL: %s", gitlabURL)
	}

	if gitlabRedirectURI := os.Getenv("GITLAB_REDIRECT_URI"); gitlabRedirectURI != "" {
		cfg.GitLab.RedirectURI = gitlabRedirectURI
		log.Printf("使用环境变量覆盖GitLab重定向URI: %s", gitlabRedirectURI)
	}

	// 前端URL配置
	if frontendURL := os.Getenv("FRONTEND_URL"); frontendURL != "" {
		cfg.App.FrontendURL = frontendURL
		log.Printf("使用环境变量覆盖前端URL: %s", frontendURL)
	}

	// CORS配置
	if corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); corsOrigins != "" {
		cfg.Security.CORSAllowedOrigins = corsOrigins
		log.Printf("使用环境变量覆盖CORS允许来源: %s", corsOrigins)
	}

	// 数据库配置（支持外部数据库）
	if dbHost := os.Getenv("DATABASE_HOST"); dbHost != "" {
		cfg.Database.Host = dbHost
		log.Printf("使用环境变量覆盖数据库主机: %s", dbHost)
	}

	if dbPort := os.Getenv("DATABASE_PORT"); dbPort != "" {
		cfg.Database.Port = dbPort
	}

	if dbUser := os.Getenv("DATABASE_USER"); dbUser != "" {
		cfg.Database.User = dbUser
	}

	if dbPassword := os.Getenv("DATABASE_PASSWORD"); dbPassword != "" {
		cfg.Database.Password = dbPassword
	}

	if dbName := os.Getenv("DATABASE_NAME"); dbName != "" {
		cfg.Database.Name = dbName
	}

	// Redis配置（支持外部Redis）
	if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" {
		cfg.Redis.Host = redisHost
		log.Printf("使用环境变量覆盖Redis主机: %s", redisHost)
	}

	if redisPort := os.Getenv("REDIS_PORT"); redisPort != "" {
		cfg.Redis.Port = redisPort
	}

	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		cfg.Redis.Password = redisPassword
	}

	// 服务器配置
	if serverHost := os.Getenv("SERVER_HOST"); serverHost != "" {
		cfg.Server.Host = serverHost
		log.Printf("使用环境变量覆盖服务器主机: %s", serverHost)
	}

	if serverPort := os.Getenv("SERVER_PORT"); serverPort != "" {
		cfg.Server.Port = serverPort
	}
}

// buildDatabaseURL 构建数据库连接URL
func buildDatabaseURL(cfg *Config) string {
	// 优先使用完整的DATABASE_URL
	if cfg.Database.URL != "" {
		return cfg.Database.URL
	}

	// 否则构建URL
	return "host=" + cfg.Database.Host +
		" port=" + cfg.Database.Port +
		" user=" + cfg.Database.User +
		" password=" + cfg.Database.Password +
		" dbname=" + cfg.Database.Name +
		" sslmode=disable TimeZone=Asia/Shanghai"
}
