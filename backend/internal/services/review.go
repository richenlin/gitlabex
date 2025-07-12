package services

import (
	"fmt"
	"time"

	"gitlabex/internal/models"

	"gorm.io/gorm"
)

// ReviewService 评审管理服务
type ReviewService struct {
	db                *gorm.DB
	permissionService *PermissionService
}

// NewReviewService 创建评审管理服务
func NewReviewService(db *gorm.DB, permissionService *PermissionService) *ReviewService {
	return &ReviewService{
		db:                db,
		permissionService: permissionService,
	}
}

// CreateReviewRequest 创建评审请求
type CreateReviewRequest struct {
	SubmissionID uint   `json:"submission_id" binding:"required"`
	Score        int    `json:"score" binding:"required"`
	Comment      string `json:"comment"`
	Feedback     string `json:"feedback"`
}

// UpdateReviewRequest 更新评审请求
type UpdateReviewRequest struct {
	Score    *int   `json:"score"`
	Comment  string `json:"comment"`
	Feedback string `json:"feedback"`
	Status   string `json:"status"`
}

// CreateReview 老师创建评审
func (s *ReviewService) CreateReview(teacherID uint, req *CreateReviewRequest) (*models.Review, error) {
	// 验证提交是否存在
	var submission models.AssignmentSubmission
	if err := s.db.Preload("Assignment").First(&submission, req.SubmissionID).Error; err != nil {
		return nil, fmt.Errorf("submission not found: %w", err)
	}

	// 验证老师是否有权限评审（是否是作业的创建者）
	if submission.Assignment.TeacherID != teacherID {
		return nil, fmt.Errorf("access denied: not the assignment creator")
	}

	// 检查是否已经评审过
	var existingReview models.Review
	if err := s.db.Where("submission_id = ?", req.SubmissionID).First(&existingReview).Error; err == nil {
		return nil, fmt.Errorf("submission already reviewed")
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing review: %w", err)
	}

	// 创建评审
	review := &models.Review{
		SubmissionID: req.SubmissionID,
		TeacherID:    teacherID,
		Score:        req.Score,
		Comment:      req.Comment,
		Feedback:     req.Feedback,
		Status:       "completed",
		ReviewedAt:   time.Now(),
	}

	if err := s.db.Create(review).Error; err != nil {
		return nil, fmt.Errorf("failed to create review: %w", err)
	}

	// 更新作业提交状态和分数
	submission.Status = "reviewed"
	submission.Score = req.Score
	if err := s.db.Save(&submission).Error; err != nil {
		return nil, fmt.Errorf("failed to update submission status: %w", err)
	}

	// 预加载关联数据
	if err := s.db.Preload("Submission").Preload("Teacher").First(review, review.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load review: %w", err)
	}

	// TODO: 发送通知给学生（待通知服务实现）

	return review, nil
}

// GetReviewByID 根据ID获取评审
func (s *ReviewService) GetReviewByID(reviewID uint) (*models.Review, error) {
	var review models.Review
	err := s.db.Preload("Submission").
		Preload("Submission.Assignment").
		Preload("Submission.Student").
		Preload("Teacher").
		First(&review, reviewID).Error

	if err != nil {
		return nil, fmt.Errorf("review not found: %w", err)
	}

	return &review, nil
}

// GetReviewBySubmissionID 根据提交ID获取评审
func (s *ReviewService) GetReviewBySubmissionID(submissionID uint) (*models.Review, error) {
	var review models.Review
	err := s.db.Preload("Submission").
		Preload("Submission.Assignment").
		Preload("Submission.Student").
		Preload("Teacher").
		Where("submission_id = ?", submissionID).
		First(&review).Error

	if err != nil {
		return nil, fmt.Errorf("review not found: %w", err)
	}

	return &review, nil
}

// GetReviewsByTeacher 获取老师的评审列表
func (s *ReviewService) GetReviewsByTeacher(teacherID uint) ([]models.Review, error) {
	var reviews []models.Review
	err := s.db.Preload("Submission").
		Preload("Submission.Assignment").
		Preload("Submission.Student").
		Where("teacher_id = ?", teacherID).
		Order("reviewed_at DESC").
		Find(&reviews).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get reviews: %w", err)
	}

	return reviews, nil
}

