package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"gitlabex/internal/middleware"
	"gitlabex/internal/models"
	"gitlabex/internal/services"
)

// SyncHandler 第三方系统同步处理器
type SyncHandler struct {
	userService         *services.UserService
	gitlabService       *services.GitLabService
	externalUserService *services.ExternalUserService
	jwtSecret           string
}

// NewSyncHandler 创建同步处理器
func NewSyncHandler(userService *services.UserService, gitlabService *services.GitLabService, externalUserService *services.ExternalUserService, jwtSecret string) *SyncHandler {
	return &SyncHandler{
		userService:         userService,
		gitlabService:       gitlabService,
		externalUserService: externalUserService,
		jwtSecret:           jwtSecret,
	}
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Role     string `json:"role" binding:"required"`

	// 扩展信息
	AvatarURL  string `json:"avatar_url,omitempty"`
	Department string `json:"department,omitempty"`
	StudentID  string `json:"student_id,omitempty"`
	TeacherID  string `json:"teacher_id,omitempty"`
	Phone      string `json:"phone,omitempty"`

	// 第三方系统标识（必需）
	ExternalID     string `json:"external_id" binding:"required"`
	ExternalSource string `json:"external_source" binding:"required"`
}

// CreateUserResponse 创建用户响应
type CreateUserResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Data    *UserTokenData `json:"data,omitempty"`
	Error   string         `json:"error,omitempty"`
}

// UserTokenData 用户和Token数据
type UserTokenData struct {
	User         *UserInfo         `json:"user"`
	Token        string            `json:"token"`
	ExternalUser *ExternalUserInfo `json:"external_user,omitempty"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	AvatarURL string    `json:"avatar_url"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// ExternalUserInfo 外部用户信息
type ExternalUserInfo struct {
	ID             uint   `json:"id"`
	ExternalID     string `json:"external_id"`
	ExternalSource string `json:"external_source"`
	Department     string `json:"department,omitempty"`
	StudentID      string `json:"student_id,omitempty"`
	TeacherID      string `json:"teacher_id,omitempty"`
	Phone          string `json:"phone,omitempty"`
}

// CreateUser 创建用户接口
// @Summary 第三方系统创建用户
// @Description 为第三方系统提供用户创建和同步接口，创建成功后返回登录Token
// @Tags 同步接口
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "用户创建请求"
// @Param X-API-Key header string true "API密钥"
// @Success 201 {object} CreateUserResponse "用户创建成功"
// @Failure 400 {object} CreateUserResponse "请求参数错误"
// @Failure 401 {object} CreateUserResponse "API密钥无效"
// @Failure 403 {object} CreateUserResponse "权限不足"
// @Failure 409 {object} CreateUserResponse "用户已存在"
// @Failure 500 {object} CreateUserResponse "服务器内部错误"
// @Router /api/v1/sync/users [post]
func (h *SyncHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, CreateUserResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 获取API密钥信息
	keyInfo, exists := middleware.GetAPIKeyInfo(c)
	if !exists {
		c.JSON(http.StatusInternalServerError, CreateUserResponse{
			Success: false,
			Message: "服务器内部错误",
			Error:   "无法获取API密钥信息",
		})
		return
	}

	// 检查权限：是否可以创建管理员
	if req.Role == "admin" && !keyInfo.CanAdmin {
		c.JSON(http.StatusForbidden, CreateUserResponse{
			Success: false,
			Message: "权限不足",
			Error:   "当前API密钥无权创建管理员用户",
		})
		return
	}

	// 获取系统管理员Token进行GitLab操作
	adminToken, err := h.getSystemAdminToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, CreateUserResponse{
			Success: false,
			Message: "服务器内部错误",
			Error:   "无法获取系统管理员权限",
		})
		return
	}

	// 准备同步数据
	syncData := &services.ExternalUserSyncData{
		ExternalID:     req.ExternalID,
		ExternalSource: req.ExternalSource,
		Username:       req.Username,
		Password:       req.Password,
		Email:          req.Email,
		Name:           req.Name,
		Role:           req.Role,
		Department:     req.Department,
		StudentID:      req.StudentID,
		TeacherID:      req.TeacherID,
		Phone:          req.Phone,
	}

	// 同步用户到GitLab
	mapping, err := h.externalUserService.SyncExternalUser(
		adminToken,
		syncData,
		string(keyInfo.Type),
		c.ClientIP(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, CreateUserResponse{
			Success: false,
			Message: "用户同步失败",
			Error:   err.Error(),
		})
		return
	}

	// 生成JWT Token
	token, err := h.generateJWTToken(mapping.GitLabUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, CreateUserResponse{
			Success: false,
			Message: "Token生成失败",
			Error:   err.Error(),
		})
		return
	}

	// 构建响应数据
	userInfo := &UserInfo{
		ID:        mapping.GitLabUser.ID,
		Username:  mapping.GitLabUser.Username,
		Email:     mapping.GitLabUser.Email,
		Name:      mapping.GitLabUser.Name,
		Role:      mapping.GitLabUser.Role.GetEducationRole(),
		AvatarURL: mapping.GitLabUser.AvatarURL,
		IsActive:  true,
		CreatedAt: mapping.ExternalUser.CreatedAt,
	}

	externalUserInfo := &ExternalUserInfo{
		ID:             mapping.ExternalUser.ID,
		ExternalID:     mapping.ExternalUser.ExternalID,
		ExternalSource: mapping.ExternalUser.ExternalSource,
		Department:     mapping.ExternalUser.Department,
		StudentID:      mapping.ExternalUser.StudentID,
		TeacherID:      mapping.ExternalUser.TeacherID,
		Phone:          mapping.ExternalUser.Phone,
	}

	c.JSON(http.StatusCreated, CreateUserResponse{
		Success: true,
		Message: "用户创建成功",
		Data: &UserTokenData{
			User:         userInfo,
			Token:        token,
			ExternalUser: externalUserInfo,
		},
	})
}

