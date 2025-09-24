package middleware

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gitlabex/internal/services"

	"github.com/gin-gonic/gin"
)

// CacheMiddleware 缓存中间件
func CacheMiddleware(redisService *services.RedisService, expiration time.Duration, keyPrefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果Redis服务不可用，跳过缓存
		if redisService == nil {
			c.Next()
			return
		}

		// 只对GET请求进行缓存
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		// 生成缓存键
		cacheKey := generateCacheKey(keyPrefix, c.Request.URL.Path, c.Request.URL.RawQuery)

		// 尝试从缓存获取数据
		var responseData interface{}
		ctx := context.Background()
		err := redisService.GetCache(ctx, cacheKey, &responseData)
		if err == nil {
			// 缓存命中，直接返回
			c.JSON(http.StatusOK, responseData)
			c.Abort()
			return
		}

		// 缓存未命中，继续处理请求
		// 创建一个自定义的ResponseWriter来捕获响应
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           make([]byte, 0),
		}
		c.Writer = writer

		c.Next()

		// 如果响应成功，将数据存入缓存
		if writer.statusCode == http.StatusOK && len(writer.body) > 0 {
			// 异步更新缓存
			go func() {
				ctx := context.Background()
				// 解析响应体为JSON对象
				var responseData interface{}
				if err := json.Unmarshal(writer.body, &responseData); err == nil {
					redisService.SetCache(ctx, cacheKey, responseData, expiration)
				}
			}()
		}
	}
}

// CacheWithTagsMiddleware 带标签的缓存中间件
func CacheWithTagsMiddleware(redisService *services.RedisService, expiration time.Duration, keyPrefix string, tags []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果Redis服务不可用，跳过缓存
		if redisService == nil {
			c.Next()
			return
		}

		// 只对GET请求进行缓存
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		// 生成缓存键
		cacheKey := generateCacheKey(keyPrefix, c.Request.URL.Path, c.Request.URL.RawQuery)

		// 尝试从缓存获取数据
		var responseData interface{}
		ctx := context.Background()
		err := redisService.GetCache(ctx, cacheKey, &responseData)
		if err == nil {
			// 缓存命中，直接返回
			c.JSON(http.StatusOK, responseData)
			c.Abort()
			return
		}

		// 缓存未命中，继续处理请求
		// 创建一个自定义的ResponseWriter来捕获响应
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           make([]byte, 0),
		}
		c.Writer = writer

		c.Next()

		// 如果响应成功，将数据存入缓存
		if writer.statusCode == http.StatusOK && len(writer.body) > 0 {
			// 异步更新缓存
			go func() {
				ctx := context.Background()
				// 解析响应体为JSON对象
				var responseData interface{}
				if err := json.Unmarshal(writer.body, &responseData); err == nil {
					redisService.SetCacheWithTags(ctx, cacheKey, responseData, tags, expiration)
				}
			}()
		}
	}
}

// InvalidateCacheMiddleware 缓存失效中间件
func InvalidateCacheMiddleware(redisService *services.RedisService, patterns []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只对POST、PUT、DELETE、PATCH请求进行缓存失效
		if c.Request.Method == "GET" {
			c.Next()
			return
		}

		c.Next()

		// 请求处理完成后，异步清理相关缓存
		go func() {
			ctx := context.Background()
			for _, pattern := range patterns {
				redisService.DeleteCachePattern(ctx, pattern)
			}
		}()
	}
}

// InvalidateCacheByTagMiddleware 根据标签失效缓存中间件
func InvalidateCacheByTagMiddleware(redisService *services.RedisService, tags []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只对POST、PUT、DELETE、PATCH请求进行缓存失效
		if c.Request.Method == "GET" {
			c.Next()
			return
		}

		c.Next()

		// 请求处理完成后，异步清理相关缓存
		go func() {
			ctx := context.Background()
			for _, tag := range tags {
				redisService.DeleteCacheByTag(ctx, tag)
			}
		}()
	}
}

// responseWriter 自定义响应写入器，用于捕获响应数据
type responseWriter struct {
	gin.ResponseWriter
	body       []byte
	statusCode int
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.body = append(w.body, data...)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// generateCacheKey 生成缓存键
func generateCacheKey(prefix, path, query string) string {
	// 将路径和查询参数组合
	fullPath := path
	if query != "" {
		fullPath += "?" + query
	}

	// 使用MD5生成唯一的缓存键
	hash := md5.Sum([]byte(fullPath))
	key := fmt.Sprintf("%s:%x", prefix, hash)
	return key
}

// CacheConfig 缓存配置
type CacheConfig struct {
	KeyPrefix  string
	Expiration time.Duration
	Tags       []string
	Patterns   []string
}

// GetCacheConfig 获取缓存配置
func GetCacheConfig(endpoint string) *CacheConfig {
	configs := map[string]*CacheConfig{
		// 热门项目
		"/api/v1/projects/hot": {
			KeyPrefix:  "cache:hot_projects",
			Expiration: 10 * time.Minute,
			Tags:       []string{"projects", "hot"},
			Patterns:   []string{"cache:hot_projects:*", "cache:projects:*"},
		},
		// 热门话题
		"/api/v1/topics/hot": {
			KeyPrefix:  "cache:hot_topics",
			Expiration: 10 * time.Minute,
			Tags:       []string{"topics", "hot"},
			Patterns:   []string{"cache:hot_topics:*", "cache:topics:*"},
		},
		// 最新活动
		"/api/v1/activities/recent": {
			KeyPrefix:  "cache:recent_activities",
			Expiration: 5 * time.Minute,
			Tags:       []string{"activities", "recent"},
			Patterns:   []string{"cache:recent_activities:*", "cache:activities:*"},
		},
		// 文档列表
		"/api/v1/documents": {
			KeyPrefix:  "cache:documents",
			Expiration: 15 * time.Minute,
			Tags:       []string{"documents"},
			Patterns:   []string{"cache:documents:*"},
		},
		// 权限矩阵
		"/api/v1/permissions/matrix": {
			KeyPrefix:  "cache:permissions",
			Expiration: 30 * time.Minute,
			Tags:       []string{"permissions"},
			Patterns:   []string{"cache:permissions:*"},
		},
		// 用户资料
		"/api/v1/users/me": {
			KeyPrefix:  "cache:user_profile",
			Expiration: 20 * time.Minute,
			Tags:       []string{"user", "profile"},
			Patterns:   []string{"cache:user_profile:*", "cache:user:*"},
		},
		// 通知
		"/api/v1/users/me/notifications": {
			KeyPrefix:  "cache:notifications",
			Expiration: 2 * time.Minute,
			Tags:       []string{"notifications"},
			Patterns:   []string{"cache:notifications:*"},
		},
		// 话题统计
		"/api/v1/topics/stats": {
			KeyPrefix:  "cache:topic_stats",
			Expiration: 10 * time.Minute,
			Tags:       []string{"topics", "stats"},
			Patterns:   []string{"cache:topic_stats:*", "cache:topics:*"},
		},
		// 个人统计
		"/api/v1/users/me/stats": {
			KeyPrefix:  "cache:user_stats",
			Expiration: 15 * time.Minute,
			Tags:       []string{"user", "stats"},
			Patterns:   []string{"cache:user_stats:*", "cache:user:*"},
		},
	}

	return configs[endpoint]
}
