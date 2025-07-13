package services

import (
	"encoding/json"
	"fmt"
	"net/http"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code is required"})
		return
	}

	// 交换访问令牌
	token, err := s.exchangeCodeForToken(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code for token"})
		return
	}

	// 获取用户信息
	gitlabUser, err := s.fetchGitLabUser(token.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user info"})
		return
	}

	// 同步用户到本地数据库
	user, err := s.syncUserFromGitLab(gitlabUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync user"})
		return
	}

	// 生成JWT令牌
	jwtToken, err := s.generateJWTToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT token"})
		return
	}

	// 返回认证结果
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   jwtToken,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"name":     user.Name,
			"avatar":   user.Avatar,
			"role":     user.Role,
		},
	})
}

// exchangeCodeForToken 交换认证码为访问令牌
func (s *AuthService) exchangeCodeForToken(code string) (*GitLabToken, error) {
	// 使用外部URL进行token交换，确保OAuth认证正常工作
	tokenURL := fmt.Sprintf("%s/oauth/token", s.config.GitLab.URL)

	fmt.Printf("DEBUG: Exchange token URL: %s\n", tokenURL)
	fmt.Printf("DEBUG: Client ID: %s\n", s.config.GitLab.ClientID[:10]+"...")
	fmt.Printf("DEBUG: Redirect URI: %s\n", s.config.GitLab.RedirectURI)

	formData := map[string][]string{
		"client_id":     {s.config.GitLab.ClientID},
		"client_secret": {s.config.GitLab.ClientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {s.config.GitLab.RedirectURI},
	}

	resp, err := http.PostForm(tokenURL, formData)
	if err != nil {
		fmt.Printf("DEBUG: HTTP request error: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Printf("DEBUG: Response status: %s\n", resp.Status)
	fmt.Printf("DEBUG: Response headers: %v\n", resp.Header)

	if resp.StatusCode != http.StatusOK {
		// 读取错误响应内容
		body, _ := json.Marshal(resp.Body)
		fmt.Printf("DEBUG: Error response body: %s\n", string(body))
		return nil, fmt.Errorf("failed to exchange token: %s", resp.Status)
	}

	var token GitLabToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		fmt.Printf("DEBUG: JSON decode error: %v\n", err)
		return nil, err
	}

	fmt.Printf("DEBUG: Token exchange successful, access_token length: %d\n", len(token.AccessToken))

	return &token, nil
}

// fetchGitLabUser 从GitLab API获取用户信息
func (s *AuthService) fetchGitLabUser(accessToken string) (*GitLabUser, error) {
	// 使用外部URL获取用户信息，与token交换保持一致
	userURL := fmt.Sprintf("%s/api/v4/user", s.config.GitLab.URL)

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

		// 设置用户信息到上下文
		c.Set("user_id", claims.UserID)
		c.Set("gitlab_id", claims.GitLabID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("name", claims.Name)
		c.Set("role", claims.Role)

		c.Next()
	}
}