// BatchCreateUsersRequest 批量创建用户请求
type BatchCreateUsersRequest struct {
	Users []CreateUserRequest `json:"users" binding:"required,min=1"`
}

// BatchCreateUsersResponse 批量创建用户响应
type BatchCreateUsersResponse struct {
	Success      bool                 `json:"success"`
	Message      string               `json:"message"`
	TotalCount   int                  `json:"total_count"`
	SuccessCount int                  `json:"success_count"`
	FailureCount int                  `json:"failure_count"`
	Results      []CreateUserResponse `json:"results"`
}

// BatchCreateUsers 批量创建用户接口
// @Summary 批量创建用户
// @Description 批量创建多个用户，返回每个用户的创建结果和Token
// @Tags 同步接口
// @Accept json
// @Produce json
// @Param request body BatchCreateUsersRequest true "批量用户创建请求"
// @Param X-API-Key header string true "API密钥"
// @Success 200 {object} BatchCreateUsersResponse "批量创建结果"
// @Failure 400 {object} BatchCreateUsersResponse "请求参数错误"
// @Failure 401 {object} BatchCreateUsersResponse "API密钥无效"
// @Failure 403 {object} BatchCreateUsersResponse "权限不足"
// @Router /api/v1/sync/users/batch [post]
func (h *SyncHandler) BatchCreateUsers(c *gin.Context) {
	var req BatchCreateUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, BatchCreateUsersResponse{
			Success: false,
			Message: "请求参数错误",
		})
		return
	}

	// 获取API密钥信息
	keyInfo, exists := middleware.GetAPIKeyInfo(c)
	if !exists {
		c.JSON(http.StatusInternalServerError, BatchCreateUsersResponse{
			Success: false,
			Message: "服务器内部错误",
		})
		return
	}

	// 检查批量限制
	if len(req.Users) > keyInfo.MaxBatch {
		c.JSON(http.StatusBadRequest, BatchCreateUsersResponse{
			Success: false,
			Message: fmt.Sprintf("批量创建数量超出限制，当前API密钥最多支持 %d 个用户", keyInfo.MaxBatch),
		})
		return
	}

	// 检查是否包含管理员角色
	if !keyInfo.CanAdmin {
		for _, user := range req.Users {
			if user.Role == "admin" {
				c.JSON(http.StatusForbidden, BatchCreateUsersResponse{
					Success: false,
					Message: "权限不足：无权创建管理员用户",
				})
				return
			}
		}
	}

	// 获取系统管理员Token
	adminToken, err := h.getSystemAdminToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, BatchCreateUsersResponse{
			Success: false,
			Message: "服务器内部错误",
		})
		return
	}

	// 准备批量同步数据
	syncDataList := make([]services.ExternalUserSyncData, len(req.Users))
	for i, userReq := range req.Users {
		syncDataList[i] = services.ExternalUserSyncData{
			ExternalID:     userReq.ExternalID,
			ExternalSource: userReq.ExternalSource,
			Username:       userReq.Username,
			Password:       userReq.Password,
			Email:          userReq.Email,
			Name:           userReq.Name,
			Role:           userReq.Role,
			Department:     userReq.Department,
			StudentID:      userReq.StudentID,
			TeacherID:      userReq.TeacherID,
			Phone:          userReq.Phone,
		}
	}

	// 执行批量同步
	batchResult, err := h.externalUserService.BatchSyncExternalUsers(
		adminToken,
		syncDataList,
		string(keyInfo.Type),
		c.ClientIP(),
		keyInfo.MaxBatch,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, BatchCreateUsersResponse{
			Success: false,
			Message: "批量同步失败",
		})
		return
	}

	// 构建响应结果
	results := make([]CreateUserResponse, len(batchResult.Results))
	for i, result := range batchResult.Results {
		if result.Success {
			// 为成功的用户生成Token
			gitlabUser := &models.GitLabUser{
				ID:       result.GitLabUserID,
				Username: result.Username,
				Email:    result.Email,
			}

			token, tokenErr := h.generateJWTToken(gitlabUser)
			if tokenErr != nil {
				results[i] = CreateUserResponse{
					Success: false,
					Message: "Token生成失败",
					Error:   tokenErr.Error(),
				}
			} else {
				results[i] = CreateUserResponse{
					Success: true,
					Message: "用户创建成功",
					Data: &UserTokenData{
						User: &UserInfo{
							ID:       result.GitLabUserID,
							Username: result.Username,
							Email:    result.Email,
							IsActive: true,
						},
						Token: token,
					},
				}
			}
		} else {
			results[i] = CreateUserResponse{
				Success: false,
				Message: "用户创建失败",
				Error:   result.ErrorMessage,
			}
		}
	}

	c.JSON(http.StatusOK, BatchCreateUsersResponse{
		Success:      batchResult.FailureCount == 0,
		Message:      "批量创建完成",
		TotalCount:   batchResult.TotalCount,
		SuccessCount: batchResult.SuccessCount,
		FailureCount: batchResult.FailureCount,
		Results:      results,
	})
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Username   string `json:"username,omitempty"`
	Email      string `json:"email,omitempty"`
	Name       string `json:"name,omitempty"`
	Role       string `json:"role,omitempty"`
	IsActive   *bool  `json:"is_active,omitempty"`
	AvatarURL  string `json:"avatar_url,omitempty"`
	Department string `json:"department,omitempty"`
	StudentID  string `json:"student_id,omitempty"`
	TeacherID  string `json:"teacher_id,omitempty"`
	Phone      string `json:"phone,omitempty"`
}

