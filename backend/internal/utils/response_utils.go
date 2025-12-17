package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RespondSuccess 统一成功响应
func RespondSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// RespondCreated 统一创建成功响应
func RespondCreated(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, data)
}

// RespondError 统一错误响应
func RespondError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{"error": message})
}

// RespondBadRequest 400错误响应
func RespondBadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": message})
}

// RespondUnauthorized 401未授权响应
func RespondUnauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, gin.H{"error": message})
	c.Abort()
}

// RespondForbidden 403禁止访问响应
func RespondForbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, gin.H{"error": message})
	c.Abort()
}

// RespondNotFound 404未找到响应
func RespondNotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, gin.H{"error": message})
}

// RespondInternalError 500内部错误响应
func RespondInternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": message})
}

// RespondWithMessageAndData 带消息和数据的成功响应
func RespondWithMessageAndData(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"message": message,
		"data":    data,
	})
}

// RespondWithMessage 仅带消息的成功响应
func RespondWithMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{"message": message})
}
