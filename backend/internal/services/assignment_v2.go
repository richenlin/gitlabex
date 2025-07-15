package services

import (
	"fmt"
	"time"

	"gitlabex/internal/models"

	"gorm.io/gorm"
)

// AssignmentServiceV2 作业管理服务V2
type AssignmentServiceV2 struct {
	db                *gorm.DB
	permissionService *PermissionService
	gitlabService     *GitLabService
	projectService    *ProjectServiceV2
}

// NewAssignmentServiceV2 创建作业管理服务V2
func NewAssignmentServiceV2(db *gorm.DB, permissionService *PermissionService, gitlabService *GitLabService, projectService *ProjectServiceV2) *AssignmentServiceV2 {
	return &AssignmentServiceV2{
		db:                db,
		permissionService: permissionService,
		gitlabService:     gitlabService,
		projectService:    projectService,
	}
}

// CreateAssignmentRequestV2 创建作业请求V2
type CreateAssignmentRequestV2 struct {
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	ProjectID   uint      `json:"project_id" binding:"required"`
	DueDate     time.Time `json:"due_date" binding:"required"`
	Type        string    `json:"type"` // homework, project, quiz

	// GitLab相关字段
	RequiredFiles     []string `json:"required_files"`      // 必须提交的文件列表
	SubmissionBranch  string   `json:"submission_branch"`   // 提交分支前缀
	AutoCreateMR      bool     `json:"auto_create_mr"`      // 是否自动创建合并请求
	RequireCodeReview bool     `json:"require_code_review"` // 是否需要代码审查
	MaxFileSize       int64    `json:"max_file_size"`       // 最大文件大小
	AllowedFileTypes  []string `json:"allowed_file_types"`  // 允许的文件类型
}

// UpdateAssignmentRequestV2 更新作业请求V2
type UpdateAssignmentRequestV2 struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	DueDate     *time.Time `json:"due_date"`
	Type        string     `json:"type"`
	Status      string     `json:"status"`

	// GitLab相关字段
	RequiredFiles     []string `json:"required_files"`
	SubmissionBranch  string   `json:"submission_branch"`
	AutoCreateMR      *bool    `json:"auto_create_mr"`
	RequireCodeReview *bool    `json:"require_code_review"`
	MaxFileSize       *int64   `json:"max_file_size"`
	AllowedFileTypes  []string `json:"allowed_file_types"`
}

// SubmitAssignmentRequestV2 提交作业请求V2
type SubmitAssignmentRequestV2 struct {
	Content        string   `json:"content"`
	CommitHash     string   `json:"commit_hash" binding:"required"`
	CommitMessage  string   `json:"commit_message" binding:"required"`
	BranchName     string   `json:"branch_name" binding:"required"`
	FilesSubmitted []string `json:"files_submitted"`
	FilesSummary   string   `json:"files_summary"`
}

// ReviewAssignmentRequestV2 评审作业请求V2
type ReviewAssignmentRequestV2 struct {
	Score    float64 `json:"score" binding:"required"`
	Feedback string  `json:"feedback"`

	// 详细评审报告
	CodeQuality    float64 `json:"code_quality"`
	Functionality  float64 `json:"functionality"`
	Documentation  float64 `json:"documentation"`
	CodeStyle      float64 `json:"code_style"`
	TestCoverage   float64 `json:"test_coverage"`
	Creativity     float64 `json:"creativity"`
	Suggestions    string  `json:"suggestions"`
	Strengths      string  `json:"strengths"`
	Weaknesses     string  `json:"weaknesses"`
	OverallComment string  `json:"overall_comment"`
}

// AssignmentSimpleV2 简化的作业结构V2
type AssignmentSimpleV2 struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ProjectID   uint      `json:"project_id"`
	ProjectName string    `json:"project_name"`
	TeacherID   uint      `json:"teacher_id"`
	TeacherName string    `json:"teacher_name"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	DueDate     time.Time `json:"due_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 提交统计
	TotalSubmissions    int `json:"total_submissions"`
	ReviewedSubmissions int `json:"reviewed_submissions"`
	PendingSubmissions  int `json:"pending_submissions"`
}

