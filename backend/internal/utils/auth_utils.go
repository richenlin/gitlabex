package utils

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// AuthInfo 认证信息结构
type AuthInfo struct {
	UserID      int64
	Username    string
	Email       string
	Name        string
	AccessToken string
	IsAdmin     bool
	IsGuest     bool
}

// GetAuthInfo 从gin.Context中安全获取认证信息
func GetAuthInfo(c *gin.Context) (*AuthInfo, error) {
	// 检查是否为游客
	isGuest, _ := GetBoolFromContext(c, "is_guest")
	if isGuest {
		return nil, fmt.Errorf("用户未登录")
	}

	// 获取用户ID
	userID, exists := GetInt64FromContext(c, "gitlab_user_id")
	if !exists {
		return nil, fmt.Errorf("无效的用户ID")
	}

	// 获取访问令牌
	accessToken, exists := GetStringFromContext(c, "gitlab_access_token")
	if !exists {
		return nil, fmt.Errorf("缺少访问令牌")
	}

	// 获取其他信息（可选）
	username, _ := GetStringFromContext(c, "username")
	email, _ := GetStringFromContext(c, "email")
	name, _ := GetStringFromContext(c, "name")
	isAdmin := GetBoolFromContext(c, "is_admin")

	return &AuthInfo{
		UserID:      userID,
		Username:    username,
		Email:       email,
		Name:        name,
		AccessToken: accessToken,
		IsAdmin:     isAdmin,
		IsGuest:     false,
	}, nil
}

// MustGetAuthInfo 获取认证信息，如果失败则中断请求
func MustGetAuthInfo(c *gin.Context) (*AuthInfo, bool) {
	authInfo, err := GetAuthInfo(c)
	if err != nil {
		RespondUnauthorized(c, err.Error())
		return nil, false
	}
	return authInfo, true
}

// GetInt64FromContext 从gin.Context中安全获取int64值
func GetInt64FromContext(c *gin.Context, key string) (int64, bool) {
	value, exists := c.Get(key)
	if !exists {
		return 0, false
	}

	int64Value, ok := value.(int64)
	if !ok {
		return 0, false
	}

	return int64Value, true
}

// GetStringFromContext 从gin.Context中安全获取string值
func GetStringFromContext(c *gin.Context, key string) (string, bool) {
	value, exists := c.Get(key)
	if !exists {
		return "", false
	}

	strValue, ok := value.(string)
	if !ok {
		return "", false
	}

	return strValue, true
}

// GetBoolFromContext 从gin.Context中安全获取bool值
func GetBoolFromContext(c *gin.Context, key string) bool {
	value, exists := c.Get(key)
	if !exists {
		return false
	}

	boolValue, ok := value.(bool)
	if !ok {
		return false
	}

	return boolValue
}

// GetIntFromContext 从gin.Context中安全获取int值
func GetIntFromContext(c *gin.Context, key string) (int, bool) {
	value, exists := c.Get(key)
	if !exists {
		return 0, false
	}

	intValue, ok := value.(int)
	if !ok {
		return 0, false
	}

	return intValue, true
}
