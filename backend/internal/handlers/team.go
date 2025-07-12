package handlers

import (
	"net/http"
	"strconv"

	"gitlabex/internal/models"
	"gitlabex/internal/services"

	"github.com/gin-gonic/gin"
)

// TeamHandler 团队管理处理器
type TeamHandler struct {
	teamService *services.TeamService
	userService *services.UserService
}

// NewTeamHandler 创建团队管理处理器
func NewTeamHandler(teamService *services.TeamService, userService *services.UserService) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
		userService: userService,
	}
}

// RegisterRoutes 注册路由
func (h *TeamHandler) RegisterRoutes(router *gin.RouterGroup) {
	teams := router.Group("/teams")
	{
		// 班级管理
		teams.POST("/classes", h.CreateClass)
		teams.POST("/teams", h.CreateTeam)
		teams.GET("/user/:userId", h.GetUserTeams)
		teams.GET("/:teamId", h.GetTeamDetails)
		teams.GET("/:teamId/activity", h.GetTeamActivity)
		teams.GET("/search", h.SearchTeams)

		// 成员管理
		teams.POST("/:teamId/members", h.AddTeamMember)
		teams.DELETE("/:teamId/members/:userId", h.RemoveTeamMember)
		teams.PUT("/:teamId/members/:userId", h.UpdateMemberRole)
		teams.GET("/:teamId/members", h.GetTeamMembers)
	}
}

// CreateClass 创建班级
func (h *TeamHandler) CreateClass(c *gin.Context) {
	var req CreateClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	group, err := h.teamService.CreateClass(req.Name, req.Description, req.TeacherID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建班级失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "班级创建成功",
		"data":    group,
	})
}

// CreateTeam 创建团队
func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var req CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	group, err := h.teamService.CreateTeam(req.ParentGroupID, req.Name, req.Description, req.LeaderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建团队失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "团队创建成功",
		"data":    group,
	})
}

// GetUserTeams 获取用户所属团队
func (h *TeamHandler) GetUserTeams(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID格式错误"})
		return
	}

	teams, err := h.teamService.GetUserTeams(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户团队失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取用户团队成功",
		"data":    teams,
	})
}

// GetTeamDetails 获取团队详情
func (h *TeamHandler) GetTeamDetails(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "团队ID格式错误"})
		return
	}

	details, err := h.teamService.GetTeamDetails(teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取团队详情失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取团队详情成功",
		"data":    details,
	})
}

// GetTeamActivity 获取团队活动
func (h *TeamHandler) GetTeamActivity(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "团队ID格式错误"})
		return
	}

	// 获取限制数量，默认为20
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	activities, err := h.teamService.GetTeamActivity(teamID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取团队活动失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取团队活动成功",
		"data":    activities,
	})
}

// SearchTeams 搜索团队
func (h *TeamHandler) SearchTeams(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	teams, err := h.teamService.SearchTeams(query, userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索团队失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "搜索团队成功",
		"data":    teams,
	})
}

// AddTeamMember 添加团队成员
func (h *TeamHandler) AddTeamMember(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "团队ID格式错误"})
		return
	}

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	err = h.teamService.AddTeamMember(teamID, req.UserID, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加团队成员失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "添加团队成员成功",
	})
}

// RemoveTeamMember 移除团队成员
func (h *TeamHandler) RemoveTeamMember(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "团队ID格式错误"})
		return
	}

	userIDStr := c.Param("userId")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID格式错误"})
		return
	}

	err = h.teamService.RemoveTeamMember(teamID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "移除团队成员失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "移除团队成员成功",
	})
}

// UpdateMemberRole 更新成员角色
func (h *TeamHandler) UpdateMemberRole(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "团队ID格式错误"})
		return
	}

	userIDStr := c.Param("userId")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID格式错误"})
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	err = h.teamService.UpdateMemberRole(teamID, userID, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新成员角色失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成员角色成功",
	})
}

// GetTeamMembers 获取团队成员列表
func (h *TeamHandler) GetTeamMembers(c *gin.Context) {
	teamIDStr := c.Param("teamId")
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "团队ID格式错误"})
		return
	}

	members, err := h.teamService.GetTeamMembers(teamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取团队成员失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取团队成员成功",
		"data":    members,
	})
}

// 请求数据结构

// CreateClassRequest 创建班级请求
type CreateClassRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	TeacherID   int    `json:"teacher_id" binding:"required"`
}

// CreateTeamRequest 创建团队请求
type CreateTeamRequest struct {
	ParentGroupID int    `json:"parent_group_id" binding:"required"`
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	LeaderID      int    `json:"leader_id" binding:"required"`
}

// AddMemberRequest 添加成员请求
type AddMemberRequest struct {
	UserID int                  `json:"user_id" binding:"required"`
	Role   models.EducationRole `json:"role" binding:"required"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Role models.EducationRole `json:"role" binding:"required"`
}