// CreateAssignment 创建作业（仅教师）
func (s *AssignmentServiceV2) CreateAssignment(teacherID uint, req *CreateAssignmentRequestV2) (*models.Assignment, error) {
	// 验证课题权限（必须是课题创建者）
	project, err := s.projectService.GetProjectByID(req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	if project.TeacherID != teacherID {
		return nil, fmt.Errorf("only project creator can create assignments")
	}

	// 设置默认值
	if req.SubmissionBranch == "" {
		req.SubmissionBranch = "assignment-submission"
	}
	if req.MaxFileSize == 0 {
		req.MaxFileSize = 10485760 // 10MB
	}

	// 创建作业记录
	assignment := &models.Assignment{
		Title:       req.Title,
		Description: req.Description,
		ProjectID:   req.ProjectID,
		TeacherID:   teacherID,
		Type:        req.Type,
		Status:      "active",
		DueDate:     req.DueDate,

		RequiredFiles:     req.RequiredFiles,
		SubmissionBranch:  req.SubmissionBranch,
		AutoCreateMR:      req.AutoCreateMR,
		RequireCodeReview: req.RequireCodeReview,
		MaxFileSize:       req.MaxFileSize,
		AllowedFileTypes:  req.AllowedFileTypes,
	}

	// 保存到数据库
	if err := s.db.Create(assignment).Error; err != nil {
		return nil, fmt.Errorf("failed to create assignment: %w", err)
	}

	return assignment, nil
}

// GetAssignmentsByTeacher 获取教师的所有作业（来自教师创建的所有课题）
func (s *AssignmentServiceV2) GetAssignmentsByTeacher(teacherID uint) ([]AssignmentSimpleV2, error) {
	var assignments []models.Assignment

	// 查询教师创建的所有课题的作业
	err := s.db.Preload("Project").Preload("Teacher").
		Joins("JOIN projects ON assignments.project_id = projects.id").
		Where("projects.teacher_id = ?", teacherID).
		Find(&assignments).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get teacher assignments: %w", err)
	}

	return s.convertToSimpleAssignments(assignments), nil
}

// GetAssignmentsByStudent 获取学生的作业（来自学生参与的课题）
func (s *AssignmentServiceV2) GetAssignmentsByStudent(studentID uint) ([]AssignmentSimpleV2, error) {
	var assignments []models.Assignment

	// 查询学生参与的课题的作业
	err := s.db.Preload("Project").Preload("Teacher").
		Joins("JOIN projects ON assignments.project_id = projects.id").
		Joins("JOIN project_members ON projects.id = project_members.project_id").
		Where("project_members.user_id = ? AND project_members.is_active = true", studentID).
		Find(&assignments).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get student assignments: %w", err)
	}

	return s.convertToSimpleAssignments(assignments), nil
}

// SubmitAssignment 提交作业（学生）
func (s *AssignmentServiceV2) SubmitAssignment(studentID uint, assignmentID uint, req *SubmitAssignmentRequestV2) (*models.AssignmentSubmission, error) {
	// 验证作业权限（必须是课题成员）
	assignment, err := s.GetAssignmentByID(assignmentID)
	if err != nil {
		return nil, fmt.Errorf("assignment not found: %w", err)
	}

	// 检查学生是否是课题成员
	var member models.ProjectMember
	err = s.db.Where("project_id = ? AND user_id = ? AND is_active = true",
		assignment.ProjectID, studentID).First(&member).Error
	if err != nil {
		return nil, fmt.Errorf("student is not a member of this project")
	}

	// 检查是否已经提交过
	var existingSubmission models.AssignmentSubmission
	err = s.db.Where("assignment_id = ? AND student_id = ?", assignmentID, studentID).First(&existingSubmission).Error
	if err == nil {
		// 更新现有提交
		existingSubmission.Content = req.Content
		existingSubmission.CommitHash = req.CommitHash
		existingSubmission.CommitMessage = req.CommitMessage
		existingSubmission.BranchName = req.BranchName
		existingSubmission.FilesSubmitted = req.FilesSubmitted
		existingSubmission.FilesSummary = req.FilesSummary
		existingSubmission.Status = "submitted"
		existingSubmission.SubmittedAt = time.Now()

		if err := s.db.Save(&existingSubmission).Error; err != nil {
			return nil, fmt.Errorf("failed to update submission: %w", err)
		}
		return &existingSubmission, nil
	}

	// 创建新提交
	submission := &models.AssignmentSubmission{
		AssignmentID:   assignmentID,
		StudentID:      studentID,
		Content:        req.Content,
		Status:         "submitted",
		CommitHash:     req.CommitHash,
		CommitMessage:  req.CommitMessage,
		BranchName:     req.BranchName,
		FilesSubmitted: req.FilesSubmitted,
		FilesSummary:   req.FilesSummary,
	}

	if err := s.db.Create(submission).Error; err != nil {
		return nil, fmt.Errorf("failed to create submission: %w", err)
	}

	return submission, nil
}

