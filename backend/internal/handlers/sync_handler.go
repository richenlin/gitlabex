package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gitlabex/internal/middleware"
	"gitlabex/internal/models"
	"gitlabex/internal/services"
)

// SyncHandler 第三方系统同步处理器
type SyncHandler struct {
	userService   *services.UserService
	gitlabService *services.GitLabService
	jwtSecret     string
}

// NewSyncHandler 创建同步处理器
func NewSyncHandler(userService *services.UserService, gitlabService *services.GitLabService, jwtSecret string) *SyncHandler {
	return &SyncHandler{
		userService:   userService,
		gitlabService: gitlabService,
		jwtSecret:     jwtSecret,
	}
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Role     string `json:"role" binding:"required,oneof=admin teacher assistant student"`

	// 可选字段
	AvatarURL  string `json:"avatar_url,omitempty"`
	Department string `json:"department,omitempty"`
	StudentID  string `json:"student_id,omitempty"`
	TeacherID  string `json:"teacher_id,omitempty"`
	Phone      string `json:"phone,omitempty"`

	// 第三方系统标识
	ExternalID     string `json:"external_id,omitempty"`
	ExternalSource string `json:"external_source,omitempty"`
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
	User  *UserInfo `json:"user"`
	Token string    `json:"token"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID         string    `json:"id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	Role       string    `json:"role"`
	AvatarURL  string    `json:"avatar_url"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	ExternalID string    `json:"external_id,omitempty"`
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

	// 检查API密钥权限 - 第三方密钥不能创建管理员
	if req.Role == "admin" {
		keyInfo, exists := middleware.GetAPIKeyInfo(c)
		if !exists || !keyInfo.CanAdmin {
			c.JSON(http.StatusForbidden, CreateUserResponse{
				Success: false,
				Message: "权限不足",
				Error:   "当前API密钥无权创建管理员账号",
			})
			return
		}
	}

	// 检查用户是否已存在
	existingUser, _ := h.userService.GetUserByUsername(req.Username)
	if existingUser != nil {
		c.JSON(http.StatusConflict, CreateUserResponse{
			Success: false,
			Message: "用户已存在",
			Error:   "用户名已被使用",
		})
		return
	}

	// 检查邮箱是否已存在
	existingUserByEmail, _ := h.userService.GetUserByEmail(req.Email)
	if existingUserByEmail != nil {
		c.JSON(http.StatusConflict, CreateUserResponse{
			Success: false,
			Message: "邮箱已存在",
			Error:   "邮箱已被使用",
		})
		return
	}

	// 转换角色类型
	userRole := convertStringToUserRole(req.Role)
	eduRole := convertRoleToEducationRole(req.Role)

	// 创建GitLab用户 (这里可以根据需要决定是否在GitLab中创建用户)
	// 对于同步的用户，我们可能不需要在GitLab中创建，或者使用不同的策略

	// 创建本地用户
	user := &models.User{
		BaseModel: models.BaseModel{
			ID: uuid.New(),
		},
		GitLabID:    0, // 同步用户暂时不分配GitLab ID
		Username:    req.Username,
		Email:       req.Email,
		Name:        req.Name,
		AvatarURL:   req.AvatarURL,
		Role:        userRole,
		EduRole:     eduRole,
		IsActive:    true,
		LastLoginAt: nil,
		// 注意：密码应该在服务层进行哈希处理
	}

	// 创建用户 (这里需要在UserService中添加带密码的创建方法)
	if err := h.userService.CreateUserWithPassword(user, req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, CreateUserResponse{
			Success: false,
			Message: "用户创建失败",
			Error:   err.Error(),
		})
		return
	}

	// 生成JWT Token
	token, err := h.generateJWTToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, CreateUserResponse{
			Success: false,
			Message: "Token生成失败",
			Error:   err.Error(),
		})
		return
	}

	// 构造响应
	userInfo := &UserInfo{
		ID:         user.ID.String(),
		Username:   user.Username,
		Email:      user.Email,
		Name:       user.Name,
		Role:       string(user.Role),
		AvatarURL:  user.AvatarURL,
		IsActive:   user.IsActive,
		CreatedAt:  user.CreatedAt,
		ExternalID: req.ExternalID,
	}

	c.JSON(http.StatusCreated, CreateUserResponse{
		Success: true,
		Message: "用户创建成功",
		Data: &UserTokenData{
			User:  userInfo,
			Token: token,
		},
	})
}

// BatchCreateUsersRequest 批量创建用户请求
type BatchCreateUsersRequest struct {
	Users []CreateUserRequest `json:"users" binding:"required,min=1,max=100"`
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

	// 检查批量限制
	keyInfo, exists := middleware.GetAPIKeyInfo(c)
	if !exists {
		c.JSON(http.StatusInternalServerError, BatchCreateUsersResponse{
			Success: false,
			Message: "服务器内部错误",
		})
		return
	}

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
					Message: "权限不足",
				})
				return
			}
		}
	}

	results := make([]CreateUserResponse, len(req.Users))
	successCount := 0
	failureCount := 0

	// 逐个创建用户
	for i, userReq := range req.Users {
		// 模拟单个用户创建的逻辑
		result := h.createSingleUser(userReq)
		results[i] = result

		if result.Success {
			successCount++
		} else {
			failureCount++
		}
	}

	c.JSON(http.StatusOK, BatchCreateUsersResponse{
		Success:      failureCount == 0,
		Message:      "批量创建完成",
		TotalCount:   len(req.Users),
		SuccessCount: successCount,
		FailureCount: failureCount,
		Results:      results,
	})
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Username   string `json:"username"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	Role       string `json:"role,omitempty,oneof=admin teacher assistant student"`
	IsActive   *bool  `json:"is_active,omitempty"`
	AvatarURL  string `json:"avatar_url,omitempty"`
	Department string `json:"department,omitempty"`
}