// GetReviewsByStudent 获取学生的评审列表
func (s *ReviewService) GetReviewsByStudent(studentID uint) ([]models.Review, error) {
	var reviews []models.Review
	err := s.db.Preload("Submission").
		Preload("Submission.Assignment").
		Preload("Teacher").
		Joins("JOIN assignment_submissions ON reviews.submission_id = assignment_submissions.id").
		Where("assignment_submissions.student_id = ?", studentID).
		Order("reviews.reviewed_at DESC").
		Find(&reviews).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get reviews: %w", err)
	}

	return reviews, nil
}

// GetReviewsByAssignment 获取作业的评审列表
func (s *ReviewService) GetReviewsByAssignment(assignmentID uint) ([]models.Review, error) {
	var reviews []models.Review
	err := s.db.Preload("Submission").
		Preload("Submission.Student").
		Preload("Teacher").
		Joins("JOIN assignment_submissions ON reviews.submission_id = assignment_submissions.id").
		Where("assignment_submissions.assignment_id = ?", assignmentID).
		Order("reviews.reviewed_at DESC").
		Find(&reviews).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get reviews: %w", err)
	}

	return reviews, nil
}

// GetAllReviews 获取所有评审（管理员权限）
func (s *ReviewService) GetAllReviews(page, pageSize int) ([]models.Review, int64, error) {
	var reviews []models.Review
	var total int64

	// 计算总数
	if err := s.db.Model(&models.Review{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count reviews: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := s.db.Preload("Submission").
		Preload("Submission.Assignment").
		Preload("Submission.Student").
		Preload("Teacher").
		Order("reviewed_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&reviews).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get reviews: %w", err)
	}

	return reviews, total, nil
}

// UpdateReview 更新评审信息
func (s *ReviewService) UpdateReview(reviewID uint, req *UpdateReviewRequest) (*models.Review, error) {
	var review models.Review
	if err := s.db.First(&review, reviewID).Error; err != nil {
		return nil, fmt.Errorf("review not found: %w", err)
	}

	// 更新字段
	if req.Score != nil {
		review.Score = *req.Score
	}
	if req.Comment != "" {
		review.Comment = req.Comment
	}
	if req.Feedback != "" {
		review.Feedback = req.Feedback
	}
	if req.Status != "" {
		review.Status = req.Status
	}

	review.ReviewedAt = time.Now()

	if err := s.db.Save(&review).Error; err != nil {
		return nil, fmt.Errorf("failed to update review: %w", err)
	}

	// 如果更新了分数，也需要更新作业提交的分数
	if req.Score != nil {
		var submission models.AssignmentSubmission
		if err := s.db.First(&submission, review.SubmissionID).Error; err == nil {
			submission.Score = *req.Score
			s.db.Save(&submission)
		}
	}

	// 重新加载数据
	if err := s.db.Preload("Submission").Preload("Teacher").First(&review, review.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload review: %w", err)
	}

	return &review, nil
}

// DeleteReview 删除评审
func (s *ReviewService) DeleteReview(reviewID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取评审信息
		var review models.Review
		if err := tx.First(&review, reviewID).Error; err != nil {
			return fmt.Errorf("review not found: %w", err)
		}

		// 更新作业提交状态
		var submission models.AssignmentSubmission
		if err := tx.First(&submission, review.SubmissionID).Error; err == nil {
			submission.Status = "submitted"
			submission.Score = 0
			tx.Save(&submission)
		}

		// 删除评审
		if err := tx.Delete(&models.Review{}, reviewID).Error; err != nil {
			return fmt.Errorf("failed to delete review: %w", err)
		}

		return nil
	})
}

