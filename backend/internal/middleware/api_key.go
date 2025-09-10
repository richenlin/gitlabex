package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// APIKeyType 定义API密钥类型
type APIKeyType string

const (
	SyncAPIKey       APIKeyType = "sync"
	ThirdPartyAPIKey APIKeyType = "third_party"
)

// APIKeyInfo 存储API密钥信息
type APIKeyInfo struct {
	Type        APIKeyType
	Description string
	MaxBatch    int  // 批量操作最大数量
	CanAdmin    bool // 是否可以创建管理员
}

// RequireAPIKey API密钥认证中间件
func RequireAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "缺少API密钥",
				"error":   "请在请求头中包含 X-API-Key",
			})
			c.Abort()
			return
		}

		// 验证API密钥并获取权限信息
		keyInfo, valid := validateAPIKey(apiKey)
		if !valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "无效的API密钥",
				"error":   "提供的API密钥无效或已过期",
			})
			c.Abort()
			return
		}

		// 将密钥信息存储到上下文中供后续使用
		c.Set("api_key_info", keyInfo)
		c.Set("api_key_type", keyInfo.Type)

		// 记录API调用
		// log.Printf("API call from %s with %s key", c.ClientIP(), keyInfo.Type)

		c.Next()
	}
}

// RequireAPIKeyWithBatchLimit 带批量限制的API密钥中间件
func RequireAPIKeyWithBatchLimit() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 先执行基本的API密钥验证
		RequireAPIKey()(c)

		if c.IsAborted() {
			return
		}

		// 对于批量操作，检查批次限制
		if strings.Contains(c.Request.URL.Path, "/batch") {
			keyInfo, exists := c.Get("api_key_info")
			if !exists {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "服务器内部错误",
					"error":   "无法获取API密钥信息",
				})
				c.Abort()
				return
			}

			apiKeyInfo := keyInfo.(APIKeyInfo)
			c.Set("max_batch_size", apiKeyInfo.MaxBatch)
		}

		c.Next()
	})
}

// RequireAdminAPIKey 要求管理员权限的API密钥中间件
func RequireAdminAPIKey() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 先执行基本的API密钥验证
		RequireAPIKey()(c)

		if c.IsAborted() {
			return
		}

		// 检查是否有管理员权限
		keyInfo, exists := c.Get("api_key_info")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "服务器内部错误",
				"error":   "无法获取API密钥信息",
			})
			c.Abort()
			return
		}

		apiKeyInfo := keyInfo.(APIKeyInfo)
		if !apiKeyInfo.CanAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "权限不足",
				"error":   "当前API密钥无权执行管理员操作",
			})
			c.Abort()
			return
		}

		c.Next()
	})
}

// validateAPIKey 验证API密钥并返回权限信息
func validateAPIKey(apiKey string) (APIKeyInfo, bool) {
	// 定义所有有效的API密钥及其权限
	validKeys := map[string]APIKeyInfo{
		// 系统同步密钥（最高权限）
		os.Getenv("SYNC_API_KEY"): {
			Type:        SyncAPIKey,
			Description: "System Sync API Key",
			MaxBatch:    100,
			CanAdmin:    true,
		},
		// 第三方密钥（受限权限）
		os.Getenv("THIRD_PARTY_API_KEY"): {
			Type:        ThirdPartyAPIKey,
			Description: "Third Party API Key",
			MaxBatch:    20,
			CanAdmin:    false,
		},
		// 开发环境默认密钥（生产环境应移除）
		"gitlabex_sync_api_key_2024": {
			Type:        SyncAPIKey,
			Description: "Development Sync Key",
			MaxBatch:    100,
			CanAdmin:    true,
		},
		// 开发环境第三方密钥（生产环境应移除）
		"gitlabex_third_party_api_key_2024": {
			Type:        ThirdPartyAPIKey,
			Description: "Development Third Party Key",
			MaxBatch:    20,
			CanAdmin:    false,
		},
	}

	// 查找密钥信息
	for key, info := range validKeys {
		if key != "" && apiKey == key {
			return info, true
		}
	}

	return APIKeyInfo{}, false
}

// GetAPIKeyInfo 从上下文获取API密钥信息的辅助函数
func GetAPIKeyInfo(c *gin.Context) (APIKeyInfo, bool) {
	keyInfo, exists := c.Get("api_key_info")
	if !exists {
		return APIKeyInfo{}, false
	}
	return keyInfo.(APIKeyInfo), true
}
