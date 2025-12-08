package handlers

import (
	"fmt"
	"gitlabex/internal/dto"
	"gitlabex/internal/models"
	"gitlabex/internal/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// HomeworkHandler 作业处理器
type HomeworkHandler struct {
	homeworkService *services.HomeworkService
	userService     *services.UserService
	gitlabService   *services.GitLabService
	researchService *services.ResearchService
}

// NewHomeworkHandler 创建作业处理器
func NewHomeworkHandler(homeworkService *services.HomeworkService, userService *services.UserService, gitlabService *services.GitLabService, researchService *services.ResearchService) *HomeworkHandler {
	return &HomeworkHandler{
		homeworkService: homeworkService,
		userService:     userService,
		gitlabService:   gitlabService,
		researchService: researchService,
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

	// 安全地获取用户信息
	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	userID, ok := gitlabUserID.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的用户ID"})
		return
	}

	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少访问令牌"})
		return
	}

	token, ok := accessToken.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的访问令牌"})
		return
	}

	// 检查管理员权限
	isAdmin, _ := c.Get("is_admin")
	userIsAdmin := false
	if isAdmin != nil {
		if adminBool, ok := isAdmin.(bool); ok {
			userIsAdmin = adminBool
		}
	}

	// 权限检查：只有管理员和教师可以创建作业
	if !userIsAdmin {
		// 获取项目信息
		project, err := h.researchService.GetResearchProjectByID(projectID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "课题不存在"})
			return
		}

		// 检查是否为课题创建者
		isCreator := project.CreatorID == userID

		// 检查GitLab项目权限
		hasPermission := false
		if project.GitLabProjectID != nil {
			accessLevel, err := h.gitlabService.GetUserProjectAccessLevel(token, *project.GitLabProjectID)
			// Maintainer (40) 或 Owner (50) 可以创建作业
			if err == nil && accessLevel >= 40 {
				hasPermission = true
			}
		}

		if !isCreator && !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "权限不足",
				"message": "只有管理员和教师可以创建作业",
			})
			return
		}
	}

	homework := &models.Homework{
		ProjectID:   projectID,
		Title:       req.Title,
		Description: req.Description,
		DueDate:     req.DueDate,
		MaxGrade:    req.MaxGrade,
		CreatorID:   userID,
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

	// 获取实时统计数据
	stats, err := h.homeworkService.GetHomeworkDetailStats(homeworkID)
	if err != nil {
		// 如果统计失败，使用空数据
		stats = map[string]interface{}{
			"submission_count": 0,
			"graded_count":     0,
			"average_grade":    0.0,
		}
	}

	// 安全地获取用户ID（可能是游客访问）
	_, hasUser := getInt64FromContext(c, "gitlab_user_id")

	// 构建返回数据，包含权限信息和统计数据
	response := gin.H{
		"homework": homework,
		"stats":    stats, // 添加实时统计数据
		"permissions": gin.H{
			"can_edit":   false,
			"can_delete": false,
			"can_grade":  false,
			"can_submit": false,
		},
	}

	// 如果用户已登录，计算权限
	if hasUser {
		userID, _ := getInt64FromContext(c, "gitlab_user_id")
		isAdmin := getBoolFromContext(c, "is_admin")

		// 检查是否可以管理作业（编辑、删除）
		canManage := h.checkHomeworkManagePermission(c, homework.ProjectID)
		response["permissions"].(gin.H)["can_edit"] = canManage
		response["permissions"].(gin.H)["can_delete"] = canManage

		// 检查是否可以批改作业
		canGrade := h.checkHomeworkGradePermission(c, homework.ProjectID)
		response["permissions"].(gin.H)["can_grade"] = canGrade
		fmt.Printf("[权限调试] 批改权限: canGrade=%v\n", canGrade)

		// 检查是否可以提交作业
		// 规则：
		// 1. 管理员和教师(有管理/批改权限)不需要提交作业
		// 2. 其他登录用户(课题成员)可以提交作业
		canSubmit := false

		// 调试日志
		fmt.Printf("[权限调试] 作业ID=%s, 用户ID=%d, isAdmin=%v, canManage=%v, canGrade=%v\n",
			homework.ID, userID, isAdmin, canManage, canGrade)

		// 如果不是管理员且没有管理权限，则可以提交
		if !isAdmin && !canManage {
			// 默认允许提交(对于有访问权限的成员)
			canSubmit = true
			fmt.Printf("[权限调试] 初始允许提交: canSubmit=true\n")

			// 如果关联了GitLab项目，进一步验证访问级别
			if homework.ProjectID != uuid.Nil {
				project, err := h.researchService.GetResearchProjectByID(homework.ProjectID)
				if err == nil && project.GitLabProjectID != nil {
					accessToken, ok := getStringFromContext(c, "gitlab_access_token")
					if ok {
						accessLevel, err := h.gitlabService.GetUserProjectAccessLevel(accessToken, *project.GitLabProjectID)
						fmt.Printf("[权限调试] GitLab项目ID=%d, accessLevel=%d, err=%v\n",
							*project.GitLabProjectID, accessLevel, err)

						if err == nil {
							// Reporter(20)和Developer(30)可以提交
							// Maintainer(40+)是教师，已通过canManage过滤
							// Guest(10)及以下不能提交
							canSubmit = accessLevel >= 20
							fmt.Printf("[权限调试] 根据访问级别判断: accessLevel=%d, canSubmit=%v\n",
								accessLevel, canSubmit)
						} else {
							// 如果获取访问级别失败，暂时允许(可能是未关联GitLab的作业)
							canSubmit = true
							fmt.Printf("[权限调试] 获取访问级别失败，默认允许: canSubmit=true\n")
						}
					} else {
						fmt.Printf("[权限调试] 无法获取访问令牌\n")
					}
				} else {
					fmt.Printf("[权限调试] 项目信息: err=%v, hasGitLabID=%v\n",
						err, project != nil && project.GitLabProjectID != nil)
				}
			} else {
				fmt.Printf("[权限调试] 作业未关联项目，默认允许提交\n")
			}
		} else {
			fmt.Printf("[权限调试] 管理员或有管理权限，不允许提交: isAdmin=%v, canManage=%v\n",
				isAdmin, canManage)
		}

		fmt.Printf("[权限调试] 最终结果: canSubmit=%v\n", canSubmit)
		response["permissions"].(gin.H)["can_submit"] = canSubmit
	}

	c.JSON(http.StatusOK, response)
}

