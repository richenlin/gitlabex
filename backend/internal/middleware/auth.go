package middleware

import (
	"net/http"
	"strings"

	"gitlabex/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims JWT声明结构
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     int    `json:"role"`
	jwt.RegisteredClaims
}

// RequireAuth JWT认证中间件
func RequireAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header中获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "missing_authorization_header",
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// 检查Bearer格式
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_authorization_format",
				"message": "Authorization header must be Bearer token",
			})
			c.Abort()
			return
		}

		// 解析JWT Token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// 验证签名方法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_token",
				"message": "Failed to parse token: " + err.Error(),
			})
			c.Abort()
			return
		}

		// 验证token有效性
		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_token",
				"message": "Token is not valid",
			})
			c.Abort()
			return
		}

		// 获取声明
		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_claims",
				"message": "Failed to parse token claims",
			})
			c.Abort()
			return
		}

		// 将用户信息设置到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// OptionalAuth 可选认证中间件 - 支持游客访问，但如果有token则验证
func OptionalAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header中获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 没有token，设置为游客模式
			c.Set("is_guest", true)
			c.Set("user_id", "")
			c.Set("username", "guest")
			c.Set("email", "")
			c.Set("role", 0) // guest role
			c.Next()
			return
		}

		// 检查Bearer格式
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			// token格式错误，设置为游客模式
			c.Set("is_guest", true)
			c.Set("user_id", "")
			c.Set("username", "guest")
			c.Set("email", "")
			c.Set("role", 0)
			c.Next()
			return
		}

		// 解析JWT Token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// 验证签名方法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			// token无效，设置为游客模式
			c.Set("is_guest", true)
			c.Set("user_id", "")
			c.Set("username", "guest")
			c.Set("email", "")
			c.Set("role", 0)
			c.Next()
			return
		}

		// 获取声明
		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			// claims解析失败，设置为游客模式
			c.Set("is_guest", true)
			c.Set("user_id", "")
			c.Set("username", "guest")
			c.Set("email", "")
			c.Set("role", 0)
			c.Next()
			return
		}

		// 设置已登录用户信息
		c.Set("is_guest", false)
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}
