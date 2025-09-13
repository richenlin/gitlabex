package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gitlabex/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims JWT声明结构 - 简化结构，只包含GitLab访问令牌
type JWTClaims struct {
	GitLabAccessToken string `json:"gitlab_access_token"`
	jwt.RegisteredClaims
}

// GitLabUser GitLab用户信息
type GitLabUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	IsAdmin  bool   `json:"is_admin"`
}

// getGitLabUser 从GitLab API获取用户信息
func getGitLabUser(accessToken, gitlabURL string) (*GitLabUser, error) {
	url := fmt.Sprintf("%s/api/v4/user", gitlabURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API返回错误: %s", resp.Status)
	}

	var user GitLabUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
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

		// 验证GitLab访问令牌并获取用户信息
		gitlabUser, err := getGitLabUser(claims.GitLabAccessToken, cfg.GitLabURL)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "gitlab_auth_failed",
				"message": "GitLab authentication failed: " + err.Error(),
			})
			c.Abort()
			return
		}

		// 将GitLab用户信息存储到上下文中
		c.Set("gitlab_access_token", claims.GitLabAccessToken)
		c.Set("gitlab_user_id", gitlabUser.ID)
		c.Set("username", gitlabUser.Username)
		c.Set("email", gitlabUser.Email)
		c.Set("name", gitlabUser.Name)
		c.Set("is_admin", gitlabUser.IsAdmin)

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
			c.Set("userID", "")
			c.Set("username", "guest")
			c.Set("email", "")
			c.Set("role", "guest") // guest role
			c.Next()
			return
		}

		// 检查Bearer格式
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			// token格式错误，设置为游客模式
			c.Set("is_guest", true)
			c.Set("userID", "")
			c.Set("username", "guest")
			c.Set("email", "")
			c.Set("role", "guest")
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
			c.Set("userID", "")
			c.Set("username", "guest")
			c.Set("email", "")
			c.Set("role", "guest")
			c.Next()
			return
		}

		// 获取声明
		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			// claims解析失败，设置为游客模式
			c.Set("is_guest", true)
			c.Set("userID", "")
			c.Set("username", "guest")
			c.Set("email", "")
			c.Set("role", "guest")
			c.Next()
			return
		}

		// 验证GitLab访问令牌并获取用户信息
		gitlabUser, err := getGitLabUser(claims.GitLabAccessToken, cfg.GitLabURL)
		if err != nil {
			// GitLab认证失败，设置为游客模式
			c.Set("is_guest", true)
			c.Set("username", "guest")
			c.Set("email", "")
			c.Next()
			return
		}

		// 设置已登录用户信息
		c.Set("is_guest", false)
		c.Set("gitlab_access_token", claims.GitLabAccessToken)
		c.Set("gitlab_user_id", gitlabUser.ID)
		c.Set("username", gitlabUser.Username)
		c.Set("email", gitlabUser.Email)
		c.Set("name", gitlabUser.Name)
		c.Set("is_admin", gitlabUser.IsAdmin)

		c.Next()
	}
}
