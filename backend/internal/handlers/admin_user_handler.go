package handlers

import (
	"gitlabex/internal/services"
	"gitlabex/internal/types"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AdminUserHandler 管理员用户处理器
type AdminUserHandler struct {
	userService *services.UserService
}

// NewAdminUserHandler 创建管理员用户处理器
func NewAdminUserHandler(userService *services.UserService) *AdminUserHandler {
	return &AdminUserHandler{
		userService: userService,
	}
}

// GetUsers 获取用户列表 (管理员专用)
func (h *AdminUserHandler) GetUsers(c *gin.Context) {
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	// 检查管理员权限
	if !h.checkAdminPermission(accessToken.(string)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	search := c.Query("search")

	// 获取用户列表
	users, total, err := h.userService.GetAllUsers(accessToken.(string), page, pageSize, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取用户列表失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users":      users,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_page": (total + pageSize - 1) / pageSize,
	})
}

// CreateUser 创建用户 (管理员专用)
func (h *AdminUserHandler) CreateUser(c *gin.Context) {
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	// 检查管理员权限
	if !h.checkAdminPermission(accessToken.(string)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	var req types.AdminCreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 创建用户
	user, err := h.userService.CreateUser(accessToken.(string), &services.CreateUserData{
		Username:    req.Username,
		Name:        req.Name,
		Email:       req.Email,
		Password:    req.Password,
		IsAdmin:     req.IsAdmin,
		DefaultRole: req.DefaultRole,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建用户失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// UpdateUser 更新用户信息 (管理员专用)
func (h *AdminUserHandler) UpdateUser(c *gin.Context) {
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	// 检查管理员权限
	if !h.checkAdminPermission(accessToken.(string)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	userIDStr := c.Param("id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	var req types.AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 更新用户信息
	user, err := h.userService.UpdateUser(accessToken.(string), userIDStr, &services.UpdateUserData{
		Username: req.Username,
		Name:     req.Name,
		Email:    req.Email,
		IsAdmin:  req.IsAdmin,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新用户失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser 删除用户 (管理员专用)
func (h *AdminUserHandler) DeleteUser(c *gin.Context) {
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	// 检查管理员权限
	if !h.checkAdminPermission(accessToken.(string)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	userIDStr := c.Param("id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	// 删除用户
	err := h.userService.DeleteUser(accessToken.(string), userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "删除用户失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户删除成功"})
}

// GetUserDetails 获取用户详情 (管理员专用)
func (h *AdminUserHandler) GetUserDetails(c *gin.Context) {
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	// 检查管理员权限
	if !h.checkAdminPermission(accessToken.(string)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	userIDStr := c.Param("id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	// 转换用户ID为int64
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 获取用户详情
	user, err := h.userService.GetUserByGitLabID(accessToken.(string), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUserRoles 更新用户角色 (管理员专用)
func (h *AdminUserHandler) UpdateUserRoles(c *gin.Context) {
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	// 检查管理员权限
	if !h.checkAdminPermission(accessToken.(string)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	userIDStr := c.Param("id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	var req types.AdminUpdateUserRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 更新用户角色
	err := h.userService.UpdateUserRoles(accessToken.(string), userIDStr, &services.UpdateUserRolesData{
		IsAdmin: req.IsAdmin,
		// TODO: 处理项目角色
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新用户角色失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户角色更新成功"})
}

// GetUserProjectRoles 获取用户项目角色 (管理员专用)
func (h *AdminUserHandler) GetUserProjectRoles(c *gin.Context) {
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	// 检查管理员权限
	if !h.checkAdminPermission(accessToken.(string)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	userIDStr := c.Param("id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	// 获取用户项目角色
	roles, err := h.userService.GetUserProjectRoles(accessToken.(string), userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取用户项目角色失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

// GetUserStats 获取用户统计信息 (管理员专用)
func (h *AdminUserHandler) GetUserStats(c *gin.Context) {
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	// 检查管理员权限
	if !h.checkAdminPermission(accessToken.(string)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
		return
	}

	// 获取用户统计
	stats, err := h.userService.GetUserStats(accessToken.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取用户统计失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// checkAdminPermission 检查管理员权限
func (h *AdminUserHandler) checkAdminPermission(accessToken string) bool {
	// 获取当前用户信息并检查是否为管理员
	user, err := h.userService.GetCurrentUser(accessToken)
	if err != nil {
		return false
	}

	return user.IsAdmin
}