// UpdateUser 更新用户接口
// @Summary 更新用户信息
// @Description 根据外部系统ID更新用户信息
// @Tags 同步接口
// @Accept json
// @Produce json
// @Param external_id path string true "外部系统用户ID"
// @Param external_source query string true "外部系统来源"
// @Param request body UpdateUserRequest true "用户更新请求"
// @Param X-API-Key header string true "API密钥"
// @Success 200 {object} CreateUserResponse "更新成功"
// @Failure 400 {object} CreateUserResponse "请求参数错误"
// @Failure 401 {object} CreateUserResponse "API密钥无效"
// @Failure 404 {object} CreateUserResponse "用户不存在"
// @Router /api/v1/sync/users/{external_id} [put]
func (h *SyncHandler) UpdateUser(c *gin.Context) {
	externalID := c.Param("external_id")
	externalSource := c.Query("external_source")

	if externalID == "" || externalSource == "" {
		c.JSON(http.StatusBadRequest, CreateUserResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   "external_id和external_source参数都是必需的",
		})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, CreateUserResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 获取API密钥信息
	keyInfo, exists := middleware.GetAPIKeyInfo(c)
	if !exists {
		c.JSON(http.StatusInternalServerError, CreateUserResponse{
			Success: false,
			Message: "服务器内部错误",
		})
		return
	}

	// 检查权限：是否可以设置管理员角色
	if req.Role == "admin" && !keyInfo.CanAdmin {
		c.JSON(http.StatusForbidden, CreateUserResponse{
			Success: false,
			Message: "权限不足",
			Error:   "当前API密钥无权设置管理员角色",
		})
		return
	}

	// 查找外部用户映射
	externalUser, err := h.externalUserService.GetExternalUserByID(externalID, externalSource)
	if err != nil {
		c.JSON(http.StatusInternalServerError, CreateUserResponse{
			Success: false,
			Message: "查询用户失败",
			Error:   err.Error(),
		})
		return
	}

	if externalUser == nil {
		c.JSON(http.StatusNotFound, CreateUserResponse{
			Success: false,
			Message: "用户不存在",
			Error:   "未找到指定的外部用户",
		})
		return
	}

	// 更新外部用户信息
	if req.Username != "" {
		externalUser.Username = req.Username
	}
	if req.Email != "" {
		externalUser.Email = req.Email
	}
	if req.Name != "" {
		externalUser.Name = req.Name
	}
	if req.Role != "" {
		externalUser.Role = req.Role
	}
	if req.Department != "" {
		externalUser.Department = req.Department
	}
	if req.StudentID != "" {
		externalUser.StudentID = req.StudentID
	}
	if req.TeacherID != "" {
		externalUser.TeacherID = req.TeacherID
	}
	if req.Phone != "" {
		externalUser.Phone = req.Phone
	}

	// 更新外部用户映射
	err = h.externalUserService.UpdateExternalUserMapping(externalUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, CreateUserResponse{
			Success: false,
			Message: "更新用户失败",
			Error:   err.Error(),
		})
		return
	}

	// 获取系统管理员Token更新GitLab用户
	adminToken, err := h.getSystemAdminToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, CreateUserResponse{
			Success: false,
			Message: "服务器内部错误",
		})
		return
	}

	// 更新GitLab用户信息
	gitlabUpdateData := &services.GitLabUpdateUserData{
		Username: req.Username,
		Name:     req.Name,
		Email:    req.Email,
	}

	gitlabUser, err := h.gitlabService.UpdateUser(adminToken, externalUser.GitLabUserID, gitlabUpdateData)
	if err != nil {
		// GitLab更新失败，但外部用户映射已更新，记录警告
		c.JSON(http.StatusOK, CreateUserResponse{
			Success: true,
			Message: "用户信息已更新（GitLab同步可能存在延迟）",
		})
		return
	}

	// 构建响应数据
	userInfo := &UserInfo{
		ID:        gitlabUser.ID,
		Username:  gitlabUser.Username,
		Email:     gitlabUser.Email,
		Name:      gitlabUser.Name,
		Role:      models.ConvertExternalRoleToGitLab(externalUser.Role).GetEducationRole(),
		AvatarURL: gitlabUser.Avatar,
		IsActive:  true,
	}

	c.JSON(http.StatusOK, CreateUserResponse{
		Success: true,
		Message: "用户更新成功",
		Data: &UserTokenData{
			User: userInfo,
		},
	})
}

