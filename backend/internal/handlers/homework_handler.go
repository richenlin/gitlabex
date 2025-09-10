package handlers

import (
	"gitlabex/internal/models"
	"gitlabex/internal/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// HomeworkHandler 作业处理器
type HomeworkHandler struct {
	homeworkService *services.HomeworkService
	userService     *services.UserService
}

// NewHomeworkHandler 创建作业处理器
func NewHomeworkHandler(homeworkService *services.HomeworkService, userService *services.UserService) *HomeworkHandler {
	return &HomeworkHandler{
		homeworkService: homeworkService,
		userService:     userService,
	}
}

// CreateHomework 创建作业
func (h *HomeworkHandler) CreateHomework(c *gin.Context) {
	var req struct {
		ProjectID   string     `json:"project_id" binding:"required"`
		Title       string     `json:"title" binding:"required"`
		Description string     `json:"description"`
		DueDate     *time.Time `json:"due_date"`
		MaxGrade    int        `json:"max_grade"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	homework := &models.Homework{
		ProjectID:   projectID,
		Title:       req.Title,
		Description: req.Description,
		DueDate:     req.DueDate,
		MaxGrade:    req.MaxGrade,
		CreatorID:   userID.(uuid.UUID),
		Status:      "active",
	}

	if err := h.homeworkService.CreateHomework(homework); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建作业失败"})
		return
	}

	c.JSON(http.StatusCreated, homework)
}

// GetHomeworkByID 根据ID获取作业
func (h *HomeworkHandler) GetHomeworkByID(c *gin.Context) {
	homeworkIDStr := c.Param("id")
	homeworkID, err := uuid.Parse(homeworkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的作业ID"})
		return
	}

	homework, err := h.homeworkService.GetHomeworkByID(homeworkID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "作业不存在"})
		return
	}

	c.JSON(http.StatusOK, homework)
}

// GetHomeworkByProject 获取项目下的所有作业
func (h *HomeworkHandler) GetHomeworkByProject(c *gin.Context) {
	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目ID不能为空"})
		return
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	homeworks, err := h.homeworkService.GetHomeworkByProject(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取作业失败"})
		return
	}

	c.JSON(http.StatusOK, homeworks)
}

// UpdateHomework 更新作业信息
func (h *HomeworkHandler) UpdateHomework(c *gin.Context) {
	homeworkIDStr := c.Param("id")
	homeworkID, err := uuid.Parse(homeworkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的作业ID"})
		return
	}

	var req struct {
		Title       *string    `json:"title"`
		Description *string    `json:"description"`
		DueDate     *time.Time `json:"due_date"`
		MaxGrade    *int       `json:"max_grade"`
		Status      *string    `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.DueDate != nil {
		updates["due_date"] = *req.DueDate
	}
	if req.MaxGrade != nil {
		updates["max_grade"] = *req.MaxGrade
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := h.homeworkService.UpdateHomework(homeworkID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新作业失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "作业更新成功"})
}

// DeleteHomework 删除作业
func (h *HomeworkHandler) DeleteHomework(c *gin.Context) {
	homeworkIDStr := c.Param("id")
	homeworkID, err := uuid.Parse(homeworkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的作业ID"})
		return
	}

	if err := h.homeworkService.DeleteHomework(homeworkID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除作业失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "作业删除成功"})
}

// SubmitHomework 提交作业
func (h *HomeworkHandler) SubmitHomework(c *gin.Context) {
	homeworkIDStr := c.Param("id")
	homeworkID, err := uuid.Parse(homeworkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的作业ID"})
		return
	}

	var req struct {
		BranchName string `json:"branch_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	submission := &models.Submission{
		HomeworkID:   homeworkID,
		StudentID:    userID.(uuid.UUID),
		GitLabBranch: req.BranchName,
	}

	if err := h.homeworkService.SubmitHomework(submission); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交作业失败"})
		return
	}

	c.JSON(http.StatusCreated, submission)
}

// GetSubmissions 获取作业的所有提交
func (h *HomeworkHandler) GetSubmissions(c *gin.Context) {
	homeworkIDStr := c.Param("id")
	homeworkID, err := uuid.Parse(homeworkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的作业ID"})
		return
	}

	submissions, err := h.homeworkService.GetSubmissions(homeworkID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取提交失败"})
		return
	}

	c.JSON(http.StatusOK, submissions)
}

// GradeHomework 评分作业
func (h *HomeworkHandler) GradeHomework(c *gin.Context) {
	submissionIDStr := c.Param("id")
	submissionID, err := uuid.Parse(submissionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的提交ID"})
		return
	}

	var req struct {
		Grade    int    `json:"grade" binding:"required"`
		Feedback string `json:"feedback"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	graderID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	if err := h.homeworkService.GradeHomework(submissionID, req.Grade, req.Feedback, graderID.(uuid.UUID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "评分失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "评分成功"})
}

// GetUserSubmissions 获取用户的作业提交
func (h *HomeworkHandler) GetUserSubmissions(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	submissions, err := h.homeworkService.GetUserSubmissions(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取提交失败"})
		return
	}

	c.JSON(http.StatusOK, submissions)
}

// GetMySubmission 获取当前用户对指定作业的提交
func (h *HomeworkHandler) GetMySubmission(c *gin.Context) {
	homeworkIDStr := c.Param("id")
	homeworkID, err := uuid.Parse(homeworkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的作业ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	submission, err := h.homeworkService.GetUserSubmissionForHomework(homeworkID, userID.(uuid.UUID))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "尚未提交作业"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取提交失败"})
		}
		return
	}

	c.JSON(http.StatusOK, submission)
}

// GetSubmissionByID 根据ID获取提交
func (h *HomeworkHandler) GetSubmissionByID(c *gin.Context) {
	submissionIDStr := c.Param("id")
	submissionID, err := uuid.Parse(submissionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的提交ID"})
		return
	}

	submission, err := h.homeworkService.GetSubmissionByID(submissionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "提交不存在"})
		return
	}

	c.JSON(http.StatusOK, submission)
}

// GetHomeworkStats 获取作业统计信息
func (h *HomeworkHandler) GetHomeworkStats(c *gin.Context) {
	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目ID不能为空"})
		return
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	stats, err := h.homeworkService.GetHomeworkStats(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计信息失败"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetGradeDistribution 获取成绩分布
func (h *HomeworkHandler) GetGradeDistribution(c *gin.Context) {
	homeworkIDStr := c.Param("id")
	homeworkID, err := uuid.Parse(homeworkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的作业ID"})
		return
	}

	distribution, err := h.homeworkService.GetGradeDistribution(homeworkID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取成绩分布失败"})
		return
	}

	c.JSON(http.StatusOK, distribution)
}

// ExportGrades 导出成绩
func (h *HomeworkHandler) ExportGrades(c *gin.Context) {
	homeworkIDStr := c.Param("id")
	homeworkID, err := uuid.Parse(homeworkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的作业ID"})
		return
	}

	grades, err := h.homeworkService.ExportGrades(homeworkID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导出成绩失败"})
		return
	}

	c.JSON(http.StatusOK, grades)
}

// GetPendingReviews 获取待评分的作业
func (h *HomeworkHandler) GetPendingReviews(c *gin.Context) {
	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目ID不能为空"})
		return
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	graderID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	submissions, err := h.homeworkService.GetPendingReviews(projectID, graderID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取待评分作业失败"})
		return
	}

	c.JSON(http.StatusOK, submissions)
}

// GetStudentProgress 获取学生进度
func (h *HomeworkHandler) GetStudentProgress(c *gin.Context) {
	userIDStr := c.Query("user_id")
	projectIDStr := c.Query("project_id")

	if userIDStr == "" || projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID和项目ID不能为空"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	progress, err := h.homeworkService.GetStudentProgress(userID, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取学生进度失败"})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// GetAssignmentDetails 获取作业详情
func (h *HomeworkHandler) GetAssignmentDetails(c *gin.Context) {
	homeworkIDStr := c.Param("id")
	homeworkID, err := uuid.Parse(homeworkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的作业ID"})
		return
	}

	details, err := h.homeworkService.GetAssignmentDetails(homeworkID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取作业详情失败"})
		return
	}

	c.JSON(http.StatusOK, details)
}

// BulkUpdateDueDate 批量更新截止日期
func (h *HomeworkHandler) BulkUpdateDueDate(c *gin.Context) {
	var req struct {
		HomeworkIDs []string   `json:"homework_ids" binding:"required"`
		DueDate     *time.Time `json:"due_date"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var homeworkIDs []uuid.UUID
	for _, idStr := range req.HomeworkIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的作业ID"})
			return
		}
		homeworkIDs = append(homeworkIDs, id)
	}

	if err := h.homeworkService.BulkUpdateDueDate(homeworkIDs, req.DueDate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新截止日期失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "截止日期更新成功"})
}

// ArchiveHomework 归档作业
func (h *HomeworkHandler) ArchiveHomework(c *gin.Context) {
	homeworkIDStr := c.Param("id")
	homeworkID, err := uuid.Parse(homeworkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的作业ID"})
		return
	}

	if err := h.homeworkService.ArchiveHomework(homeworkID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "归档作业失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "作业已归档"})
}

// RestoreHomework 恢复作业
func (h *HomeworkHandler) RestoreHomework(c *gin.Context) {
	homeworkIDStr := c.Param("id")
	homeworkID, err := uuid.Parse(homeworkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的作业ID"})
		return
	}

	if err := h.homeworkService.RestoreHomework(homeworkID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "恢复作业失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "作业已恢复"})
}

// GetAssignmentTemplates 获取作业模板
func (h *HomeworkHandler) GetAssignmentTemplates(c *gin.Context) {
	isPublic := c.Query("is_public") == "true"

	var creatorID uuid.UUID
	if !isPublic {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
			return
		}
		creatorID = userID.(uuid.UUID)
	}

	templates, err := h.homeworkService.GetAssignmentTemplates(isPublic, creatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取模板失败"})
		return
	}

	c.JSON(http.StatusOK, templates)
}

// CreateAssignmentTemplate 创建作业模板
func (h *HomeworkHandler) CreateAssignmentTemplate(c *gin.Context) {
	var req struct {
		Title       string    `json:"title" binding:"required"`
		Description string    `json:"description"`
		DueDate     time.Time `json:"due_date"`
		MaxGrade    int       `json:"max_grade"`
		IsPublic    bool      `json:"is_public"`
		Tags        []string  `json:"tags"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	template := &models.AssignmentTemplate{
		Title:       req.Title,
		Description: req.Description,
		DueDate:     &req.DueDate,
		MaxGrade:    req.MaxGrade,
		CreatorID:   userID.(uuid.UUID),
		IsPublic:    req.IsPublic,
		Tags:        req.Tags,
	}

	if err := h.homeworkService.CreateAssignmentTemplate(template); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建模板失败"})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// UpdateAssignmentTemplate 更新作业模板
func (h *HomeworkHandler) UpdateAssignmentTemplate(c *gin.Context) {
	templateIDStr := c.Param("id")
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模板ID"})
		return
	}

	var req struct {
		Title       *string    `json:"title"`
		Description *string    `json:"description"`
		DueDate     *time.Time `json:"due_date"`
		MaxGrade    *int       `json:"max_grade"`
		IsPublic    *bool      `json:"is_public"`
		Tags        []string   `json:"tags"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.DueDate != nil {
		updates["due_date"] = *req.DueDate
	}
	if req.MaxGrade != nil {
		updates["max_grade"] = *req.MaxGrade
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}
	if len(req.Tags) > 0 {
		updates["tags"] = req.Tags
	}

	if err := h.homeworkService.UpdateAssignmentTemplate(templateID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新模板失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "模板更新成功"})
}

// DeleteAssignmentTemplate 删除作业模板
func (h *HomeworkHandler) DeleteAssignmentTemplate(c *gin.Context) {
	templateIDStr := c.Param("id")
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模板ID"})
		return
	}

	if err := h.homeworkService.DeleteAssignmentTemplate(templateID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除模板失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "模板删除成功"})
}

// UseAssignmentTemplate 使用模板创建作业
func (h *HomeworkHandler) UseAssignmentTemplate(c *gin.Context) {
	var req struct {
		TemplateID string `json:"template_id" binding:"required"`
		ProjectID  string `json:"project_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	templateID, err := uuid.Parse(req.TemplateID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模板ID"})
		return
	}

	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	homework, err := h.homeworkService.UseAssignmentTemplate(templateID, projectID, userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "使用模板失败"})
		return
	}

	c.JSON(http.StatusCreated, homework)
}