// ReviewAssignment 评审作业（教师）
func (s *AssignmentServiceV2) ReviewAssignment(teacherID uint, submissionID uint, req *ReviewAssignmentRequestV2) (*models.Review, error) {
	// 获取提交记录
	var submission models.AssignmentSubmission
	err := s.db.Preload("Assignment").Preload("Assignment.Project").First(&submission, submissionID).Error
	if err != nil {
		return nil, fmt.Errorf("submission not found: %w", err)
	}

	// 验证权限（必须是课题创建者）
	if submission.Assignment.Project.TeacherID != teacherID {
		return nil, fmt.Errorf("only project creator can review assignments")
	}

	// 检查是否已经评审过
	var existingReview models.Review
	err = s.db.Where("submission_id = ? AND reviewer_id = ?", submissionID, teacherID).First(&existingReview).Error
	if err == nil {
		// 更新现有评审
		existingReview.Score = req.Score
		existingReview.Feedback = req.Feedback
		existingReview.CodeQuality = req.CodeQuality
		existingReview.Functionality = req.Functionality
		existingReview.Documentation = req.Documentation
		existingReview.CodeStyle = req.CodeStyle
		existingReview.TestCoverage = req.TestCoverage
		existingReview.Creativity = req.Creativity
		existingReview.Suggestions = req.Suggestions
		existingReview.Strengths = req.Strengths
		existingReview.Weaknesses = req.Weaknesses
		existingReview.OverallComment = req.OverallComment
		existingReview.Status = "completed"

		if err := s.db.Save(&existingReview).Error; err != nil {
			return nil, fmt.Errorf("failed to update review: %w", err)
		}

		// 更新提交状态
		submission.Status = "graded"
		submission.Score = req.Score
		submission.Feedback = req.Feedback
		now := time.Now()
		submission.GradedAt = &now
		s.db.Save(&submission)

		return &existingReview, nil
	}

	// 创建新评审
	review := &models.Review{
		SubmissionID:   submissionID,
		ReviewerID:     teacherID,
		Score:          req.Score,
		Feedback:       req.Feedback,
		Status:         "completed",
		CodeQuality:    req.CodeQuality,
		Functionality:  req.Functionality,
		Documentation:  req.Documentation,
		CodeStyle:      req.CodeStyle,
		TestCoverage:   req.TestCoverage,
		Creativity:     req.Creativity,
		Suggestions:    req.Suggestions,
		Strengths:      req.Strengths,
		Weaknesses:     req.Weaknesses,
		OverallComment: req.OverallComment,
	}

	if err := s.db.Create(review).Error; err != nil {
		return nil, fmt.Errorf("failed to create review: %w", err)
	}

	// 更新提交状态
	submission.Status = "graded"
	submission.Score = req.Score
	submission.Feedback = req.Feedback
	now := time.Now()
	submission.GradedAt = &now
	s.db.Save(&submission)

	return review, nil
}

// GetAssignmentSubmissions 获取作业的所有提交（教师）
func (s *AssignmentServiceV2) GetAssignmentSubmissions(teacherID uint, assignmentID uint) ([]models.AssignmentSubmission, error) {
	// 验证权限
	assignment, err := s.GetAssignmentByID(assignmentID)
	if err != nil {
		return nil, fmt.Errorf("assignment not found: %w", err)
	}

	if assignment.TeacherID != teacherID {
		return nil, fmt.Errorf("only assignment creator can view submissions")
	}

	var submissions []models.AssignmentSubmission
	err = s.db.Preload("Student").Preload("Reviews").
		Where("assignment_id = ?", assignmentID).
		Order("submitted_at DESC").
		Find(&submissions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get submissions: %w", err)
	}

	return submissions, nil
}

// GetMyAssignmentSubmissions 获取我的作业提交（学生）
func (s *AssignmentServiceV2) GetMyAssignmentSubmissions(studentID uint) ([]models.AssignmentSubmission, error) {
	var submissions []models.AssignmentSubmission
	err := s.db.Preload("Assignment").Preload("Assignment.Project").Preload("Reviews").
		Where("student_id = ?", studentID).
		Order("submitted_at DESC").
		Find(&submissions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get my submissions: %w", err)
	}

	return submissions, nil
}