// GetUser 获取用户信息接口
// @Summary 获取用户信息
// @Description 根据外部系统ID获取用户信息
// @Tags 同步接口
// @Produce json
// @Param external_id path string true "外部系统用户ID"
// @Param external_source query string true "外部系统来源"
// @Param X-API-Key header string true "API密钥"
// @Success 200 {object} CreateUserResponse "获取成功"
// @Failure 400 {object} CreateUserResponse "请求参数错误"
// @Failure 401 {object} CreateUserResponse "API密钥无效"
// @Failure 404 {object} CreateUserResponse "用户不存在"
// @Router /api/v1/sync/users/{external_id} [get]
func (h *SyncHandler) GetUser(c *gin.Context) {
	externalID := c.Param("external_id")
	externalSource := c.Query("external_source")

	if externalID == "" || externalSource == "" {
		c.JSON(http.StatusBadRequest, CreateUserResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   "external_id和external_source参数都是必需的",
		})
		return
	}

	// 查找外部用户映射
	externalUser, err := h.externalUserService.GetExternalUserByID(externalID, externalSource)
	if err != nil {
		c.JSON(http.StatusInternalServerError, CreateUserResponse{
			Success: false,
			Message: "查询用户失败",
			Error:   err.Error(),
		})
		return
	}

	if externalUser == nil {
		c.JSON(http.StatusNotFound, CreateUserResponse{
			Success: false,
			Message: "用户不存在",
		})
		return
	}

	// 获取GitLab用户信息
	adminToken, err := h.getSystemAdminToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, CreateUserResponse{
			Success: false,
			Message: "服务器内部错误",
		})
		return
	}

	gitlabUser, err := h.gitlabService.GetUserByID(adminToken, externalUser.GitLabUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, CreateUserResponse{
			Success: false,
			Message: "获取GitLab用户信息失败",
			Error:   err.Error(),
		})
		return
	}

	// 构建响应数据
	userInfo := &UserInfo{
		ID:        gitlabUser.ID,
		Username:  gitlabUser.Username,
		Email:     gitlabUser.Email,
		Name:      gitlabUser.Name,
		Role:      models.ConvertExternalRoleToGitLab(externalUser.Role).GetEducationRole(),
		AvatarURL: gitlabUser.Avatar,
		IsActive:  externalUser.IsActive,
		CreatedAt: externalUser.CreatedAt,
	}

	externalUserInfo := &ExternalUserInfo{
		ID:             externalUser.ID,
		ExternalID:     externalUser.ExternalID,
		ExternalSource: externalUser.ExternalSource,
		Department:     externalUser.Department,
		StudentID:      externalUser.StudentID,
		TeacherID:      externalUser.TeacherID,
		Phone:          externalUser.Phone,
	}

	c.JSON(http.StatusOK, CreateUserResponse{
		Success: true,
		Message: "获取用户信息成功",
		Data: &UserTokenData{
			User:         userInfo,
			ExternalUser: externalUserInfo,
		},
	})
}

// generateJWTToken 生成JWT Token
func (h *SyncHandler) generateJWTToken(user *models.GitLabUser) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"name":     user.Name,
		"is_admin": user.IsAdmin,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}

// getSystemAdminToken 获取系统管理员Token
func (h *SyncHandler) getSystemAdminToken() (string, error) {
	// 这里应该使用配置中的系统Token或者通过其他方式获取管理员权限
	return h.gitlabService.GetSystemToken(), nil
}