// GetHomeworkByProject 获取项目下的所有作业
func (h *HomeworkHandler) GetHomeworkByProject(c *gin.Context) {
	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		// 尝试使用 projectId 参数（前端可能使用驼峰命名）
		projectIDStr = c.Query("projectId")
	}
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

	// 获取作业信息
	homework, err := h.homeworkService.GetHomeworkByID(homeworkID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "作业不存在"})
		return
	}

	// 权限检查：只有管理员和教师可以编辑作业
	if !h.checkHomeworkManagePermission(c, homework.ProjectID) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "权限不足",
			"message": "只有管理员和教师可以编辑作业",
		})
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

	// 获取作业信息
	homework, err := h.homeworkService.GetHomeworkByID(homeworkID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "作业不存在"})
		return
	}

	// 权限检查：只有管理员和教师可以删除作业
	if !h.checkHomeworkManagePermission(c, homework.ProjectID) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "权限不足",
			"message": "只有管理员和教师可以删除作业",
		})
		return
	}

	if err := h.homeworkService.DeleteHomework(homeworkID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除作业失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "作业删除成功"})
}

// checkHomeworkManagePermission 检查作业管理权限（创建、编辑、删除）
// 只有管理员和教师（Maintainer级别以上）可以管理作业
func (h *HomeworkHandler) checkHomeworkManagePermission(c *gin.Context, projectID uuid.UUID) bool {
	// 安全地获取用户信息
	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		return false
	}

	userID, ok := gitlabUserID.(int64)
	if !ok {
		return false
	}

	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		return false
	}

	token, ok := accessToken.(string)
	if !ok {
		return false
	}

	// 检查管理员权限
	isAdmin, _ := c.Get("is_admin")
	if isAdmin != nil {
		if adminBool, ok := isAdmin.(bool); ok && adminBool {
			return true
		}
	}

	// 获取项目信息
	project, err := h.researchService.GetResearchProjectByID(projectID)
	if err != nil {
		return false
	}

	// 检查是否为课题创建者
	if project.CreatorID == userID {
		return true
	}

	// 检查GitLab项目权限
	if project.GitLabProjectID != nil {
		accessLevel, err := h.gitlabService.GetUserProjectAccessLevel(token, *project.GitLabProjectID)
		// Maintainer (40) 或 Owner (50) 可以管理作业
		if err == nil && accessLevel >= 40 {
			return true
		}
	}

	return false
}

