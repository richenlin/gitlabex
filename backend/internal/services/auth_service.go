package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gitlabex/internal/config"
	"gitlabex/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type AuthService struct {
	db     *gorm.DB
	config *config.Config
}

type GitLabOAuthConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`
	GitLabURL    string `json:"gitlab_url"`
}

type GitLabUser struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	WebURL    string `json:"web_url"`
	State     string `json:"state"`
}

type GitLabToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	CreatedAt    int64  `json:"created_at"`
	Scope        string `json:"scope"`
}

type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	GitLabID int    `json:"gitlab_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     int    `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(db *gorm.DB, config *config.Config) *AuthService {
	return &AuthService{
		db:     db,
		config: config,
	}
}

// GetGitLabOAuthURL 获取GitLab OAuth认证URL
func (s *AuthService) GetGitLabOAuthURL(state string) string {
	// 将空格分隔的scopes转换为+分隔的格式
	scopes := strings.ReplaceAll(s.config.GitLab.Scopes, " ", "+")
	return fmt.Sprintf("%s/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		s.config.GitLab.URL,
		s.config.GitLab.ClientID,
		s.config.GitLab.RedirectURI,
		scopes,
		state,
	)
}

// HandleGitLabCallback 处理GitLab OAuth回调
func (s *AuthService) HandleGitLabCallback(c *gin.Context) {
	var code, state string

	// 支持GET和POST两种方式
	if c.Request.Method == "GET" {
		code = c.Query("code")
		state = c.Query("state")
	} else if c.Request.Method == "POST" {
		// 从JSON body中获取code和state
		var requestBody struct {
			Code  string `json:"code"`
			State string `json:"state"`
		}
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		code = requestBody.Code
		state = requestBody.State
	}

	// TODO: 验证state参数以防止CSRF攻击
	_ = state

	if code == "" {
		// 重定向到前端登录页面并显示错误
		c.Redirect(http.StatusFound, fmt.Sprintf("%s/login?error=missing_code", s.config.Frontend.URL))
		return
	}

	// 交换访问令牌
	token, err := s.exchangeCodeForToken(code)
	if err != nil {
		fmt.Printf("ERROR: Failed to exchange token: %v\n", err)
		// 重定向到前端登录页面并显示错误
		c.Redirect(http.StatusFound, fmt.Sprintf("%s/login?error=token_exchange_failed", s.config.Frontend.URL))
		return
	}

	// 获取用户信息
	gitlabUser, err := s.fetchGitLabUser(token.AccessToken)
	if err != nil {
		fmt.Printf("ERROR: Failed to fetch user: %v\n", err)
		// 重定向到前端登录页面并显示错误
		c.Redirect(http.StatusFound, fmt.Sprintf("%s/login?error=fetch_user_failed", s.config.Frontend.URL))
		return
	}

	// 同步用户到本地数据库
	user, err := s.syncUserFromGitLab(gitlabUser)
	if err != nil {
		fmt.Printf("ERROR: Failed to sync user: %v\n", err)
		// 重定向到前端登录页面并显示错误
		c.Redirect(http.StatusFound, fmt.Sprintf("%s/login?error=sync_user_failed", s.config.Frontend.URL))
		return
	}

	// 生成JWT令牌
	jwtToken, err := s.generateJWTToken(user)
	if err != nil {
		fmt.Printf("ERROR: Failed to generate JWT: %v\n", err)
		// 重定向到前端登录页面并显示错误
		c.Redirect(http.StatusFound, fmt.Sprintf("%s/login?error=jwt_generation_failed", s.config.Frontend.URL))
		return
	}

	fmt.Printf("SUCCESS: User %s logged in successfully\n", user.Username)

	// 登录成功，重定向到前端登录成功页面并传递token
	// 使用URL参数传递token（在生产环境中应该使用更安全的方式，如设置HttpOnly cookie）
	redirectURL := fmt.Sprintf("%s/login/success?token=%s", s.config.Frontend.URL, jwtToken)
	c.Redirect(http.StatusFound, redirectURL)
}