// GetAssignmentByID 根据ID获取作业详情
func (s *AssignmentServiceV2) GetAssignmentByID(assignmentID uint) (*models.Assignment, error) {
	var assignment models.Assignment
	err := s.db.Preload("Project").Preload("Teacher").First(&assignment, assignmentID).Error
	if err != nil {
		return nil, fmt.Errorf("assignment not found: %w", err)
	}
	return &assignment, nil
}

// UpdateAssignment 更新作业信息（教师）
func (s *AssignmentServiceV2) UpdateAssignment(teacherID uint, assignmentID uint, req *UpdateAssignmentRequestV2) (*models.Assignment, error) {
	assignment, err := s.GetAssignmentByID(assignmentID)
	if err != nil {
		return nil, fmt.Errorf("assignment not found: %w", err)
	}

	// 验证权限
	if assignment.TeacherID != teacherID {
		return nil, fmt.Errorf("only assignment creator can update assignment")
	}

	// 更新字段
	if req.Title != "" {
		assignment.Title = req.Title
	}
	if req.Description != "" {
		assignment.Description = req.Description
	}
	if req.DueDate != nil {
		assignment.DueDate = *req.DueDate
	}
	if req.Type != "" {
		assignment.Type = req.Type
	}
	if req.Status != "" {
		assignment.Status = req.Status
	}
	if req.SubmissionBranch != "" {
		assignment.SubmissionBranch = req.SubmissionBranch
	}
	if req.AutoCreateMR != nil {
		assignment.AutoCreateMR = *req.AutoCreateMR
	}
	if req.RequireCodeReview != nil {
		assignment.RequireCodeReview = *req.RequireCodeReview
	}
	if req.MaxFileSize != nil {
		assignment.MaxFileSize = *req.MaxFileSize
	}
	if req.RequiredFiles != nil {
		assignment.RequiredFiles = req.RequiredFiles
	}
	if req.AllowedFileTypes != nil {
		assignment.AllowedFileTypes = req.AllowedFileTypes
	}

	if err := s.db.Save(assignment).Error; err != nil {
		return nil, fmt.Errorf("failed to update assignment: %w", err)
	}

	return assignment, nil
}

// DeleteAssignment 删除作业（教师）
func (s *AssignmentServiceV2) DeleteAssignment(teacherID uint, assignmentID uint) error {
	assignment, err := s.GetAssignmentByID(assignmentID)
	if err != nil {
		return fmt.Errorf("assignment not found: %w", err)
	}

	// 验证权限
	if assignment.TeacherID != teacherID {
		return fmt.Errorf("only assignment creator can delete assignment")
	}

	// 检查是否有提交
	var submissionCount int64
	if err := s.db.Model(&models.AssignmentSubmission{}).Where("assignment_id = ?", assignmentID).Count(&submissionCount).Error; err != nil {
		return fmt.Errorf("failed to check submissions: %w", err)
	}

	if submissionCount > 0 {
		return fmt.Errorf("cannot delete assignment with submissions")
	}

	// 删除作业
	if err := s.db.Delete(&models.Assignment{}, assignmentID).Error; err != nil {
		return fmt.Errorf("failed to delete assignment: %w", err)
	}

	return nil
}

// convertToSimpleAssignments 转换为简化的作业结构
func (s *AssignmentServiceV2) convertToSimpleAssignments(assignments []models.Assignment) []AssignmentSimpleV2 {
	simpleAssignments := make([]AssignmentSimpleV2, len(assignments))

	for i, assignment := range assignments {
		// 统计提交情况
		var totalSubmissions int64
		var reviewedSubmissions int64

		s.db.Model(&models.AssignmentSubmission{}).Where("assignment_id = ?", assignment.ID).Count(&totalSubmissions)
		s.db.Model(&models.AssignmentSubmission{}).Where("assignment_id = ? AND status = 'graded'", assignment.ID).Count(&reviewedSubmissions)

		simpleAssignments[i] = AssignmentSimpleV2{
			ID:          assignment.ID,
			Title:       assignment.Title,
			Description: assignment.Description,
			ProjectID:   assignment.ProjectID,
			ProjectName: assignment.Project.Name,
			TeacherID:   assignment.TeacherID,
			TeacherName: assignment.Teacher.Name,
			Type:        assignment.Type,
			Status:      assignment.Status,
			DueDate:     assignment.DueDate,
			CreatedAt:   assignment.CreatedAt,
			UpdatedAt:   assignment.UpdatedAt,

			TotalSubmissions:    int(totalSubmissions),
			ReviewedSubmissions: int(reviewedSubmissions),
			PendingSubmissions:  int(totalSubmissions - reviewedSubmissions),
		}
	}

	return simpleAssignments
}
