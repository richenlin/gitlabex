package services

import (
	"fmt"
	"time"

	"gitlabex/internal/models"

	"gorm.io/gorm"
)

// AssignmentService 作业管理服务
type AssignmentService struct {
	db                *gorm.DB
	permissionService *PermissionService
	gitlabService     *GitLabService
	projectService    *ProjectService
}

// NewAssignmentService 创建作业管理服务
func NewAssignmentService(db *gorm.DB, permissionService *PermissionService, gitlabService *GitLabService, projectService *ProjectService) *AssignmentService {
	return &AssignmentService{
		db:                db,
		permissionService: permissionService,
		gitlabService:     gitlabService,
		projectService:    projectService,
	}
}

// CreateAssignmentRequest 创建作业请求
type CreateAssignmentRequest struct {
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

// UpdateAssignmentRequest 更新作业请求
type UpdateAssignmentRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	DueDate     *time.Time `json:"due_date"`
	Status      string     `json:"status"`
}

// SubmitAssignmentRequest 提交作业请求
type SubmitAssignmentRequest struct {
	Content      string            `json:"content"`
	Files        map[string]string `json:"files"`          // 文件路径 -> 文件内容
	FilePaths    []string          `json:"file_paths"`     // 本地文件路径
	AutoCreateMR bool              `json:"auto_create_mr"` // 是否自动创建MR
}

// CreateAssignment 创建作业（老师权限）
func (s *AssignmentService) CreateAssignment(teacherID uint, req *CreateAssignmentRequest) (*models.Assignment, error) {
	// 验证课题是否存在且属于该老师
	var project models.Project
	if err := s.db.Where("id = ? AND teacher_id = ?", req.ProjectID, teacherID).First(&project).Error; err != nil {
		return nil, fmt.Errorf("project not found or access denied: %w", err)
	}

	// 设置默认值
	if req.Type == "" {
		req.Type = "homework"
	}
	if req.MaxFileSize == 0 {
		req.MaxFileSize = 10485760 // 10MB
	}
	if req.SubmissionBranch == "" {
		req.SubmissionBranch = "assignment"
	}

	assignment := &models.Assignment{
		Title:       req.Title,
		Description: req.Description,
		ProjectID:   req.ProjectID,
		TeacherID:   teacherID,
		DueDate:     req.DueDate,
		Type:        req.Type,
		Status:      "active",
		// GitLab相关字段
		RequiredFiles:     req.RequiredFiles,
		SubmissionBranch:  req.SubmissionBranch,
		AutoCreateMR:      req.AutoCreateMR,
		RequireCodeReview: req.RequireCodeReview,
		MaxFileSize:       req.MaxFileSize,
		AllowedFileTypes:  req.AllowedFileTypes,
		MRTitle:           fmt.Sprintf("Assignment: %s", req.Title),
		MRDescription:     fmt.Sprintf("Assignment submission for: %s\n\n%s", req.Title, req.Description),
	}

	if err := s.db.Create(assignment).Error; err != nil {
		return nil, fmt.Errorf("failed to create assignment: %w", err)
	}

	// 预加载关联数据
	if err := s.db.Preload("Teacher").Preload("Project").First(assignment, assignment.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load assignment: %w", err)
	}

	// TODO: 发送通知给课题成员（待通知服务实现）

	return assignment, nil
}

// GetAssignmentByID 根据ID获取作业
func (s *AssignmentService) GetAssignmentByID(assignmentID uint) (*models.Assignment, error) {
	var assignment models.Assignment
	err := s.db.Preload("Teacher").
		Preload("Project").
		Preload("Submissions").
		First(&assignment, assignmentID).Error

	if err != nil {
		return nil, fmt.Errorf("assignment not found: %w", err)
	}

	return &assignment, nil
}

// GetAssignmentsByProject 获取课题的作业列表
func (s *AssignmentService) GetAssignmentsByProject(projectID uint) ([]models.Assignment, error) {
	var assignments []models.Assignment
	err := s.db.Preload("Teacher").
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		Find(&assignments).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get assignments: %w", err)
	}

	return assignments, nil
}

// GetAssignmentsByTeacher 获取老师创建的作业列表
func (s *AssignmentService) GetAssignmentsByTeacher(teacherID uint) ([]models.Assignment, error) {
	var assignments []models.Assignment
	err := s.db.Preload("Teacher").
		Preload("Project").
		Where("teacher_id = ?", teacherID).
		Order("created_at DESC").
		Find(&assignments).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get assignments: %w", err)
	}

	return assignments, nil
}