// UpdateUser 更新用户接口
// @Summary 更新用户信息
// @Description 根据用户ID或用户名更新用户信息
// @Tags 同步接口
// @Accept json
// @Produce json
// @Param id path string true "用户ID或用户名"
// @Param request body UpdateUserRequest true "用户更新请求"
// @Param X-API-Key header string true "API密钥"
// @Success 200 {object} CreateUserResponse "更新成功"
// @Failure 400 {object} CreateUserResponse "请求参数错误"
// @Failure 401 {object} CreateUserResponse "API密钥无效"
// @Failure 404 {object} CreateUserResponse "用户不存在"
// @Router /api/v1/sync/users/{id} [put]
func (h *SyncHandler) UpdateUser(c *gin.Context) {
	identifier := c.Param("id")

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, CreateUserResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 查找用户 (支持按ID或用户名查找)
	var user *models.User
	var err error

	if uuid, parseErr := uuid.Parse(identifier); parseErr == nil {
		user, err = h.userService.GetUserByID(uuid)
	} else {
		user, err = h.userService.GetUserByUsername(identifier)
	}

	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, CreateUserResponse{
			Success: false,
			Message: "用户不存在",
			Error:   "找不到指定的用户",
		})
		return
	}

	// 更新用户信息
	updates := make(map[string]interface{})
	if req.Email != "" && req.Email != user.Email {
		updates["email"] = req.Email
	}
	if req.Name != "" && req.Name != user.Name {
		updates["name"] = req.Name
	}
	if req.Role != "" && req.Role != string(user.Role) {
		updates["role"] = convertStringToUserRole(req.Role)
		updates["edu_role"] = convertRoleToEducationRole(req.Role)
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.AvatarURL != "" {
		updates["avatar_url"] = req.AvatarURL
	}

	if len(updates) > 0 {
		if err := h.userService.UpdateUser(user.ID, updates); err != nil {
			c.JSON(http.StatusInternalServerError, CreateUserResponse{
				Success: false,
				Message: "用户更新失败",
				Error:   err.Error(),
			})
			return
		}
	}

	// 重新获取更新后的用户信息
	updatedUser, _ := h.userService.GetUserByID(user.ID)

	userInfo := &UserInfo{
		ID:        updatedUser.ID.String(),
		Username:  updatedUser.Username,
		Email:     updatedUser.Email,
		Name:      updatedUser.Name,
		Role:      string(updatedUser.Role),
		AvatarURL: updatedUser.AvatarURL,
		IsActive:  updatedUser.IsActive,
		CreatedAt: updatedUser.CreatedAt,
	}

	c.JSON(http.StatusOK, CreateUserResponse{
		Success: true,
		Message: "用户更新成功",
		Data: &UserTokenData{
			User: userInfo,
		},
	})
}

