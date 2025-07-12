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
}

// NewAssignmentService 创建作业管理服务
func NewAssignmentService(db *gorm.DB, permissionService *PermissionService) *AssignmentService {
	return &AssignmentService{
		db:                db,
		permissionService: permissionService,
	}
}

// CreateAssignmentRequest 创建作业请求
type CreateAssignmentRequest struct {
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	ProjectID   uint      `json:"project_id" binding:"required"`
	DueDate     time.Time `json:"due_date" binding:"required"`
	MaxScore    int       `json:"max_score"`
}

// UpdateAssignmentRequest 更新作业请求
type UpdateAssignmentRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	DueDate     *time.Time `json:"due_date"`
	MaxScore    *int       `json:"max_score"`
	Status      string     `json:"status"`
}

// SubmitAssignmentRequest 提交作业请求
type SubmitAssignmentRequest struct {
	Content  string `json:"content"`
	FilePath string `json:"file_path"`
}

// CreateAssignment 创建作业（老师权限）
func (s *AssignmentService) CreateAssignment(teacherID uint, req *CreateAssignmentRequest) (*models.Assignment, error) {
	// 验证课题是否存在且属于该老师
	var project models.Project
	if err := s.db.Where("id = ? AND teacher_id = ?", req.ProjectID, teacherID).First(&project).Error; err != nil {
		return nil, fmt.Errorf("project not found or access denied: %w", err)
	}

	maxScore := req.MaxScore
	if maxScore <= 0 {
		maxScore = 100
	}

	assignment := &models.Assignment{
		Title:       req.Title,
		Description: req.Description,
		ProjectID:   req.ProjectID,
		TeacherID:   teacherID,
		DueDate:     req.DueDate,
		MaxScore:    maxScore,
		Status:      "active",
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
	if req.MaxScore != nil {
		assignment.MaxScore = *req.MaxScore
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

// SubmitAssignment 学生提交作业
func (s *AssignmentService) SubmitAssignment(studentID uint, assignmentID uint, req *SubmitAssignmentRequest) (*models.AssignmentSubmission, error) {
	// 验证作业是否存在
	var assignment models.Assignment
	if err := s.db.Preload("Project").First(&assignment, assignmentID).Error; err != nil {
		return nil, fmt.Errorf("assignment not found: %w", err)
	}

	// 验证学生是否有权限提交（是否在课题中）
	var member models.ProjectMember
	if err := s.db.Where("project_id = ? AND student_id = ? AND status = 'active'",
		assignment.ProjectID, studentID).First(&member).Error; err != nil {
		return nil, fmt.Errorf("student not in project: %w", err)
	}

	// 检查是否已经提交过
	var existingSubmission models.AssignmentSubmission
	err := s.db.Where("assignment_id = ? AND student_id = ?", assignmentID, studentID).First(&existingSubmission).Error
	if err == nil {
		// 更新现有提交
		existingSubmission.Content = req.Content
		existingSubmission.FilePath = req.FilePath
		existingSubmission.SubmittedAt = time.Now()
		existingSubmission.Status = "submitted"

		if err := s.db.Save(&existingSubmission).Error; err != nil {
			return nil, fmt.Errorf("failed to update submission: %w", err)
		}

		// 重新加载数据
		if err := s.db.Preload("Assignment").Preload("Student").First(&existingSubmission, existingSubmission.ID).Error; err != nil {
			return nil, fmt.Errorf("failed to reload submission: %w", err)
		}

		// TODO: 发送通知给老师（待通知服务实现）

		return &existingSubmission, nil
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing submission: %w", err)
	}

	// 创建新的提交
	submission := &models.AssignmentSubmission{
		AssignmentID: assignmentID,
		StudentID:    studentID,
		Content:      req.Content,
		FilePath:     req.FilePath,
		SubmittedAt:  time.Now(),
		Status:       "submitted",
	}

	if err := s.db.Create(submission).Error; err != nil {
		return nil, fmt.Errorf("failed to create submission: %w", err)
	}

	// 预加载关联数据
	if err := s.db.Preload("Assignment").Preload("Student").First(submission, submission.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load submission: %w", err)
	}

	// TODO: 发送通知给老师（待通知服务实现）

	return submission, nil
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
