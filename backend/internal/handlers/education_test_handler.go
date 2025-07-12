package handlers

import (
	"net/http"
	"strconv"

	"gitlabex/internal/services"

	"github.com/gin-gonic/gin"
)

// EducationTestHandler 教育工作流测试处理器
type EducationTestHandler struct {
	educationService *services.EducationServiceSimplified
}

// NewEducationTestHandler 创建教育工作流测试处理器
func NewEducationTestHandler(educationService *services.EducationServiceSimplified) *EducationTestHandler {
	return &EducationTestHandler{
		educationService: educationService,
	}
}

// RegisterRoutes 注册路由
func (h *EducationTestHandler) RegisterRoutes(router *gin.RouterGroup) {
	test := router.Group("/education/test")
	{
		test.POST("/workflow/:groupId", h.TestWorkflow)
		test.GET("/stats/:userId", h.GetEducationStats)
		test.POST("/project/:groupId", h.CreateProject)
		test.POST("/assignment/:projectId", h.CreateAssignment)
		test.POST("/announcement/:projectId", h.CreateAnnouncement)
		test.GET("/assignments/:projectId", h.GetAssignments)
		test.GET("/submissions/:projectId", h.GetSubmissions)
	}
}

// TestWorkflow 测试完整的教育工作流
func (h *EducationTestHandler) TestWorkflow(c *gin.Context) {
	groupID, err := strconv.Atoi(c.Param("groupId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	result, err := h.educationService.TestEducationWorkflow(groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetEducationStats 获取教育统计数据
func (h *EducationTestHandler) GetEducationStats(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	stats, err := h.educationService.GetSimpleEducationStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// CreateProject 创建项目
func (h *EducationTestHandler) CreateProject(c *gin.Context) {
	groupID, err := strconv.Atoi(c.Param("groupId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.educationService.CreateSimpleProject(groupID, req.Title, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// CreateAssignment 创建作业
func (h *EducationTestHandler) CreateAssignment(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("projectId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	assignment, err := h.educationService.CreateSimpleAssignment(projectID, req.Title, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, assignment)
}

// CreateAnnouncement 创建公告
func (h *EducationTestHandler) CreateAnnouncement(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("projectId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	announcement, err := h.educationService.CreateSimpleAnnouncement(projectID, req.Title, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, announcement)
}

// GetAssignments 获取作业列表
func (h *EducationTestHandler) GetAssignments(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("projectId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	assignments, err := h.educationService.GetSimpleAssignments(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, assignments)
}

// GetSubmissions 获取提交列表
func (h *EducationTestHandler) GetSubmissions(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("projectId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	submissions, err := h.educationService.GetSimpleSubmissions(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, submissions)
}