// checkHomeworkGradePermission 检查作业批改权限
// 管理员、教师（Maintainer）和研究员（Developer）可以批改作业
func (h *HomeworkHandler) checkHomeworkGradePermission(c *gin.Context, projectID uuid.UUID) bool {
	// 安全地获取用户信息
	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		return false
	}

	userID, ok := gitlabUserID.(int64)
	if !ok {
		return false
	}

	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		return false
	}

	token, ok := accessToken.(string)
	if !ok {
		return false
	}

	// 检查管理员权限
	isAdmin, _ := c.Get("is_admin")
	if isAdmin != nil {
		if adminBool, ok := isAdmin.(bool); ok && adminBool {
			return true
		}
	}

	// 获取项目信息
	project, err := h.researchService.GetResearchProjectByID(projectID)
	if err != nil {
		return false
	}

	// 检查是否为课题创建者
	if project.CreatorID == userID {
		return true
	}

	// 检查GitLab项目权限
	if project.GitLabProjectID != nil {
		accessLevel, err := h.gitlabService.GetUserProjectAccessLevel(token, *project.GitLabProjectID)
		fmt.Printf("[批改权限调试] 用户ID=%d, GitLab项目ID=%d, accessLevel=%d, err=%v\n",
			userID, *project.GitLabProjectID, accessLevel, err)
		// Developer (30)、Maintainer (40) 或 Owner (50) 可以批改作业
		// 研究员和教师都可以批改
		if err == nil && accessLevel >= 30 {
			fmt.Printf("[批改权限调试] 允许批改: accessLevel=%d >= 30 (研究员/教师)\n", accessLevel)
			return true
		}
		fmt.Printf("[批改权限调试] 不允许批改: accessLevel=%d < 30 (普通学生)\n", accessLevel)
	}

	return false
}

