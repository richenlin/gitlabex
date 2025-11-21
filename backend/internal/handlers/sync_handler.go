package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"gitlabex/internal/dto"
	"gitlabex/internal/services"
)

// SyncHandler 第三方系统同步处理器 - 最终精简版
type SyncHandler struct {
	gitlabService *services.GitLabService
}

// NewSyncHandler 创建同步处理器
func NewSyncHandler(gitlabService *services.GitLabService) *SyncHandler {
	return &SyncHandler{
		gitlabService: gitlabService,
	}
}

// CreateUser 创建用户接口
// @Summary 第三方系统创建用户
// @Description 为第三方系统提供用户创建接口，如果用户已存在则直接返回用户信息
// @Tags 第三方同步接口
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "用户创建请求"
// @Param X-API-Key header string true "第三方API密钥"
// @Success 201 {object} CreateUserResponse "用户创建成功"
// @Success 200 {object} CreateUserResponse "用户已存在"
// @Failure 400 {object} CreateUserResponse "请求参数错误"
// @Failure 401 {object} CreateUserResponse "API密钥无效"
// @Failure 500 {object} CreateUserResponse "服务器内部错误"
// @Router /api/v1/sync/users [post]
func (h *SyncHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.CreateUserResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 获取系统管理员Token进行GitLab操作
	adminToken := h.gitlabService.GetSystemToken()
	if adminToken == "" {
		c.JSON(http.StatusInternalServerError, dto.CreateUserResponse{
			Success: false,
			Message: "服务器配置错误",
			Error:   "系统管理员Token未配置",
		})
		return
	}

	// 检查用户是否已存在
	existingUser, err := h.gitlabService.GetUserByUsername(adminToken, req.Username)
	if err == nil && existingUser != nil {
		// 用户已存在，直接返回用户信息
		userData := &dto.UserData{
			ID:       existingUser.ID,
			Username: existingUser.Username,
			Email:    existingUser.Email,
			Name:     existingUser.Name,
			Role:     getRoleDisplayName(req.Role),
		}

		c.JSON(http.StatusOK, dto.CreateUserResponse{
			Success: true,
			Message: "用户已存在，可通过GitLab OAuth登录GitLabEx",
			Data:    userData,
		})
		return
	}

	// 在GitLab中创建用户
	gitlabCreateData := &dto.GitLabCreateUserData{
		Email:            req.Email,
		Username:         req.Username,
		Name:             req.Name,
		Password:         req.Password,
		SkipConfirmation: true, // 跳过邮箱确认，允许直接登录
		Admin:            req.Role == "admin",
	}

	gitlabUser, err := h.gitlabService.CreateUser(adminToken, gitlabCreateData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.CreateUserResponse{
			Success: false,
			Message: "GitLab用户创建失败",
			Error:   err.Error(),
		})
		return
	}

	// 根据角色将用户分配到相应的GitLab用户组
	if err := h.gitlabService.AssignUserToRoleGroup(adminToken, gitlabUser.ID, req.Role); err != nil {
		// 用户组分配失败不应该阻止用户创建，只记录警告
		fmt.Printf("警告：为用户 %d 分配用户组失败: %v\n", gitlabUser.ID, err)
	}

	// 构建响应数据
	userData := &dto.UserData{
		ID:       gitlabUser.ID,
		Username: gitlabUser.Username,
		Email:    gitlabUser.Email,
		Name:     gitlabUser.Name,
		Role:     getRoleDisplayName(req.Role),
	}

	c.JSON(http.StatusCreated, dto.CreateUserResponse{
		Success: true,
		Message: "用户创建成功，可通过GitLab OAuth登录GitLabEx",
		Data:    userData,
	})
}

// GetUser 获取用户信息接口
// @Summary 获取用户信息
// @Description 根据用户名获取GitLab用户信息
// @Tags 第三方同步接口
// @Produce json
// @Param username path string true "用户名"
// @Param X-API-Key header string true "第三方API密钥"
// @Success 200 {object} GetUserResponse "获取成功"
// @Failure 401 {object} GetUserResponse "API密钥无效"
// @Failure 404 {object} GetUserResponse "用户不存在"
// @Failure 500 {object} GetUserResponse "服务器内部错误"
// @Router /api/v1/sync/users/{username} [get]
func (h *SyncHandler) GetUser(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, dto.GetUserResponse{
			Success: false,
			Message: "用户名参数不能为空",
		})
		return
	}

	// 获取系统管理员Token
	adminToken := h.gitlabService.GetSystemToken()
	if adminToken == "" {
		c.JSON(http.StatusInternalServerError, dto.GetUserResponse{
			Success: false,
			Message: "服务器配置错误",
			Error:   "系统管理员Token未配置",
		})
		return
	}

	// 查询GitLab用户信息
	gitlabUser, err := h.gitlabService.GetUserByUsername(adminToken, username)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.GetUserResponse{
			Success: false,
			Message: "用户不存在",
			Error:   err.Error(),
		})
		return
	}

	// 判断用户角色（基于GitLab的admin权限）
	roleDisplayName := "用户"
	if gitlabUser.IsAdmin {
		roleDisplayName = "管理员"
	}

	// 构建响应数据
	userData := &dto.UserData{
		ID:       gitlabUser.ID,
		Username: gitlabUser.Username,
		Email:    gitlabUser.Email,
		Name:     gitlabUser.Name,
		Role:     roleDisplayName,
	}

	c.JSON(http.StatusOK, dto.GetUserResponse{
		Success: true,
		Message: "获取用户信息成功",
		Data:    userData,
	})
}

// getRoleDisplayName 获取角色显示名称
func getRoleDisplayName(role string) string {
	switch role {
	case "admin":
		return "管理员"
	case "teacher":
		return "教师"
	case "researcher":
		return "研究员"
	case "student":
		return "学生"
	case "guest":
		return "访客"
	default:
		return "用户"
	}
}
