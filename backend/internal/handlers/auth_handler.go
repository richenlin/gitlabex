package handlers

import (
	"fmt"
	"gitlabex/internal/config"
	"gitlabex/internal/dto"
	"gitlabex/internal/services"
	"gitlabex/internal/utils"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	HandlerHelper
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

// GitLabAuth GitLab OAuth认证
func (h *AuthHandler) GitLabAuth(c *gin.Context) {
	// 生成随机state参数防止CSRF攻击
	state, err := utils.GenerateRandomState()
	if err != nil {
		h.Error(c, http.StatusInternalServerError, fmt.Errorf("生成状态参数失败"))
		return
	}

	// 将state存储到Redis中，10分钟有效期（如果Redis可用）
	if h.redisService != nil {
		if err := h.redisService.SetOAuthState(state, 10*time.Minute); err != nil {
			h.Error(c, http.StatusInternalServerError, fmt.Errorf("存储状态参数失败"))
			return
		}
	}
	// 注意：如果Redis不可用，我们仍然继续OAuth流程，但state验证将被跳过

	// 构建GitLab OAuth授权URL
	// 将scope中的空格替换为加号，符合URL编码规范
	scopes := strings.ReplaceAll(h.config.GitLab.Scopes, " ", "+")
	authURL := fmt.Sprintf("%s/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		h.config.GitLab.URL,
		h.config.GitLab.ClientID,
		url.QueryEscape(h.config.GitLab.RedirectURI),
		scopes,
		state)

	h.Success(c, dto.AuthURLResponse{
		AuthURL: authURL,
		State:   state,
	})
}

// GitLabCallback GitLab OAuth回调
func (h *AuthHandler) GitLabCallback(c *gin.Context) {
	// 获取授权码和状态参数
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		h.Error(c, http.StatusBadRequest, fmt.Errorf("缺少授权码"))
		return
	}

	// 验证state参数（如果Redis可用）
	if h.redisService != nil {
		valid, err := h.redisService.ValidateOAuthState(state)
		if err != nil {
			h.Error(c, http.StatusInternalServerError, fmt.Errorf("验证状态参数失败"))
			return
		}
		if !valid {
			h.Error(c, http.StatusBadRequest, fmt.Errorf("无效的状态参数"))
			return
		}
	}
	// 注意：如果Redis不可用，我们跳过state验证（降级处理）

	// 交换授权码获取访问令牌
	oauthResp, err := h.gitlabService.ExchangeToken(code)
	if err != nil {
		h.Error(c, http.StatusInternalServerError, fmt.Errorf("获取访问令牌失败: %v", err))
		return
	}

	// 使用访问令牌获取用户信息
	gitlabUser, err := h.gitlabService.GetUser(oauthResp.AccessToken)
	if err != nil {
		h.Error(c, http.StatusInternalServerError, fmt.Errorf("获取用户信息失败: %v", err))
		return
	}

	// 生成JWT令牌 - 只包含GitLab访问令牌
	jwtToken, err := h.generateJWT(oauthResp.AccessToken)
	if err != nil {
		h.Error(c, http.StatusInternalServerError, fmt.Errorf("生成JWT令牌失败: %v", err))
		return
	}

	// 返回认证结果 - 用户信息直接从GitLab获取
	h.Success(c, dto.LoginSuccessResponse{
		Message: "登录成功",
		Token:   jwtToken,
		User: map[string]interface{}{
			"id":         gitlabUser.ID,
			"username":   gitlabUser.Username,
			"name":       gitlabUser.Name,
			"email":      gitlabUser.Email,
			"avatar_url": gitlabUser.Avatar,
		},
	})
}

// RefreshToken 刷新访问令牌
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// 从请求头获取当前JWT令牌
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		h.Error(c, http.StatusUnauthorized, fmt.Errorf("缺少认证令牌"))
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		h.Error(c, http.StatusUnauthorized, fmt.Errorf("无效的令牌格式"))
		return
	}

	// 解析JWT令牌（即使过期也要解析出用户信息）
	token, err := jwt.ParseWithClaims(tokenString, &dto.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.config.JWT.Secret), nil
	})

	var claims *dto.JWTClaims
	if err != nil {
		// JWT v5版本的错误处理
		if parsedClaims, ok := token.Claims.(*dto.JWTClaims); ok {
			// 即使令牌过期，我们也可以获取声明来刷新
			claims = parsedClaims
		}
		if claims == nil {
			h.Error(c, http.StatusUnauthorized, fmt.Errorf("无效的令牌"))
			return
		}
	} else {
		claims = token.Claims.(*dto.JWTClaims)
	}

	// 验证当前GitLab访问令牌是否仍然有效
	_, err = h.gitlabService.GetUser(claims.GitLabAccessToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "gitlab_token_expired",
			"message": "GitLab访问令牌已过期，请重新登录",
		})
		return
	}

	// 生成新的JWT令牌
	newJWTToken, err := h.generateJWT(claims.GitLabAccessToken)
	if err != nil {
		h.Error(c, http.StatusInternalServerError, fmt.Errorf("生成新令牌失败"))
		return
	}

	h.Success(c, dto.TokenRefreshResponse{
		Message: "令牌刷新成功",
		Token:   newJWTToken,
	})
}

// Logout 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	// 从上下文获取GitLab用户ID
	gitlabUserID, exists := c.Get("gitlab_user_id")
	if exists {
		// 可以在这里添加登出逻辑，比如将令牌加入黑名单
		fmt.Printf("GitLab用户 %v 登出\n", gitlabUserID)
	}

	h.Success(c, gin.H{"message": "登出成功"})
}

// generateJWT 生成JWT令牌 - 简化结构，只包含GitLab访问令牌
func (h *AuthHandler) generateJWT(gitlabAccessToken string) (string, error) {
	now := time.Now()
	claims := &dto.JWTClaims{
		GitLabAccessToken: gitlabAccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(h.config.JWT.ExpirationHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "gitlabex",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.config.JWT.Secret))
}