// GetAssignmentsByStudent 获取学生的作业列表
func (s *AssignmentService) GetAssignmentsByStudent(studentID uint) ([]models.Assignment, error) {
	var assignments []models.Assignment
	err := s.db.Preload("Teacher").
		Preload("Project").
		Joins("JOIN project_members ON assignments.project_id = project_members.project_id").
		Where("project_members.student_id = ? AND project_members.status = 'active'", studentID).
		Order("assignments.created_at DESC").
		Find(&assignments).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get assignments: %w", err)
	}

	return assignments, nil
}

// GetAllAssignments 获取所有作业（管理员权限）
func (s *AssignmentService) GetAllAssignments(page, pageSize int) ([]models.Assignment, int64, error) {
	var assignments []models.Assignment
	var total int64

	// 计算总数
	if err := s.db.Model(&models.Assignment{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count assignments: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := s.db.Preload("Teacher").
		Preload("Project").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&assignments).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get assignments: %w", err)
	}

	return assignments, total, nil
}

// UpdateAssignment 更新作业信息
func (s *AssignmentService) UpdateAssignment(assignmentID uint, req *UpdateAssignmentRequest) (*models.Assignment, error) {
	var assignment models.Assignment
	if err := s.db.First(&assignment, assignmentID).Error; err != nil {
		return nil, fmt.Errorf("assignment not found: %w", err)
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
	if req.Status != "" {
		assignment.Status = req.Status
	}

	if err := s.db.Save(&assignment).Error; err != nil {
		return nil, fmt.Errorf("failed to update assignment: %w", err)
	}

	// 重新加载数据
	if err := s.db.Preload("Teacher").Preload("Project").First(&assignment, assignment.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload assignment: %w", err)
	}

	return &assignment, nil
}

// DeleteAssignment 删除作业
func (s *AssignmentService) DeleteAssignment(assignmentID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除评审
		if err := tx.Where("submission_id IN (SELECT id FROM assignment_submissions WHERE assignment_id = ?)", assignmentID).Delete(&models.Review{}).Error; err != nil {
			return fmt.Errorf("failed to delete reviews: %w", err)
		}

		// 删除作业提交
		if err := tx.Where("assignment_id = ?", assignmentID).Delete(&models.AssignmentSubmission{}).Error; err != nil {
			return fmt.Errorf("failed to delete submissions: %w", err)
		}

		// 删除作业
		if err := tx.Delete(&models.Assignment{}, assignmentID).Error; err != nil {
			return fmt.Errorf("failed to delete assignment: %w", err)
		}

		return nil
	})
}

// SubmitAssignment 提交作业（支持GitLab集成）
func (s *AssignmentService) SubmitAssignment(studentID uint, assignmentID uint, req *SubmitAssignmentRequest) (*models.AssignmentSubmission, error) {
	// 验证作业是否存在且学生有权限提交
	var assignment models.Assignment
	if err := s.db.Preload("Project").First(&assignment, assignmentID).Error; err != nil {
		return nil, fmt.Errorf("assignment not found: %w", err)
	}

	// 验证学生是否在课题中
	var member models.ProjectMember
	if err := s.db.Where("project_id = ? AND student_id = ? AND status = 'active'", assignment.ProjectID, studentID).First(&member).Error; err != nil {
		return nil, fmt.Errorf("student not in project or inactive: %w", err)
	}

	// 检查是否已经提交过
	var existingSubmission models.AssignmentSubmission
	if err := s.db.Where("assignment_id = ? AND student_id = ?", assignmentID, studentID).First(&existingSubmission).Error; err == nil {
		return nil, fmt.Errorf("assignment already submitted")
	}

	// 检查截止日期
	if time.Now().After(assignment.DueDate) {
		return nil, fmt.Errorf("assignment deadline has passed")
	}

	// 如果有文件需要提交到GitLab
	if len(req.Files) > 0 && assignment.Project.GitLabProjectID > 0 {
		// 使用ProjectService的GitLab集成功能
		submission, err := s.projectService.SubmitAssignmentToGitLab(assignment.ProjectID, studentID, assignmentID, req.Files)
		if err != nil {
			return nil, fmt.Errorf("failed to submit to GitLab: %w", err)
		}

		// 如果需要自动创建MR
		if req.AutoCreateMR || assignment.AutoCreateMR {
			mrTitle := fmt.Sprintf("Assignment: %s - %s", assignment.Title, member.Student.Username)
			mrDescription := fmt.Sprintf("Assignment submission for: %s\n\nSubmitted by: %s\nSubmission time: %s",
				assignment.Title, member.Student.Username, time.Now().Format("2006-01-02 15:04:05"))

			mr, err := s.gitlabService.CreateMergeRequestForAssignment(
				assignment.Project.GitLabProjectID,
				member.PersonalBranch,
				mrTitle,
				mrDescription,
				assignment.Project.Teacher.GitLabID,
			)
			if err != nil {
				// MR创建失败不影响作业提交
				fmt.Printf("Warning: Failed to create MR: %v\n", err)
			} else {
				submission.MergeRequestID = mr.IID
				submission.MergeRequestURL = mr.WebURL
				s.db.Save(submission)
			}
		}

		return submission, nil
	} else {
		// 传统方式提交（不使用GitLab）
		submission := &models.AssignmentSubmission{
			AssignmentID: assignmentID,
			StudentID:    studentID,
			Content:      req.Content,
			Status:       "submitted",
			SubmittedAt:  time.Now(),
		}

		if err := s.db.Create(submission).Error; err != nil {
			return nil, fmt.Errorf("failed to create submission: %w", err)
		}

		return submission, nil
	}
}