// SubmitHomework 创建学生作业分支并返回Web IDE链接
func (h *HomeworkHandler) SubmitHomework(c *gin.Context) {
	homeworkIDStr := c.Param("id")
	homeworkID, err := uuid.Parse(homeworkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的作业ID"})
		return
	}

	// 获取作业信息
	homework, err := h.homeworkService.GetHomeworkByID(homeworkID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "作业不存在"})
		return
	}

	// 检查作业状态
	if homework.Status != "active" && homework.Status != "published" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "作业未开放"})
		return
	}

	// 检查截止日期
	if homework.DueDate != nil && time.Now().After(*homework.DueDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "作业已过截止日期"})
		return
	}

	userID, exists := getInt64FromContext(c, "gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	accessToken, exists := getStringFromContext(c, "gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无法获取访问令牌"})
		return
	}

	// 获取用户信息
	user, err := h.gitlabService.GetUser(accessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	username := user.Username

	// 获取项目信息
	project, err := h.researchService.GetResearchProjectByID(homework.ProjectID)
	if err != nil || project.GitLabProjectID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "作业未关联GitLab项目"})
		return
	}

	gitlabProjectID := *project.GitLabProjectID

	// 获取系统token
	systemToken := h.gitlabService.GetSystemToken()
	if systemToken == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统配置错误"})
		return
	}

	// 获取默认分支
	gitlabProject, err := h.gitlabService.GetProject(systemToken, gitlabProjectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取项目信息失败"})
		return
	}

	defaultBranch := gitlabProject.DefaultBranch
	if defaultBranch == "" {
		defaultBranch = "main"
	}

	// 创建学生专属分支
	branchName := fmt.Sprintf("homework-%s-%s", homeworkID.String()[:8], username)
	submissionPath := fmt.Sprintf("submissions/%s", username)

	// 检查分支是否已存在
	branches, err := h.gitlabService.GetBranches(systemToken, gitlabProjectID)
	branchExists := false
	if err == nil {
		for _, b := range branches {
			if b.Name == branchName {
				branchExists = true
				break
			}
		}
	}

	// 如果分支不存在，创建它
	if !branchExists {
		createReq := &dto.CreateBranchRequest{
			Branch: branchName,
			Ref:    defaultBranch,
		}
		if _, err := h.gitlabService.CreateBranch(systemToken, gitlabProjectID, createReq); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建分支失败: " + err.Error()})
			return
		}

		// 创建初始README文件
		readmePath := fmt.Sprintf("%s/README.md", submissionPath)
		readmeContent := fmt.Sprintf("# %s - 作业提交\n\n**学生**: %s (%s)\n**作业标题**: %s\n**创建时间**: %s\n\n---\n\n## 说明\n\n请在此分支提交你的作业代码。\n\n## 提交方式\n\n在GitLab Web IDE中直接编辑和提交代码即可。\n\n## 常见问题\n\n如果无法在Web IDE中创建文件，请尝试：\n1. 刷新页面\n2. 使用GitLab界面上的 '+ New file' 按钮\n3. 或者先在本地克隆分支，然后推送代码\n",
			homework.Title, user.Name, username, homework.Title, time.Now().Format("2006-01-02 15:04:05"))

		readmeReq := &dto.CreateFileRequest{
			Branch:        branchName,
			Content:       readmeContent,
			CommitMessage: fmt.Sprintf("初始化作业分支: %s", homework.Title),
		}

		// 使用系统token创建README（因为是新分支，学生还没有提交过）
		if err := h.gitlabService.CreateFile(systemToken, gitlabProjectID, readmePath, readmeReq); err != nil {
			fmt.Printf("创建初始README失败: %v\n", err)
		}

		fmt.Printf("分支 %s 创建成功，学生 %s 可以推送代码\n", branchName, username)
	}

	// 保存或更新提交记录
	submission, err := h.homeworkService.GetUserSubmissionForHomework(homeworkID, userID)
	if err != nil || submission == nil {
		// 首次提交，创建记录
		submission = &models.Submission{
			HomeworkID:   homeworkID,
			StudentID:    userID,
			GitLabBranch: branchName,
			Status:       "pending",
		}
		if err := h.homeworkService.SubmitHomework(submission); err != nil {
			fmt.Printf("创建提交记录失败: %v\n", err)
		}
	} else {
		// 已存在提交记录，确保分支名正确（SubmitHomework 方法会处理更新）
		if submission.GitLabBranch != branchName {
			submission.GitLabBranch = branchName
			// SubmitHomework 会检测已存在的记录并更新
			if err := h.homeworkService.SubmitHomework(submission); err != nil {
				fmt.Printf("更新提交记录的分支名失败: %v\n", err)
			} else {
				fmt.Printf("已更新提交记录的分支名: %s -> %s\n", submission.GitLabBranch, branchName)
			}
		}
	}

	// 返回Web IDE链接（使用项目路径）
	projectPath := gitlabProject.PathWithNamespace
	webIDEURL := h.gitlabService.GetWebIDEURL(projectPath, branchName, submissionPath)

	c.JSON(http.StatusOK, gin.H{
		"branch":       branchName,
		"path":         submissionPath,
		"web_ide_url":  webIDEURL,
		"project_path": projectPath,
		"message":      fmt.Sprintf("分支已准备好，请在在线编辑器中提交代码"),
	})
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

	// 获取提交信息
	submission, err := h.homeworkService.GetSubmissionByID(submissionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "提交不存在"})
		return
	}

	// 获取作业信息
	homework, err := h.homeworkService.GetHomeworkByID(submission.HomeworkID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "作业不存在"})
		return
	}

	// 权限检查：教师和研究员可以批改作业
	if !h.checkHomeworkGradePermission(c, homework.ProjectID) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "权限不足",
			"message": "只有管理员、教师和研究员可以批改作业",
		})
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

	// 安全地获取批改者ID
	graderID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	userID, ok := graderID.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的用户ID"})
		return
	}

	if err := h.homeworkService.GradeHomework(submissionID, req.Grade, req.Feedback, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "评分失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "评分成功"})
}