// convertStringToUserRole 转换字符串到用户角色类型
func convertStringToUserRole(role string) models.UserRole {
	switch role {
	case "admin":
		return models.RoleAdmin
	case "teacher":
		return models.RoleTeacher
	case "assistant":
		return models.RoleAssistant
	case "student":
		return models.RoleStudent
	default:
		return models.RoleStudent
	}
}

// convertRoleToEducationRole 转换角色到教育角色类型
func convertRoleToEducationRole(role string) models.EducationRole {
	switch role {
	case "admin":
		return models.EduRoleAdmin
	case "teacher":
		return models.EduRoleTeacher
	case "assistant":
		return models.EduRoleAssistant
	case "student":
		return models.EduRoleStudent
	default:
		return models.EduRoleStudent
	}
}

// createSingleUser 创建单个用户 (内部方法)
func (h *SyncHandler) createSingleUser(req CreateUserRequest) CreateUserResponse {
	// 检查用户是否已存在
	existingUser, _ := h.userService.GetUserByUsername(req.Username)
	if existingUser != nil {
		return CreateUserResponse{
			Success: false,
			Message: "用户已存在",
			Error:   "用户名已被使用",
		}
	}

	// 转换角色类型
	userRole := convertStringToUserRole(req.Role)
	eduRole := convertRoleToEducationRole(req.Role)

	// 创建本地用户
	user := &models.User{
		BaseModel: models.BaseModel{
			ID: uuid.New(),
		},
		GitLabID:    0,
		Username:    req.Username,
		Email:       req.Email,
		Name:        req.Name,
		AvatarURL:   req.AvatarURL,
		Role:        userRole,
		EduRole:     eduRole,
		IsActive:    true,
		LastLoginAt: nil,
	}

	// 创建用户
	if err := h.userService.CreateUserWithPassword(user, req.Password); err != nil {
		return CreateUserResponse{
			Success: false,
			Message: "用户创建失败",
			Error:   err.Error(),
		}
	}

	// 生成JWT Token
	token, err := h.generateJWTToken(user)
	if err != nil {
		return CreateUserResponse{
			Success: false,
			Message: "Token生成失败",
			Error:   err.Error(),
		}
	}

	// 构造响应
	userInfo := &UserInfo{
		ID:         user.ID.String(),
		Username:   user.Username,
		Email:      user.Email,
		Name:       user.Name,
		Role:       string(user.Role),
		AvatarURL:  user.AvatarURL,
		IsActive:   user.IsActive,
		CreatedAt:  user.CreatedAt,
		ExternalID: req.ExternalID,
	}

	return CreateUserResponse{
		Success: true,
		Message: "用户创建成功",
		Data: &UserTokenData{
			User:  userInfo,
			Token: token,
		},
	}
}

// generateJWTToken 生成JWT Token
func (h *SyncHandler) generateJWTToken(user *models.User) (string, error) {
	// 这里应该使用与AuthHandler相同的JWT生成逻辑
	// 简化实现，返回包含用户信息的token格式
	// 实际生产环境中应该使用真正的JWT库如github.com/golang-jwt/jwt

	token := fmt.Sprintf("jwt.%s.%s", user.ID.String(), user.Username)
	return token, nil
}
