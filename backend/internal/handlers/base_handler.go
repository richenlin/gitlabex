package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// BaseHandler 基础处理器接口
type BaseHandler interface {
	Error(c *gin.Context, code int, err error)
	Success(c *gin.Context, data interface{})
}

// HandlerHelper 处理器辅助函数集合
type HandlerHelper struct{}

// GetGitLabToken 从Context获取GitLab Token
func (h *HandlerHelper) GetGitLabToken(c *gin.Context) (string, error) {
	token, exists := c.Get("gitlab_access_token")
	if !exists {
		return "", fmt.Errorf("用户未登录")
	}
	return token.(string), nil
}

// GetGitLabUserID 从Context获取GitLab User ID
func (h *HandlerHelper) GetGitLabUserID(c *gin.Context) (int64, error) {
	userID, exists := c.Get("gitlab_user_id")
	if !exists {
		return 0, fmt.Errorf("用户未登录")
	}
	return userID.(int64), nil
}

// Error 返回错误响应
func (h *HandlerHelper) Error(c *gin.Context, code int, err error) {
	c.JSON(code, gin.H{"error": err.Error()})
}

// Success 返回成功响应
func (h *HandlerHelper) Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// GetQueryInt 获取整型查询参数
func (h *HandlerHelper) GetQueryInt(c *gin.Context, key string, defaultValue int) int {
	valStr := c.Query(key)
	if valStr == "" {
		return defaultValue
	}
	// 简单的转换逻辑，如果有复杂校验应在外部处理
	// 这里假设简单的数字转换
	var val int
	fmt.Sscanf(valStr, "%d", &val)
	if val == 0 && valStr != "0" {
		return defaultValue
	}
	return val
}

