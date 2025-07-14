package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gitlabex/internal/config"
	"gitlabex/internal/models"
	"gitlabex/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// OAuthMiddleware OAuth认证中间件
type OAuthMiddleware struct {
	config      *config.Config
	db          *gorm.DB
	userService *services.UserService
}

// NewOAuthMiddleware 创建OAuth认证中间件
func NewOAuthMiddleware(config *config.Config, db *gorm.DB, userService *services.UserService) *OAuthMiddleware {
	return &OAuthMiddleware{
		config:      config,
		db:          db,
		userService: userService,
	}
}

// ThirdPartyAuth 第三方API认证中间件
func (m *OAuthMiddleware) ThirdPartyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求中获取认证信息
		authHeader := c.GetHeader("Authorization")
		apiKey := c.GetHeader("X-API-Key")

		var user *models.User
		var err error

		// 优先尝试JWT Token认证
		if authHeader != "" {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			user, err = m.validateJWTToken(tokenString)
			if err == nil {
				c.Set("current_user", user)
				c.Set("auth_type", "jwt")
				c.Next()
				return
			}
		}

		// 尝试API Key认证
		if apiKey != "" {
			user, err = m.validateAPIKey(apiKey)
			if err == nil {
				c.Set("current_user", user)
				c.Set("auth_type", "api_key")
				c.Next()
				return
			}
		}

		// 认证失败
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication required",
			"message": "Please provide valid JWT token or API key",
		})
		c.Abort()
	}
}

// RequireRole 要求特定角色权限
func (m *OAuthMiddleware) RequireRole(roles ...int) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("current_user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		currentUser := user.(*models.User)

		// 检查用户角色
		hasPermission := false
		for _, role := range roles {
			if currentUser.Role == role {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":          "Insufficient permissions",
				"required_roles": roles,
				"user_role":      currentUser.Role,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireScope 要求特定OAuth范围
func (m *OAuthMiddleware) RequireScope(scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authType, exists := c.Get("auth_type")
		if !exists || authType != "api_key" {
			// JWT Token默认有所有权限，API Key需要检查scope
			c.Next()
			return
		}

		// 这里可以扩展API Key的scope检查逻辑
		// 暂时允许所有API Key访问所有scope
		c.Next()
	}
}

// validateJWTToken 验证JWT Token
func (m *OAuthMiddleware) validateJWTToken(tokenString string) (*models.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(m.config.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := uint(claims["user_id"].(float64))
		return m.userService.GetUserByID(userID)
	}

	return nil, jwt.ErrInvalidKey
}

// validateAPIKey 验证API Key
func (m *OAuthMiddleware) validateAPIKey(apiKey string) (*models.User, error) {
	// API Key格式验证
	if len(apiKey) < 32 {
		return nil, jwt.ErrInvalidKey
	}

	// 从API Key中提取用户信息（简化实现）
	// 实际生产环境中应该有专门的API Key管理表

	// 解析API Key：user_id.timestamp.signature
	parts := strings.Split(apiKey, ".")
	if len(parts) != 3 {
		return nil, jwt.ErrInvalidKey
	}

	userID, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return nil, err
	}

	timestamp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, err
	}

	// 检查时间戳（API Key 有效期7天）
	if time.Now().Unix()-timestamp > 7*24*3600 {
		return nil, jwt.ErrTokenExpired
	}

	// 验证签名
	expectedSig := m.generateAPIKeySignature(uint(userID), timestamp)
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return nil, jwt.ErrSignatureInvalid
	}

	return m.userService.GetUserByID(uint(userID))
}

// GenerateAPIKey 生成API Key
func (m *OAuthMiddleware) GenerateAPIKey(userID uint) string {
	timestamp := time.Now().Unix()
	signature := m.generateAPIKeySignature(userID, timestamp)
	return strconv.FormatUint(uint64(userID), 10) + "." + strconv.FormatInt(timestamp, 10) + "." + signature
}

// generateAPIKeySignature 生成API Key签名
func (m *OAuthMiddleware) generateAPIKeySignature(userID uint, timestamp int64) string {
	h := hmac.New(sha256.New, []byte(m.config.JWT.Secret))
	h.Write([]byte(strconv.FormatUint(uint64(userID), 10) + "." + strconv.FormatInt(timestamp, 10)))
	return hex.EncodeToString(h.Sum(nil))
}

// LogAPIAccess API访问日志中间件
func (m *OAuthMiddleware) LogAPIAccess() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		var userInfo string
		if user, exists := param.Keys["current_user"]; exists {
			currentUser := user.(*models.User)
			userInfo = " user=" + currentUser.Username
		}

		return fmt.Sprintf("[THIRD_PARTY_API] %v | %3d | %13v | %15s | %-7s %#v%s\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Method,
			param.Path,
			userInfo,
		)
	})
}

// CORS 跨域中间件
func (m *OAuthMiddleware) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 允许的域名列表（生产环境中应该严格控制）
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"http://localhost:8080",
			"http://127.0.0.1:8080",
		}

		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RateLimit 限流中间件（简化实现）
func (m *OAuthMiddleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 简化的限流实现
		// 生产环境中应该使用Redis等缓存系统

		// 获取客户端IP或用户ID
		clientID := c.ClientIP()
		if user, exists := c.Get("current_user"); exists {
			currentUser := user.(*models.User)
			clientID = "user_" + strconv.FormatUint(uint64(currentUser.ID), 10)
		}

		// 这里应该实现真正的限流逻辑
		// 暂时允许所有请求通过
		c.Set("client_id", clientID)
		c.Next()
	}
}

// isOAuthBypassEnabled 检查是否启用OAuth绕过（临时测试功能）