// BulkCreateHomework 批量创建作业
func (h *HomeworkHandler) BulkCreateHomework(c *gin.Context) {
	var req []struct {
		ProjectID   string     `json:"project_id" binding:"required"`
		Title       string     `json:"title" binding:"required"`
		Description string     `json:"description"`
		DueDate     *time.Time `json:"due_date"`
		MaxGrade    int        `json:"max_grade"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	var homeworks []models.Homework
	for _, hw := range req {
		projectID, err := uuid.Parse(hw.ProjectID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
			return
		}

		homework := models.Homework{
			ProjectID:   projectID,
			Title:       hw.Title,
			Description: hw.Description,
			DueDate:     hw.DueDate,
			MaxGrade:    hw.MaxGrade,
			CreatorID:   userID.(uuid.UUID),
			Status:      "active",
		}
		homeworks = append(homeworks, homework)
	}

	if err := h.homeworkService.BulkCreateHomework(homeworks); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量创建作业失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "批量创建作业成功", "count": len(homeworks)})
}

// ImportGrades 导入成绩
func (h *HomeworkHandler) ImportGrades(c *gin.Context) {
	homeworkIDStr := c.Param("id")
	homeworkID, err := uuid.Parse(homeworkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的作业ID"})
		return
	}

	var grades []map[string]interface{}
	if err := c.ShouldBindJSON(&grades); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.homeworkService.ImportGrades(homeworkID, grades); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导入成绩失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "成绩导入成功"})
}

// GenerateReport 生成作业报告
func (h *HomeworkHandler) GenerateReport(c *gin.Context) {
	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目ID不能为空"})
		return
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	report, err := h.homeworkService.GenerateReport(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成报告失败"})
		return
	}

	c.JSON(http.StatusOK, report)
}

// CreateStudentBranch 为学生创建个人作业分支
func (h *HomeworkHandler) CreateStudentBranch(c *gin.Context) {
	homeworkIDStr := c.Param("id")
	homeworkID, err := uuid.Parse(homeworkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的作业ID"})
		return
	}

	studentID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	if err := h.homeworkService.CreateStudentBranch(homeworkID, studentID.(uuid.UUID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "个人分支创建成功"})
}

// GetHomeworkBranches 获取作业的所有学生分支
func (h *HomeworkHandler) GetHomeworkBranches(c *gin.Context) {
	homeworkIDStr := c.Param("id")
	homeworkID, err := uuid.Parse(homeworkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的作业ID"})
		return
	}

	accessToken := c.GetHeader("Authorization")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "需要访问令牌"})
		return
	}

	branches, err := h.homeworkService.GetHomeworkBranches(homeworkID, accessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, branches)
}

// SubmitHomeworkToBranch 提交作业到个人分支
func (h *HomeworkHandler) SubmitHomeworkToBranch(c *gin.Context) {
	homeworkIDStr := c.Param("id")
	homeworkID, err := uuid.Parse(homeworkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的作业ID"})
		return
	}

	studentID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	var req struct {
		Content string   `json:"content" binding:"required"`
		Files   []string `json:"files"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	submission, err := h.homeworkService.SubmitHomeworkToBranch(
		homeworkID,
		studentID.(uuid.UUID),
		req.Content,
		req.Files,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, submission)
}

// GetStudentBranchInfo 获取学生分支信息
func (h *HomeworkHandler) GetStudentBranchInfo(c *gin.Context) {
	homeworkIDStr := c.Param("id")
	homeworkID, err := uuid.Parse(homeworkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的作业ID"})
		return
	}

	studentID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	branchInfo, err := h.homeworkService.GetStudentBranchInfo(homeworkID, studentID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, branchInfo)
}
