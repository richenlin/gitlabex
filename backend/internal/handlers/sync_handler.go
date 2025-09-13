package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"gitlabex/internal/middleware"
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
	// 用户管理已完全迁移到GitLab
	c.JSON(http.StatusNotImplemented, CreateUserResponse{
		Success: false,
		Message: "用户创建已迁移到GitLab",
		Error:   "请在GitLab中管理用户",
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
	Role       string `json:"role,omitempty"`
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
	// 用户管理已完全迁移到GitLab
	c.JSON(http.StatusNotImplemented, CreateUserResponse{
		Success: false,
		Message: "用户更新已迁移到GitLab",
		Error:   "请在GitLab中管理用户",
	})
}

// TODO: 移除旧的用户角色转换函数，现在使用GitLab角色系统
// convertStringToUserRole 转换字符串到用户角色类型（已废弃）
// func convertStringToUserRole(role string) models.UserRole { ... }

// TODO: 移除旧的教育角色转换函数，现在使用GitLab角色系统
// convertRoleToEducationRole 转换角色到教育角色类型（已废弃）
// func convertRoleToEducationRole(role string) models.EducationRole { ... }

// TODO: 重构用户创建逻辑以使用GitLab用户系统
// createSingleUser 创建单个用户 (内部方法) - 已废弃
func (h *SyncHandler) createSingleUser(req CreateUserRequest) CreateUserResponse {
	// 现在用户管理完全基于GitLab，不再支持本地用户创建
	return CreateUserResponse{
		Success: false,
		Message: "用户创建已迁移到GitLab",
		Error:   "请在GitLab中管理用户",
	}
}
