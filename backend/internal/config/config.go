package config

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
)

// Config 应用程序配置结构
type Config struct {
	// 服务器配置
	ServerHost  string
	ServerPort  string
	Environment string
	Debug       bool

	// 数据库配置
	DatabaseURL      string
	DatabaseHost     string
	DatabasePort     string
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string

	// Redis配置
	RedisHost     string
	RedisPort     string
	RedisPassword string

	// GitLab配置
	GitLabURL          string
	GitLabClientID     string
	GitLabClientSecret string
	GitLabRedirectURI  string
	GitLabScopes       string

	// JWT配置
	JWTSecret          string
	JWTExpirationHours int

	// 应用配置
	FrontendURL string

	// MinIO对象存储配置
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOUseSSL    bool
	MinIORegion    string

	// 日志配置
	LogLevel  string
	LogFormat string
}

// Load 加载配置
func Load() *Config {
	// 加载环境变量文件
	loadEnvFiles()

	// 加载OAuth配置
	loadOAuthConfig()
	cfg := &Config{
		// 服务器配置
		ServerHost:  getEnv("SERVER_HOST", "0.0.0.0"),
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		Environment: getEnv("APP_ENV", "development"),
		Debug:       getEnvAsBool("APP_DEBUG", true),

		// 数据库配置
		DatabaseHost:     getEnv("DB_HOST", "localhost"),
		DatabasePort:     getEnv("DB_PORT", "5432"),
		DatabaseUser:     getEnv("DB_USER", "gitlab"),
		DatabasePassword: getEnv("DB_PASSWORD", "password123"),
		DatabaseName:     getEnv("DB_NAME", "gitlabex"),

		// Redis配置
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", "password123"),

		// GitLab配置
		GitLabURL:          getEnv("GITLAB_URL", "http://localhost:8081"),
		GitLabClientID:     getEnv("GITLAB_CLIENT_ID", ""),
		GitLabClientSecret: getEnv("GITLAB_CLIENT_SECRET", ""),
		GitLabRedirectURI:  getEnv("GITLAB_REDIRECT_URI", "http://localhost:3000/auth/gitlab/callback"),
		GitLabScopes:       getEnv("SCOPES", "api read_api openid"),

		// JWT配置
		JWTSecret:          getEnv("JWT_SECRET", "your_jwt_secret_key_here_please_change_in_production"),
		JWTExpirationHours: getEnvAsInt("JWT_EXPIRATION_HOURS", 24),

		// 应用配置
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),

		// MinIO对象存储配置
		MinIOEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey: getEnv("MINIO_ACCESS_KEY", "admin"),
		MinIOSecretKey: getEnv("MINIO_SECRET_KEY", "password123"),
		MinIOUseSSL:    getEnvAsBool("MINIO_USE_SSL", false),
		MinIORegion:    getEnv("MINIO_REGION", "us-east-1"),

		// 日志配置
		LogLevel:  getEnv("LOG_LEVEL", "debug"),
		LogFormat: getEnv("LOG_FORMAT", "json"),
	}

	// 构建数据库URL
	cfg.DatabaseURL = buildDatabaseURL(cfg)

	return cfg
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

// buildDatabaseURL 构建数据库连接URL
func buildDatabaseURL(cfg *Config) string {
	// 优先使用完整的DATABASE_URL
	if databaseURL := getEnv("DATABASE_URL", ""); databaseURL != "" {
		return databaseURL
	}

	// 否则构建URL
	return "host=" + cfg.DatabaseHost +
		" port=" + cfg.DatabasePort +
		" user=" + cfg.DatabaseUser +
		" password=" + cfg.DatabasePassword +
		" dbname=" + cfg.DatabaseName +
		" sslmode=disable TimeZone=Asia/Shanghai"
}

// loadEnvFiles 加载环境变量文件
func loadEnvFiles() {
	// 尝试加载不同位置的环境文件
	envFiles := []string{
		"/app/config/backend.env",  // Docker容器内路径
		"./config/backend.env",     // 本地开发路径（从项目根目录）
		"../config/backend.env",    // 从backend目录启动时的路径
		"../../config/backend.env", // 更深层目录的路径
		"./.env",                   // 项目根目录
		"backend.env",              // 当前目录
	}

	for _, envFile := range envFiles {
		if err := loadEnvFile(envFile); err == nil {
			log.Printf("Loaded environment file: %s", envFile)
			break
		}
	}
}

// loadOAuthConfig 加载OAuth配置
func loadOAuthConfig() {
	// 尝试加载OAuth配置文件
	oauthFiles := []string{
		"/app/config/oauth.env",  // Docker容器内路径
		"./config/oauth.env",     // 本地开发路径（从项目根目录）
		"../config/oauth.env",    // 从backend目录启动时的路径
		"../../config/oauth.env", // 更深层目录的路径
		"oauth.env",              // 当前目录
	}

	for _, oauthFile := range oauthFiles {
		if err := loadEnvFile(oauthFile); err == nil {
			log.Printf("Loaded OAuth config file: %s", oauthFile)
			break
		}
	}
}

// loadEnvFile 加载指定的环境变量文件
func loadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释行
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 解析 KEY=VALUE 格式
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// 移除值两端的引号
		if len(value) >= 2 {
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
				(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}
		}

		// 只在环境变量不存在时设置
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}