// exchangeCodeForToken 交换认证码为访问令牌
func (s *AuthService) exchangeCodeForToken(code string) (*GitLabToken, error) {
	// 使用内部URL进行token交换，确保容器内部网络访问正常
	var tokenURL string
	if s.config.GitLab.InternalURL != "" {
		tokenURL = fmt.Sprintf("%s/oauth/token", s.config.GitLab.InternalURL)
	} else {
		tokenURL = fmt.Sprintf("%s/oauth/token", s.config.GitLab.URL)
	}

	fmt.Printf("DEBUG: GitLab InternalURL: %s\n", s.config.GitLab.InternalURL)
	fmt.Printf("DEBUG: GitLab URL: %s\n", s.config.GitLab.URL)
	fmt.Printf("DEBUG: Exchange token URL: %s\n", tokenURL)
	fmt.Printf("DEBUG: Client ID: %s\n", s.config.GitLab.ClientID[:10]+"...")
	fmt.Printf("DEBUG: Client Secret: %s\n", s.config.GitLab.ClientSecret[:10]+"...")
	fmt.Printf("DEBUG: Redirect URI: %s\n", s.config.GitLab.RedirectURI)
	fmt.Printf("DEBUG: Authorization Code: %s\n", code[:10]+"...")

	// 使用HTTP Basic认证发送client credentials
	formData := map[string][]string{
		"code":         {code},
		"grant_type":   {"authorization_code"},
		"redirect_uri": {s.config.GitLab.RedirectURI},
	}

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(url.Values(formData).Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置Basic认证头
	req.SetBasicAuth(s.config.GitLab.ClientID, s.config.GitLab.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	fmt.Printf("DEBUG: Request headers: %v\n", req.Header)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("DEBUG: HTTP request error: %v\n", err)
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("DEBUG: Response status: %s\n", resp.Status)
	fmt.Printf("DEBUG: Response headers: %v\n", resp.Header)

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("DEBUG: Failed to read response body: %v\n", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	fmt.Printf("DEBUG: Response body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab OAuth error (status %d): %s", resp.StatusCode, string(body))
	}

	var token GitLabToken
	if err := json.Unmarshal(body, &token); err != nil {
		fmt.Printf("DEBUG: JSON decode error: %v\n", err)
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	fmt.Printf("DEBUG: Token exchange successful, access_token length: %d\n", len(token.AccessToken))

	return &token, nil
}

// fetchGitLabUser 从GitLab API获取用户信息
func (s *AuthService) fetchGitLabUser(accessToken string) (*GitLabUser, error) {
	// 使用内部URL获取用户信息，确保容器内部网络访问正常
	var userURL string
	if s.config.GitLab.InternalURL != "" {
		userURL = fmt.Sprintf("%s/api/v4/user", s.config.GitLab.InternalURL)
	} else {
		userURL = fmt.Sprintf("%s/api/v4/user", s.config.GitLab.URL)
	}

	req, err := http.NewRequest("GET", userURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch user: %s", resp.Status)
	}

	var gitlabUser GitLabUser
	if err := json.NewDecoder(resp.Body).Decode(&gitlabUser); err != nil {
		return nil, err
	}

	return &gitlabUser, nil
}

// syncUserFromGitLab 同步GitLab用户到本地数据库
func (s *AuthService) syncUserFromGitLab(gitlabUser *GitLabUser) (*models.User, error) {
	var user models.User

	// 查找现有用户
	err := s.db.Where("gitlab_id = ?", gitlabUser.ID).First(&user).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// 更新用户信息
	user.GitLabID = gitlabUser.ID
	user.Username = gitlabUser.Username
	user.Email = gitlabUser.Email
	user.Name = gitlabUser.Name
	user.Avatar = gitlabUser.AvatarURL
	user.LastSyncAt = time.Now()
	user.Active = true

	// 如果是新用户，设置默认角色
	if user.ID == 0 {
		user.Role = 3 // 默认学生角色
	}

	// 保存到数据库
	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// generateJWTToken 生成JWT令牌
func (s *AuthService) generateJWTToken(user *models.User) (string, error) {
	claims := JWTClaims{
		UserID:   user.ID,
		GitLabID: user.GitLabID,
		Username: user.Username,
		Email:    user.Email,
		Name:     user.Name,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "gitlabex",
			Subject:   fmt.Sprintf("user_%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWT.Secret))
}

// ValidateJWTToken 验证JWT令牌
func (s *AuthService) ValidateJWTToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GetUserByID 根据ID获取用户
func (s *AuthService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// RefreshUserFromGitLab 从GitLab刷新用户信息
func (s *AuthService) RefreshUserFromGitLab(userID uint, accessToken string) (*models.User, error) {
	// 获取GitLab用户信息
	gitlabUser, err := s.fetchGitLabUser(accessToken)
	if err != nil {
		return nil, err
	}

	// 同步到本地数据库
	return s.syncUserFromGitLab(gitlabUser)
}

// Logout 用户登出
func (s *AuthService) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

// GetCurrentUser 获取当前用户
func (s *AuthService) GetCurrentUser(c *gin.Context) {
	// 从上下文获取用户信息（通过中间件设置）
	userID, exists := c.Get("user_id")
	if !exists {
		// 如果没有认证中间件，返回测试用户
		testUser := &models.User{
			ID:         1,
			GitLabID:   1,
			Username:   "testuser",
			Email:      "test@example.com",
			Name:       "Test User",
			Avatar:     "https://www.gravatar.com/avatar/default",
			Role:       2, // 学生角色
			LastSyncAt: time.Now(),
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Success",
			"data": gin.H{
				"id":           testUser.ID,
				"gitlab_id":    testUser.GitLabID,
				"username":     testUser.Username,
				"email":        testUser.Email,
				"name":         testUser.Name,
				"avatar":       testUser.Avatar,
				"role":         testUser.Role,
				"last_sync_at": testUser.LastSyncAt,
				"is_active":    true,
			},
		})
		return
	}

	user, err := s.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data": gin.H{
			"id":           user.ID,
			"gitlab_id":    user.GitLabID,
			"username":     user.Username,
			"email":        user.Email,
			"name":         user.Name,
			"avatar":       user.Avatar,
			"role":         user.Role,
			"last_sync_at": user.LastSyncAt,
			"is_active":    user.Active,
		},
	})
}

// AuthMiddleware JWT认证中间件
func (s *AuthService) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// 移除 "Bearer " 前缀
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		claims, err := s.ValidateJWTToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// 从数据库获取完整的用户信息
		user, err := s.GetUserByID(claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		c.Set("current_user", user)
		c.Set("user_id", claims.UserID)
		c.Set("gitlab_id", claims.GitLabID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("name", claims.Name)
		c.Set("role", claims.Role)

		c.Next()
	}
}