// GetSubmissionsByAssignment 获取作业的所有提交
func (s *AssignmentService) GetSubmissionsByAssignment(assignmentID uint) ([]models.AssignmentSubmission, error) {
	var submissions []models.AssignmentSubmission
	err := s.db.Preload("Student").
		Preload("Review").
		Where("assignment_id = ?", assignmentID).
		Order("submitted_at DESC").
		Find(&submissions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get submissions: %w", err)
	}

	return submissions, nil
}

// GetSubmissionsByStudent 获取学生的所有提交
func (s *AssignmentService) GetSubmissionsByStudent(studentID uint) ([]models.AssignmentSubmission, error) {
	var submissions []models.AssignmentSubmission
	err := s.db.Preload("Assignment").
		Preload("Review").
		Where("student_id = ?", studentID).
		Order("submitted_at DESC").
		Find(&submissions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get submissions: %w", err)
	}

	return submissions, nil
}

// GetSubmissionByID 根据ID获取作业提交
func (s *AssignmentService) GetSubmissionByID(submissionID uint) (*models.AssignmentSubmission, error) {
	var submission models.AssignmentSubmission
	err := s.db.Preload("Assignment").
		Preload("Student").
		Preload("Review").
		First(&submission, submissionID).Error

	if err != nil {
		return nil, fmt.Errorf("submission not found: %w", err)
	}

	return &submission, nil
}

// GetSubmissionWithGitLabInfo 获取包含GitLab信息的作业提交
func (s *AssignmentService) GetSubmissionWithGitLabInfo(submissionID uint) (*models.AssignmentSubmission, error) {
	var submission models.AssignmentSubmission
	err := s.db.Preload("Assignment").
		Preload("Assignment.Project").
		Preload("Student").
		First(&submission, submissionID).Error

	if err != nil {
		return nil, fmt.Errorf("submission not found: %w", err)
	}

	// 如果有GitLab信息，获取最新的提交状态
	if submission.CommitHash != "" && submission.Assignment.Project.GitLabProjectID > 0 {
		// 获取分支的最新提交
		var member models.ProjectMember
		if err := s.db.Where("project_id = ? AND student_id = ?", submission.Assignment.ProjectID, submission.StudentID).First(&member).Error; err == nil {
			commits, err := s.gitlabService.GetBranchCommits(submission.Assignment.Project.GitLabProjectID, member.PersonalBranch, 1)
			if err == nil && len(commits) > 0 {
				latestCommit := commits[0]
				if latestCommit.ID != submission.CommitHash {
					// 更新提交信息
					submission.CommitHash = latestCommit.ID
					submission.CommitMessage = latestCommit.Message
					submission.CommitURL = fmt.Sprintf("%s/-/commit/%s", submission.Assignment.Project.GitLabURL, latestCommit.ID)
					s.db.Save(&submission)
				}
			}
		}
	}

	return &submission, nil
}

