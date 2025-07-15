package handlers

import (
	"net/http"
	"strconv"

	"gitlabex/internal/models"
	"gitlabex/internal/services"

	"github.com/gin-gonic/gin"
)

// InteractiveDevHandler 互动开发处理器
type InteractiveDevHandler struct {
	interactiveDevService *services.InteractiveDevService
	permissionService     *services.PermissionService
}

// NewInteractiveDevHandler 创建互动开发处理器
func NewInteractiveDevHandler(interactiveDevService *services.InteractiveDevService, permissionService *services.PermissionService) *InteractiveDevHandler {
	return &InteractiveDevHandler{
		interactiveDevService: interactiveDevService,
		permissionService:     permissionService,
	}
}

// GetProjectFileTree 获取项目文件树
func (h *InteractiveDevHandler) GetProjectFileTree(c *gin.Context) {
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	currentUser := user.(*models.User)

	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid project ID",
		})
		return
	}

	branch := c.Query("branch")
	if branch == "" {
		branch = "main"
	}

	fileTree, err := h.interactiveDevService.GetProjectFileTree(uint(projectID), branch, currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get file tree",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"file_tree": fileTree,
	})
}

// GetFileContent 获取文件内容
func (h *InteractiveDevHandler) GetFileContent(c *gin.Context) {
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	currentUser := user.(*models.User)

	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid project ID",
		})
		return
	}

	filePath := c.Query("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File path is required",
		})
		return
	}

	branch := c.Query("branch")
	if branch == "" {
		branch = "main"
	}

	content, err := h.interactiveDevService.GetFileContent(uint(projectID), filePath, branch, currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get file content",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"content": content,
		"path":    filePath,
		"branch":  branch,
	})
}

// SaveFileContent 保存文件内容
func (h *InteractiveDevHandler) SaveFileContent(c *gin.Context) {
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	currentUser := user.(*models.User)

	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid project ID",
		})
		return
	}

	var req services.CodeEditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	req.ProjectID = uint(projectID)

	err = h.interactiveDevService.SaveFileContent(&req, currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to save file",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "File saved successfully",
	})
}

// CreateStudentBranch 为学生创建个人分支
func (h *InteractiveDevHandler) CreateStudentBranch(c *gin.Context) {
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	currentUser := user.(*models.User)

	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid project ID",
		})
		return
	}

	err = h.interactiveDevService.CreateStudentBranch(uint(projectID), currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create student branch",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Student branch created successfully",
	})
}

// StartEditSession 开始编辑会话
func (h *InteractiveDevHandler) StartEditSession(c *gin.Context) {
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	currentUser := user.(*models.User)

	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid project ID",
		})
		return
	}

	var req services.EditSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	req.ProjectID = uint(projectID)

	session, err := h.interactiveDevService.StartEditSession(&req, currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to start edit session",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"session": session,
	})
}

// UpdateEditSession 更新编辑会话
func (h *InteractiveDevHandler) UpdateEditSession(c *gin.Context) {
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	currentUser := user.(*models.User)

	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Session ID is required",
		})
		return
	}

	err := h.interactiveDevService.UpdateEditSession(sessionID, currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update edit session",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Edit session updated successfully",
	})
}

// EndEditSession 结束编辑会话
func (h *InteractiveDevHandler) EndEditSession(c *gin.Context) {
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	currentUser := user.(*models.User)

	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Session ID is required",
		})
		return
	}

	err := h.interactiveDevService.EndEditSession(sessionID, currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to end edit session",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Edit session ended successfully",
	})
}

// GetActiveEditSessions 获取活跃编辑会话
func (h *InteractiveDevHandler) GetActiveEditSessions(c *gin.Context) {
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	currentUser := user.(*models.User)

	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid project ID",
		})
		return
	}

	// 检查权限
	if !h.permissionService.CanAccessProject(currentUser, uint(projectID), "read") {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Permission denied",
		})
		return
	}

	sessions, err := h.interactiveDevService.GetActiveEditSessions(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get active edit sessions",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"sessions": sessions,
	})
}

// GetCurrentMember 获取当前用户的项目成员信息
func (h *InteractiveDevHandler) GetCurrentMember(c *gin.Context) {
	user, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	currentUser := user.(*models.User)

	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid project ID",
		})
		return
	}

	// 检查权限
	if !h.permissionService.CanAccessProject(currentUser, uint(projectID), "read") {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Permission denied",
		})
		return
	}

	// 查找项目成员信息
	member, err := h.interactiveDevService.GetCurrentMember(uint(projectID), currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get member info",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"member":  member,
	})
}