// GetUserSubmissions 获取用户的作业提交
func (h *HomeworkHandler) GetUserSubmissions(c *gin.Context) {
	userID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	submissions, err := h.homeworkService.GetUserSubmissions(userID.(int64))
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

	userID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	submission, err := h.homeworkService.GetUserSubmissionForHomework(homeworkID, userID.(int64))
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

	// 获取访问令牌以调用GitLab API
	accessToken, exists := getStringFromContext(c, "gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少访问令牌"})
		return
	}

	// 获取成绩数据
	grades, err := h.homeworkService.ExportGrades(homeworkID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导出成绩失败"})
		return
	}

	// 补充学生信息(从GitLab API获取)
	for i, grade := range grades {
		studentID, ok := grade["student_id"].(int64)
		if !ok {
			continue
		}

		// 从GitLab获取用户信息
		user, err := h.gitlabService.GetUserByID(accessToken, studentID)
		if err == nil {
			grades[i]["student_name"] = user.Name
			grades[i]["student_username"] = user.Username
			grades[i]["student_email"] = user.Email
		} else {
			// 如果获取失败,保留student_id供前端处理
			grades[i]["student_name"] = fmt.Sprintf("User_%d", studentID)
			grades[i]["student_username"] = fmt.Sprintf("user_%d", studentID)
			grades[i]["student_email"] = ""
		}
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

	graderID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	submissions, err := h.homeworkService.GetPendingReviews(projectID, graderID.(int64))
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

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
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

	var creatorID int64
	if !isPublic {
		userID, exists := c.Get("gitlab_user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
			return
		}
		creatorID = userID.(int64)
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

	userID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	template := &models.AssignmentTemplate{
		Title:       req.Title,
		Description: req.Description,
		DueDate:     &req.DueDate,
		MaxGrade:    req.MaxGrade,
		CreatorID:   userID.(int64),
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

	gitlabUserID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	homework, err := h.homeworkService.UseAssignmentTemplate(templateID, projectID, gitlabUserID.(int64))
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

	gitlabUserID, exists := c.Get("gitlab_user_id")
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
			CreatorID:   gitlabUserID.(int64),
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

	studentID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	// 获取访问令牌
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少访问令牌"})
		return
	}

	if err := h.homeworkService.CreateStudentBranch(homeworkID, studentID.(int64), accessToken.(string)); err != nil {
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

	// 获取访问令牌
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少访问令牌"})
		return
	}

	branches, err := h.homeworkService.GetHomeworkBranches(homeworkID, accessToken.(string))
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

	studentID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	// 获取访问令牌
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少访问令牌"})
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
		studentID.(int64),
		req.Content,
		req.Files,
		accessToken.(string),
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

	studentID, exists := c.Get("gitlab_user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	// 获取访问令牌
	accessToken, exists := c.Get("gitlab_access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少访问令牌"})
		return
	}

	branchInfo, err := h.homeworkService.GetStudentBranchInfo(homeworkID, studentID.(int64), accessToken.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, branchInfo)
}

// GetSubmissionViewURL 获取学生作业提交的GitLab Web IDE链接
func (h *HomeworkHandler) GetSubmissionViewURL(c *gin.Context) {
	submissionIDStr := c.Param("submissionId")
	submissionID, err := uuid.Parse(submissionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的提交ID"})
		return
	}

	// 获取提交信息
	submission, err := h.homeworkService.GetSubmissionByID(submissionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "提交不存在"})
		return
	}

	// 获取作业信息
	homework, err := h.homeworkService.GetHomeworkByID(submission.HomeworkID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "作业不存在"})
		return
	}

	// 获取项目信息
	project, err := h.researchService.GetResearchProjectByID(homework.ProjectID)
	if err != nil || project.GitLabProjectID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "作业未关联GitLab项目"})
		return
	}

	// 获取学生用户信息
	accessToken, _ := getStringFromContext(c, "gitlab_access_token")
	user, err := h.gitlabService.GetUserByID(accessToken, submission.StudentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	username := user.Username

	// 获取GitLab项目详情以获取项目路径
	systemToken := h.gitlabService.GetSystemToken()
	gitlabProject, err := h.gitlabService.GetProject(systemToken, *project.GitLabProjectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取GitLab项目信息失败"})
		return
	}

	// 构建提交路径（与提交时保持一致）
	submissionPath := fmt.Sprintf("submissions/%s", username)

	// 获取或构建分支名称
	branchName := submission.GitLabBranch
	if branchName == "" {
		// 如果数据库中没有分支名，使用与提交时相同的命名规则
		branchName = fmt.Sprintf("homework-%s-%s", submission.HomeworkID.String()[:8], username)
		fmt.Printf("警告: 提交记录中分支名为空，使用构建的分支名: %s\n", branchName)
	}

	fmt.Printf("查看作业 - 学生: %s, 分支: %s, 路径: %s\n", username, branchName, submissionPath)

	// 生成Web IDE链接（使用项目路径）
	projectPath := gitlabProject.PathWithNamespace
	webIDEURL := h.gitlabService.GetWebIDEURL(projectPath, branchName, submissionPath)

	c.JSON(http.StatusOK, gin.H{
		"view_url":     webIDEURL,
		"branch":       branchName,
		"path":         submissionPath,
		"project_path": projectPath,
		"student_name": user.Name,
	})
}