// GenerateAssignmentReport 生成作业报告
func (s *ReviewService) GenerateAssignmentReport(assignmentID uint) (*AssignmentReport, error) {
	// 获取作业信息
	var assignment models.Assignment
	if err := s.db.Preload("Teacher").Preload("Project").First(&assignment, assignmentID).Error; err != nil {
		return nil, fmt.Errorf("assignment not found: %w", err)
	}

	// 获取所有提交
	var submissions []models.AssignmentSubmission
	if err := s.db.Preload("Student").Preload("Review").
		Where("assignment_id = ?", assignmentID).Find(&submissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get submissions: %w", err)
	}

	// 生成报告
	report := &AssignmentReport{
		Assignment:    &assignment,
		Submissions:   submissions,
		GeneratedAt:   time.Now(),
		TotalStudents: len(submissions),
	}

	// 计算统计信息
	reviewedCount := 0
	totalScore := 0
	var maxScore, minScore int
	maxScoreSet := false

	for _, submission := range submissions {
		if submission.Review != nil {
			reviewedCount++
			totalScore += submission.Score

			if !maxScoreSet {
				maxScore = submission.Score
				minScore = submission.Score
				maxScoreSet = true
			} else {
				if submission.Score > maxScore {
					maxScore = submission.Score
				}
				if submission.Score < minScore {
					minScore = submission.Score
				}
			}
		}
	}

	report.ReviewedCount = reviewedCount
	report.PendingCount = report.TotalStudents - reviewedCount

	if reviewedCount > 0 {
		report.AverageScore = float64(totalScore) / float64(reviewedCount)
		report.MaxScore = maxScore
		report.MinScore = minScore
	}

	return report, nil
}

// AssignmentReport 作业报告结构
type AssignmentReport struct {
	Assignment    *models.Assignment            `json:"assignment"`
	Submissions   []models.AssignmentSubmission `json:"submissions"`
	GeneratedAt   time.Time                     `json:"generated_at"`
	TotalStudents int                           `json:"total_students"`
	ReviewedCount int                           `json:"reviewed_count"`
	PendingCount  int                           `json:"pending_count"`
	AverageScore  float64                       `json:"average_score"`
	MaxScore      int                           `json:"max_score"`
	MinScore      int                           `json:"min_score"`
}

// GetPendingReviews 获取待评审的提交列表
func (s *ReviewService) GetPendingReviews(teacherID uint) ([]models.AssignmentSubmission, error) {
	var submissions []models.AssignmentSubmission
	err := s.db.Preload("Assignment").
		Preload("Student").
		Joins("JOIN assignments ON assignment_submissions.assignment_id = assignments.id").
		Where("assignments.teacher_id = ? AND assignment_submissions.status = 'submitted'", teacherID).
		Order("assignment_submissions.submitted_at DESC").
		Find(&submissions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get pending reviews: %w", err)
	}

	return submissions, nil
}

// GetReviewProgress 获取评审进度统计
func (s *ReviewService) GetReviewProgress(teacherID uint) (*ReviewProgress, error) {
	progress := &ReviewProgress{}

	// 总的待评审数量
	var totalPending int64
	if err := s.db.Model(&models.AssignmentSubmission{}).
		Joins("JOIN assignments ON assignment_submissions.assignment_id = assignments.id").
		Where("assignments.teacher_id = ? AND assignment_submissions.status = 'submitted'", teacherID).
		Count(&totalPending).Error; err != nil {
		return nil, fmt.Errorf("failed to count pending submissions: %w", err)
	}
	progress.TotalPending = int(totalPending)

	// 今日完成的评审数量
	today := time.Now().Format("2006-01-02")
	var todayCompleted int64
	if err := s.db.Model(&models.Review{}).
		Where("teacher_id = ? AND DATE(reviewed_at) = ?", teacherID, today).
		Count(&todayCompleted).Error; err != nil {
		return nil, fmt.Errorf("failed to count today's completed reviews: %w", err)
	}
	progress.TodayCompleted = int(todayCompleted)

	// 总的评审数量
	var totalReviews int64
	if err := s.db.Model(&models.Review{}).
		Where("teacher_id = ?", teacherID).
		Count(&totalReviews).Error; err != nil {
		return nil, fmt.Errorf("failed to count total reviews: %w", err)
	}
	progress.TotalReviews = int(totalReviews)

	return progress, nil
}

// ReviewProgress 评审进度统计
type ReviewProgress struct {
	TotalPending   int `json:"total_pending"`
	TodayCompleted int `json:"today_completed"`
	TotalReviews   int `json:"total_reviews"`
}
