package dto

import "github.com/golang-jwt/jwt/v5"

// GitLabOAuthResponse OAuth响应结构
type GitLabOAuthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// JWTClaims JWT声明 - 简化结构，只包含GitLab访问令牌
type JWTClaims struct {
	GitLabAccessToken string `json:"gitlab_access_token"`
	jwt.RegisteredClaims
}

// AuthURLResponse 授权URL响应
type AuthURLResponse struct {
	AuthURL string `json:"auth_url"`
	State   string `json:"state"`
}

// LoginSuccessResponse 登录成功响应
type LoginSuccessResponse struct {
	Message string                 `json:"message"`
	Token   string                 `json:"token"`
	User    map[string]interface{} `json:"user"`
}

// TokenRefreshResponse 令牌刷新响应
type TokenRefreshResponse struct {
	Message string `json:"message"`
	Token   string `json:"token"`
}