// GetAssignmentStats 获取作业统计信息
func (s *AssignmentService) GetAssignmentStats(assignmentID uint) (*models.AssignmentStats, error) {
	stats := &models.AssignmentStats{}

	// 总提交数
	var totalSubmissions int64
	if err := s.db.Model(&models.AssignmentSubmission{}).
		Where("assignment_id = ?", assignmentID).
		Count(&totalSubmissions).Error; err != nil {
		return nil, fmt.Errorf("failed to count total submissions: %w", err)
	}
	stats.TotalSubmissions = int(totalSubmissions)

	// 已评审提交数
	var reviewedSubmissions int64
	if err := s.db.Model(&models.AssignmentSubmission{}).
		Where("assignment_id = ? AND status = 'reviewed'", assignmentID).
		Count(&reviewedSubmissions).Error; err != nil {
		return nil, fmt.Errorf("failed to count reviewed submissions: %w", err)
	}
	stats.ReviewedSubmissions = int(reviewedSubmissions)

	// 待评审提交数
	stats.PendingSubmissions = stats.TotalSubmissions - stats.ReviewedSubmissions

	// 平均分数
	if stats.ReviewedSubmissions > 0 {
		var totalScore float64
		err := s.db.Model(&models.AssignmentSubmission{}).
			Where("assignment_id = ? AND status = 'reviewed'", assignmentID).
			Select("AVG(score)").
			Scan(&totalScore).Error
		if err != nil {
			return nil, fmt.Errorf("failed to calculate average score: %w", err)
		}
		stats.AverageScore = totalScore
	}

	return stats, nil
}

// GetAssignmentSubmissionsWithGitLabStats 获取作业提交列表及GitLab统计
func (s *AssignmentService) GetAssignmentSubmissionsWithGitLabStats(assignmentID uint) ([]models.AssignmentSubmission, map[uint]*GitLabSubmissionStats, error) {
	var submissions []models.AssignmentSubmission
	err := s.db.Preload("Student").
		Where("assignment_id = ?", assignmentID).
		Find(&submissions).Error

	if err != nil {
		return nil, nil, fmt.Errorf("failed to get submissions: %w", err)
	}

	// 获取GitLab统计信息
	stats := make(map[uint]*GitLabSubmissionStats)
	var assignment models.Assignment
	if err := s.db.Preload("Project").First(&assignment, assignmentID).Error; err == nil {
		for _, submission := range submissions {
			if submission.BranchName != "" {
				// 获取分支提交统计
				commits, err := s.gitlabService.GetBranchCommits(assignment.Project.GitLabProjectID, submission.BranchName, 10)
				if err == nil {
					stats[submission.ID] = &GitLabSubmissionStats{
						TotalCommits: len(commits),
						LastCommitAt: &submission.SubmittedAt,
						BranchName:   submission.BranchName,
						BranchURL:    submission.BranchURL,
						MRStatus:     submission.CodeReviewStatus,
						MRUrl:        submission.MergeRequestURL,
					}
				}
			}
		}
	}

	return submissions, stats, nil
}

// CreateAssignmentWithGitLabIntegration 创建作业并设置GitLab集成
func (s *AssignmentService) CreateAssignmentWithGitLabIntegration(teacherID uint, req *CreateAssignmentRequest) (*models.Assignment, error) {
	// 首先创建作业
	assignment, err := s.CreateAssignment(teacherID, req)
	if err != nil {
		return nil, err
	}

	// 如果项目有GitLab集成，创建作业相关的Issue
	if assignment.Project.GitLabProjectID > 0 {
		issueTitle := fmt.Sprintf("Assignment: %s", assignment.Title)
		issueDescription := fmt.Sprintf("## Assignment Description\n\n%s\n\n## Due Date\n\n%s\n\n## Instructions\n\n1. Create your submission in your personal branch\n2. Submit files to the `students/[username]/assignment-%d/` directory\n3. Create a merge request when ready for review\n\n## Required Files\n\n%v",
			assignment.Description,
			assignment.DueDate.Format("2006-01-02 15:04:05"),
			assignment.ID,
			assignment.RequiredFiles)

		_, err := s.gitlabService.CreateIssue(
			assignment.Project.GitLabProjectID,
			issueTitle,
			issueDescription,
			[]string{"assignment", "homework"},
			&assignment.DueDate,
		)
		if err != nil {
			fmt.Printf("Warning: Failed to create assignment issue: %v\n", err)
		}
	}

	return assignment, nil
}

// GitLabSubmissionStats GitLab提交统计
type GitLabSubmissionStats struct {
	TotalCommits int        `json:"total_commits"`
	LastCommitAt *time.Time `json:"last_commit_at"`
	BranchName   string     `json:"branch_name"`
	BranchURL    string     `json:"branch_url"`
	MRStatus     string     `json:"mr_status"`
	MRUrl        string     `json:"mr_url"`
}
