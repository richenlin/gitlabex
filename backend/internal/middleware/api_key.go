package middleware

import (
	"fmt"
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
	MaxBatch    int      // 批量操作最大数量
	CanAdmin    bool     // 是否可以创建管理员
	CanUpdate   bool     // 是否可以更新用户
	CanQuery    bool     // 是否可以查询用户
	Sources     []string // 允许的外部系统来源
	RateLimit   int      // 每小时请求限制
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
	// 定义所有有效的API密钥及其权限 - 精简版本，只保留第三方密钥
	validKeys := map[string]APIKeyInfo{
		// 第三方密钥
		os.Getenv("THIRD_PARTY_API_KEY"): {
			Type:        ThirdPartyAPIKey,
			Description: "Third Party API Key",
			MaxBatch:    50,
			CanAdmin:    false,
			CanUpdate:   true,
			CanQuery:    true,
			Sources:     []string{"*"}, // 允许所有第三方系统
			RateLimit:   500,           // 每小时500次请求
		},
		// 开发环境第三方密钥（生产环境应移除）
		"gitlabex_third_party_api_key_2024_change_in_production": {
			Type:        ThirdPartyAPIKey,
			Description: "Development Third Party Key",
			MaxBatch:    50,
			CanAdmin:    false,
			CanUpdate:   true,
			CanQuery:    true,
			Sources:     []string{"*"},
			RateLimit:   500,
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

// RequireUpdatePermission 要求更新权限的中间件
func RequireUpdatePermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		keyInfo, exists := GetAPIKeyInfo(c)
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "服务器内部错误",
				"error":   "无法获取API密钥信息",
			})
			c.Abort()
			return
		}

		if !keyInfo.CanUpdate {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "权限不足",
				"error":   "当前API密钥无更新权限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireQueryPermission 要求查询权限的中间件
func RequireQueryPermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		keyInfo, exists := GetAPIKeyInfo(c)
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "服务器内部错误",
				"error":   "无法获取API密钥信息",
			})
			c.Abort()
			return
		}

		if !keyInfo.CanQuery {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "权限不足",
				"error":   "当前API密钥无查询权限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireSourcePermission 要求特定来源权限的中间件
func RequireSourcePermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		keyInfo, exists := GetAPIKeyInfo(c)
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "服务器内部错误",
				"error":   "无法获取API密钥信息",
			})
			c.Abort()
			return
		}

		// 如果允许所有来源，直接通过
		for _, source := range keyInfo.Sources {
			if source == "*" {
				c.Next()
				return
			}
		}

		// 检查请求中的外部系统来源
		externalSource := c.Query("external_source")
		if externalSource == "" {
			// 从请求体中获取
			var requestBody map[string]interface{}
			if c.ShouldBindJSON(&requestBody) == nil {
				if source, ok := requestBody["external_source"].(string); ok {
					externalSource = source
				}
			}
		}

		if externalSource == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "请求参数错误",
				"error":   "缺少external_source参数",
			})
			c.Abort()
			return
		}

		// 检查来源权限
		hasPermission := false
		for _, allowedSource := range keyInfo.Sources {
			if allowedSource == externalSource {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "权限不足",
				"error":   fmt.Sprintf("当前API密钥无权访问来源: %s", externalSource),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
