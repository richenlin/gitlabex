package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gitlabex/internal/config"
	"gitlabex/internal/models"
	"gitlabex/internal/services"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	userService   *services.UserService
	gitlabService *services.GitLabService
	config        *config.Config
	redisService  *services.RedisService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(userService *services.UserService, gitlabService *services.GitLabService, cfg *config.Config, redisService *services.RedisService) *AuthHandler {
	return &AuthHandler{
		userService:   userService,
		gitlabService: gitlabService,
		config:        cfg,
		redisService:  redisService,
	}
}

// GitLabOAuthResponse OAuth响应结构
type GitLabOAuthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// JWTClaims JWT声明
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GitLabAuth GitLab OAuth认证
func (h *AuthHandler) GitLabAuth(c *gin.Context) {
	// 生成随机state参数防止CSRF攻击
	state, err := generateRandomState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成状态参数失败"})
		return
	}

	// 将state存储到Redis中，10分钟有效期
	if err := h.redisService.SetOAuthState(state, 10*time.Minute); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "存储状态参数失败"})
		return
	}

	// 构建GitLab OAuth授权URL
	// 将scope中的空格替换为加号，符合URL编码规范
	scopes := strings.ReplaceAll(h.config.GitLabScopes, " ", "+")
	authURL := fmt.Sprintf("%s/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		h.config.GitLabURL,
		h.config.GitLabClientID,
		url.QueryEscape(h.config.GitLabRedirectURI),
		scopes,
		state)

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"state":    state,
	})
}

// GitLabCallback GitLab OAuth回调
func (h *AuthHandler) GitLabCallback(c *gin.Context) {
	// 获取授权码和状态参数
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少授权码"})
		return
	}

	// 验证state参数
	valid, err := h.redisService.ValidateOAuthState(state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "验证状态参数失败"})
		return
	}
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的状态参数"})
		return
	}

	// 交换授权码获取访问令牌
	oauthResp, err := h.exchangeCodeForToken(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取访问令牌失败: " + err.Error()})
		return
	}

	// 使用访问令牌获取用户信息
	gitlabUser, err := h.gitlabService.GetUser(oauthResp.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败: " + err.Error()})
		return
	}

	// 创建或更新用户
	user, err := h.userService.CreateOrUpdateUserFromGitLab(gitlabUser, oauthResp.AccessToken, oauthResp.RefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败: " + err.Error()})
		return
	}

	// 生成JWT令牌
	jwtToken, err := h.generateJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成JWT令牌失败: " + err.Error()})
		return
	}

	// 更新用户最后登录时间
	now := time.Now()
	user.LastLoginAt = &now
	if err := h.userService.UpdateUser(user.ID, map[string]interface{}{
		"last_login_at": now,
	}); err != nil {
		// 记录错误但不影响登录流程
		fmt.Printf("更新用户登录时间失败: %v\n", err)
	}

	// 返回认证结果
	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"token":   jwtToken,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"name":     user.Name,
			"email":    user.Email,
			"role":     user.Role,
			"edu_role": user.EduRole,
		},
	})
}

// RefreshToken 刷新访问令牌
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// 从请求头获取当前JWT令牌
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少认证令牌"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的令牌格式"})
		return
	}

	// 解析JWT令牌（即使过期也要解析出用户信息）
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.config.JWTSecret), nil
	})

	var claims *JWTClaims
	if err != nil {
		// JWT v5版本的错误处理
		if parsedClaims, ok := token.Claims.(*JWTClaims); ok {
			// 即使令牌过期，我们也可以获取声明来刷新
			claims = parsedClaims
		}
		if claims == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的令牌"})
			return
		}
	} else {
		claims = token.Claims.(*JWTClaims)
	}

	// 获取用户信息
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的用户ID"})
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
		return
	}

	// 使用用户的refresh token刷新GitLab访问令牌
	if user.RefreshToken != "" {
		newOAuthResp, err := h.refreshGitLabToken(user.RefreshToken)
		if err == nil {
			// 更新用户的访问令牌
			user.AccessToken = newOAuthResp.AccessToken
			if newOAuthResp.RefreshToken != "" {
				user.RefreshToken = newOAuthResp.RefreshToken
			}

			// 保存到数据库
			h.userService.UpdateUser(user.ID, map[string]interface{}{
				"access_token":  user.AccessToken,
				"refresh_token": user.RefreshToken,
			})
		}
	}

	// 生成新的JWT令牌
	newJWTToken, err := h.generateJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成新令牌失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "令牌刷新成功",
		"token":   newJWTToken,
	})
}

// Logout 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("userID")
	if exists {
		// 可以在这里添加登出逻辑，比如将令牌加入黑名单
		fmt.Printf("用户 %v 登出\n", userID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "登出成功"})
}

// exchangeCodeForToken 交换授权码获取访问令牌
func (h *AuthHandler) exchangeCodeForToken(code string) (*GitLabOAuthResponse, error) {
	data := url.Values{}
	data.Set("client_id", h.config.GitLabClientID)
	data.Set("client_secret", h.config.GitLabClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", h.config.GitLabRedirectURI)

	req, err := http.NewRequest("POST", h.config.GitLabURL+"/oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 读取错误响应体以获取更多信息
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OAuth token exchange failed with status: %s, response: %s", resp.Status, string(body))
	}

	var oauthResp GitLabOAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&oauthResp); err != nil {
		return nil, err
	}

	return &oauthResp, nil
}

// refreshGitLabToken 刷新GitLab访问令牌
func (h *AuthHandler) refreshGitLabToken(refreshToken string) (*GitLabOAuthResponse, error) {
	data := url.Values{}
	data.Set("client_id", h.config.GitLabClientID)
	data.Set("client_secret", h.config.GitLabClientSecret)
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")

	req, err := http.NewRequest("POST", h.config.GitLabURL+"/oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed with status: %s", resp.Status)
	}

	var oauthResp GitLabOAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&oauthResp); err != nil {
		return nil, err
	}

	return &oauthResp, nil
}

// generateJWT 生成JWT令牌
func (h *AuthHandler) generateJWT(user *models.User) (string, error) {
	now := time.Now()
	claims := &JWTClaims{
		UserID:   user.ID.String(),
		Username: user.Username,
		Role:     string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(h.config.JWTExpirationHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "gitlabex",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.config.JWTSecret))
}

// generateRandomState 生成随机状态参数
func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
