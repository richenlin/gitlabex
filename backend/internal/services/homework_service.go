package services

import (
	"fmt"
	"gitlabex/internal/models"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// HomeworkService 作业服务
type HomeworkService struct {
	*BaseService
	GitLabService *GitLabService
}

// NewHomeworkService 创建作业服务
func NewHomeworkService(db *gorm.DB, gitlabService *GitLabService) *HomeworkService {
	return &HomeworkService{
		BaseService:   NewBaseService(db, gitlabService.Config),
		GitLabService: gitlabService,
	}
}

// CreateHomework 创建作业
func (s *HomeworkService) CreateHomework(homework *models.Homework) error {
	return s.DB.Create(homework).Error
}

// GetHomeworkByID 根据ID获取作业
func (s *HomeworkService) GetHomeworkByID(id uuid.UUID) (*models.Homework, error) {
	var homework models.Homework
	err := s.DB.Preload("Project").Preload("Creator").First(&homework, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &homework, nil
}

// GetHomeworkByProject 获取项目下的所有作业
func (s *HomeworkService) GetHomeworkByProject(projectID uuid.UUID) ([]models.Homework, error) {
	var homeworks []models.Homework
	err := s.DB.Preload("Project").Preload("Creator").
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		Find(&homeworks).Error
	return homeworks, err
}

// UpdateHomework 更新作业信息
func (s *HomeworkService) UpdateHomework(id uuid.UUID, updates map[string]interface{}) error {
	return s.DB.Model(&models.Homework{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteHomework 删除作业
func (s *HomeworkService) DeleteHomework(id uuid.UUID) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		// 删除相关的作业提交
		if err := tx.Delete(&models.Submission{}, "homework_id = ?", id).Error; err != nil {
			return err
		}

		// 删除作业
		return tx.Delete(&models.Homework{}, "id = ?", id).Error
	})
}

// SubmitHomework 提交作业
func (s *HomeworkService) SubmitHomework(submission *models.Submission) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		// 检查是否已经提交过
		var existingSubmission models.Submission
		err := tx.Where("homework_id = ? AND student_id = ?",
			submission.HomeworkID, submission.StudentID).First(&existingSubmission).Error

		if err == nil {
			// 已存在提交，更新提交
			now := time.Now()
			return tx.Model(&existingSubmission).
				Updates(map[string]interface{}{
					"git_lab_branch": submission.GitLabBranch,
					"submitted_at":   &now,
					"status":         models.SubmissionStatusSubmitted,
				}).Error
		}

		// 创建新提交
		now := time.Now()
		submission.SubmittedAt = &now
		submission.Status = models.SubmissionStatusSubmitted

		if err := tx.Create(submission).Error; err != nil {
			return err
		}

		// 更新作业提交计数
		return tx.Model(&models.Homework{}).
			Where("id = ?", submission.HomeworkID).
			UpdateColumn("submission_count", gorm.Expr("submission_count + 1")).Error
	})
}

// GetSubmissions 获取作业的所有提交
func (s *HomeworkService) GetSubmissions(homeworkID uuid.UUID) ([]models.Submission, error) {
	var submissions []models.Submission
	err := s.DB.Preload("Student").Preload("Grader").
		Where("homework_id = ?", homeworkID).
		Order("submitted_at DESC").
		Find(&submissions).Error
	return submissions, err
}

// GetUserSubmissions 获取用户的作业提交
func (s *HomeworkService) GetUserSubmissions(userID uuid.UUID) ([]models.Submission, error) {
	var submissions []models.Submission
	err := s.DB.Preload("Homework").Preload("Homework.Project").
		Where("student_id = ?", userID).
		Order("submitted_at DESC").
		Find(&submissions).Error
	return submissions, err
}

// GetUserSubmissionForHomework 获取用户对特定作业的提交
func (s *HomeworkService) GetUserSubmissionForHomework(homeworkID uuid.UUID, userID uuid.UUID) (*models.Submission, error) {
	var submission models.Submission
	err := s.DB.Preload("Homework").Preload("Student").
		Where("homework_id = ? AND student_id = ?", homeworkID, userID).
		First(&submission).Error
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

// GradeHomework 评分作业
func (s *HomeworkService) GradeHomework(submissionID uuid.UUID, grade int, feedback string, graderID uuid.UUID) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		// 更新提交
		err := tx.Model(&models.Submission{}).
			Where("id = ?", submissionID).
			Updates(map[string]interface{}{
				"grade":     grade,
				"feedback":  feedback,
				"grader_id": graderID,
				"graded_at": time.Now(),
				"status":    "graded",
			}).Error

		if err != nil {
			return err
		}

		// 获取提交信息
		var submission models.Submission
		if err := tx.First(&submission, "id = ?", submissionID).Error; err != nil {
			return err
		}

		// 更新作业已评分计数
		return tx.Model(&models.Homework{}).
			Where("id = ?", submission.HomeworkID).
			UpdateColumn("graded_count", gorm.Expr("graded_count + 1")).Error
	})
}

// GetSubmissionByID 根据ID获取提交
func (s *HomeworkService) GetSubmissionByID(id uuid.UUID) (*models.Submission, error) {
	var submission models.Submission
	err := s.DB.Preload("Homework").Preload("Student").Preload("Grader").
		First(&submission, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

// GetSubmissionByUserAndHomework 根据用户和作业获取提交
func (s *HomeworkService) GetSubmissionByUserAndHomework(userID, homeworkID uuid.UUID) (*models.Submission, error) {
	var submission models.Submission
	err := s.DB.Where("student_id = ? AND homework_id = ?", userID, homeworkID).First(&submission).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &submission, nil
}

// GetHomeworkStats 获取作业统计信息
func (s *HomeworkService) GetHomeworkStats(projectID uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总作业数量
	var totalCount int64
	s.DB.Model(&models.Homework{}).Where("project_id = ?", projectID).Count(&totalCount)
	stats["total_count"] = totalCount

	// 按状态统计
	var statusCounts []struct {
		Status string
		Count  int64
	}
	s.DB.Model(&models.Homework{}).
		Where("project_id = ?", projectID).
		Select("status, count(*) as count").
		Group("status").
		Find(&statusCounts)

	statusStats := make(map[string]int64)
	for _, sc := range statusCounts {
		statusStats[sc.Status] = sc.Count
	}
	stats["statuses"] = statusStats

	// 提交统计
	var submissionCount int64
	s.DB.Model(&models.Submission{}).
		Joins("JOIN homeworks ON submissions.homework_id = homeworks.id").
		Where("homeworks.project_id = ?", projectID).
		Count(&submissionCount)
	stats["total_submissions"] = submissionCount

	// 已评分统计
	var gradedCount int64
	s.DB.Model(&models.Submission{}).
		Joins("JOIN homeworks ON submissions.homework_id = homeworks.id").
		Where("homeworks.project_id = ? AND submissions.status = ?", projectID, "graded").
		Count(&gradedCount)
	stats["graded_submissions"] = gradedCount

	return stats, nil
}

// GetPendingReviews 获取待评分的作业
func (s *HomeworkService) GetPendingReviews(projectID uuid.UUID, graderID uuid.UUID) ([]models.Submission, error) {
	var submissions []models.Submission
	err := s.DB.Preload("Homework").Preload("Student").
		Joins("JOIN homeworks ON submissions.homework_id = homeworks.id").
		Where("homeworks.project_id = ? AND submissions.status = ? AND submissions.grade IS NULL",
			projectID, "submitted").
		Order("submitted_at ASC").
		Find(&submissions).Error
	return submissions, err
}

// AutoGradeSubmissions 自动评分（用于测试或批量评分）
func (s *HomeworkService) AutoGradeSubmissions(homeworkID uuid.UUID, defaultGrade int, feedback string, graderID uuid.UUID) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var submissions []models.Submission
		err := tx.Where("homework_id = ? AND grade IS NULL", homeworkID).Find(&submissions).Error
		if err != nil {
			return err
		}

		for _, submission := range submissions {
			if err := s.GradeHomework(submission.ID, defaultGrade, feedback, graderID); err != nil {
				return err
			}
		}

		return nil
	})
}

// CheckDueDate 检查作业是否逾期
func (s *HomeworkService) CheckDueDate(homeworkID uuid.UUID) error {
	var homework models.Homework
	err := s.DB.First(&homework, "id = ?", homeworkID).Error
	if err != nil {
		return err
	}

	if homework.DueDate != nil && time.Now().After(*homework.DueDate) {
		// 更新逾期提交的状态
		return s.DB.Model(&models.Submission{}).
			Where("homework_id = ? AND status = ?", homeworkID, "submitted").
			Update("status", "late").Error
	}

	return nil
}

// GetGradeDistribution 获取成绩分布
func (s *HomeworkService) GetGradeDistribution(homeworkID uuid.UUID) (map[string]interface{}, error) {
	distribution := make(map[string]interface{})

	var stats struct {
		Min   *int
		Max   *int
		Avg   float64
		Count int64
	}

	err := s.DB.Model(&models.Submission{}).
		Where("homework_id = ? AND grade IS NOT NULL", homeworkID).
		Select("MIN(grade) as min, MAX(grade) as max, AVG(grade) as avg, COUNT(*) as count").
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	distribution["min"] = stats.Min
	distribution["max"] = stats.Max
	distribution["average"] = stats.Avg
	distribution["count"] = stats.Count

	// 分数段统计
	var gradeRanges []struct {
		Range string
		Count int64
	}

	s.DB.Raw(`
		SELECT 
			CASE 
				WHEN grade >= 90 THEN '90-100'
				WHEN grade >= 80 THEN '80-89'
				WHEN grade >= 70 THEN '70-79'
				WHEN grade >= 60 THEN '60-69'
				ELSE '0-59'
			END as range,
			COUNT(*) as count
		FROM submissions 
		WHERE homework_id = ? AND grade IS NOT NULL
		GROUP BY range
		ORDER BY range
	`, homeworkID).Scan(&gradeRanges)

	ranges := make(map[string]int64)
	for _, gr := range gradeRanges {
		ranges[gr.Range] = gr.Count
	}
	distribution["ranges"] = ranges

	return distribution, nil
}

// CloneSubmission 克隆作业提交（用于创建新分支）
func (s *HomeworkService) CloneSubmission(originalSubmissionID uuid.UUID, newBranchName string, newStudentID uuid.UUID) (*models.Submission, error) {
	var original models.Submission
	err := s.DB.First(&original, "id = ?", originalSubmissionID).Error
	if err != nil {
		return nil, err
	}

	newSubmission := &models.Submission{
		HomeworkID:   original.HomeworkID,
		StudentID:    newStudentID,
		GitLabBranch: newBranchName,
		Status:       models.SubmissionStatusSubmitted,
	}

	err = s.SubmitHomework(newSubmission)
	return newSubmission, err
}

// ExportGrades 导出成绩
func (s *HomeworkService) ExportGrades(homeworkID uuid.UUID) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	err := s.DB.Table("submissions").
		Select(`
			users.username as student_name,
			users.email as student_email,
			submissions.branch_name,
			submissions.submitted_at,
			submissions.grade,
			submissions.feedback,
			submissions.status
		`).
		Joins("JOIN users ON submissions.student_id = users.id").
		Where("submissions.homework_id = ?", homeworkID).
		Order("users.username").
		Scan(&results).Error

	return results, err
}

// ImportGrades 导入成绩（批量更新）
func (s *HomeworkService) ImportGrades(homeworkID uuid.UUID, grades []map[string]interface{}) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		for _, gradeData := range grades {
			studentID, ok := gradeData["student_id"].(uuid.UUID)
			if !ok {
				continue
			}

			grade, ok := gradeData["grade"].(int)
			if !ok {
				continue
			}

			feedback, _ := gradeData["feedback"].(string)

			var submission models.Submission
			err := tx.Where("homework_id = ? AND student_id = ?", homeworkID, studentID).First(&submission).Error
			if err != nil {
				continue
			}

			// 更新成绩
			err = s.GradeHomework(submission.ID, grade, feedback, studentID)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// GenerateReport 生成作业报告
func (s *HomeworkService) GenerateReport(projectID uuid.UUID) (map[string]interface{}, error) {
	report := make(map[string]interface{})

	// 获取项目所有作业
	homeworks, err := s.GetHomeworkByProject(projectID)
	if err != nil {
		return nil, err
	}

	report["homeworks"] = homeworks

	// 获取统计信息
	stats, err := s.GetHomeworkStats(projectID)
	if err != nil {
		return nil, err
	}

	report["stats"] = stats

	return report, nil
}

// BulkCreateHomework 批量创建作业
func (s *HomeworkService) BulkCreateHomework(homeworks []models.Homework) error {
	return s.DB.CreateInBatches(homeworks, 100).Error
}

// ArchiveHomework 归档作业
func (s *HomeworkService) ArchiveHomework(homeworkID uuid.UUID) error {
	return s.DB.Model(&models.Homework{}).
		Where("id = ?", homeworkID).
		Update("status", "archived").Error
}

// RestoreHomework 恢复作业
func (s *HomeworkService) RestoreHomework(homeworkID uuid.UUID) error {
	return s.DB.Model(&models.Homework{}).
		Where("id = ?", homeworkID).
		Update("status", "active").Error
}

// GetStudentProgress 获取学生进度
func (s *HomeworkService) GetStudentProgress(userID uuid.UUID, projectID uuid.UUID) (map[string]interface{}, error) {
	progress := make(map[string]interface{})

	var totalHomeworks int64
	s.DB.Model(&models.Homework{}).
		Where("project_id = ?", projectID).
		Count(&totalHomeworks)
	progress["total_homeworks"] = totalHomeworks

	var submittedCount int64
	s.DB.Model(&models.Submission{}).
		Joins("JOIN homeworks ON submissions.homework_id = homeworks.id").
		Where("homeworks.project_id = ? AND submissions.student_id = ?", projectID, userID).
		Count(&submittedCount)
	progress["submitted_count"] = submittedCount

	var gradedCount int64
	s.DB.Model(&models.Submission{}).
		Joins("JOIN homeworks ON submissions.homework_id = homeworks.id").
		Where("homeworks.project_id = ? AND submissions.student_id = ? AND submissions.grade IS NOT NULL",
			projectID, userID).
		Count(&gradedCount)
	progress["graded_count"] = gradedCount

	var totalGrade int64
	s.DB.Model(&models.Submission{}).
		Joins("JOIN homeworks ON submissions.homework_id = homeworks.id").
		Where("homeworks.project_id = ? AND submissions.student_id = ? AND submissions.grade IS NOT NULL",
			projectID, userID).
		Select("COALESCE(SUM(grade), 0)").
		Scan(&totalGrade)
	progress["total_grade"] = totalGrade

	return progress, nil
}

// GetAssignmentDetails 获取作业分配详情
func (s *HomeworkService) GetAssignmentDetails(homeworkID uuid.UUID) (map[string]interface{}, error) {
	details := make(map[string]interface{})

	// 获取作业基本信息
	homework, err := s.GetHomeworkByID(homeworkID)
	if err != nil {
		return nil, err
	}
	details["homework"] = homework

	// 获取提交列表
	submissions, err := s.GetSubmissions(homeworkID)
	if err != nil {
		return nil, err
	}
	details["submissions"] = submissions

	// 获取成绩分布
	gradeDistribution, err := s.GetGradeDistribution(homeworkID)
	if err != nil {
		return nil, err
	}
	details["grade_distribution"] = gradeDistribution

	return details, nil
}

// CreateAssignmentTemplate 创建作业模板
func (s *HomeworkService) CreateAssignmentTemplate(template *models.AssignmentTemplate) error {
	return s.DB.Create(template).Error
}

// UseAssignmentTemplate 使用模板创建作业
func (s *HomeworkService) UseAssignmentTemplate(templateID uuid.UUID, projectID uuid.UUID, creatorID int64) (*models.Homework, error) {
	var template models.AssignmentTemplate
	err := s.DB.First(&template, "id = ?", templateID).Error
	if err != nil {
		return nil, err
	}

	homework := &models.Homework{
		ProjectID:   projectID,
		Title:       template.Title,
		Description: template.Description,
		DueDate:     template.DueDate,
		MaxGrade:    template.MaxGrade,
		CreatorID:   creatorID,
		Status:      "active",
	}

	err = s.CreateHomework(homework)
	return homework, err
}

// AssignmentTemplate 作业模板（已移至models包）

// GetAssignmentTemplates 获取作业模板
func (s *HomeworkService) GetAssignmentTemplates(isPublic bool, creatorID uuid.UUID) ([]models.AssignmentTemplate, error) {
	var templates []models.AssignmentTemplate

	query := s.DB.Model(&models.AssignmentTemplate{})

	if isPublic {
		query = query.Where("is_public = true")
	} else {
		query = query.Where("creator_id = ?", creatorID)
	}

	err := query.Preload("Creator").Order("created_at DESC").Find(&templates).Error
	return templates, err
}

// UpdateAssignmentTemplate 更新作业模板
func (s *HomeworkService) UpdateAssignmentTemplate(id uuid.UUID, updates map[string]interface{}) error {
	return s.DB.Model(&models.AssignmentTemplate{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteAssignmentTemplate 删除作业模板
func (s *HomeworkService) DeleteAssignmentTemplate(id uuid.UUID) error {
	return s.DB.Delete(&models.AssignmentTemplate{}, "id = ?", id).Error
}

// ValidateGrade 验证成绩是否有效
func (s *HomeworkService) ValidateGrade(homeworkID uuid.UUID, grade int) error {
	var homework models.Homework
	err := s.DB.First(&homework, "id = ?", homeworkID).Error
	if err != nil {
		return err
	}

	if grade < 0 || grade > homework.MaxGrade {
		return gorm.ErrInvalidData
	}

	return nil
}

// GetOverdueSubmissions 获取逾期提交
func (s *HomeworkService) GetOverdueSubmissions() ([]models.Submission, error) {
	var submissions []models.Submission

	err := s.DB.Preload("Homework").Preload("Student").
		Joins("JOIN homeworks ON submissions.homework_id = homeworks.id").
		Where("homeworks.due_date IS NOT NULL AND submissions.submitted_at > homeworks.due_date").
		Order("submitted_at DESC").
		Find(&submissions).Error

	return submissions, err
}

// SendReminder 发送作业提醒
func (s *HomeworkService) SendReminder(homeworkID uuid.UUID, userIDs []uuid.UUID) error {
	// 这里应该集成通知服务
	// 目前返回空实现
	return nil
}

// BulkUpdateDueDate 批量更新截止日期
func (s *HomeworkService) BulkUpdateDueDate(homeworkIDs []uuid.UUID, newDueDate *time.Time) error {
	return s.DB.Model(&models.Homework{}).
		Where("id IN ?", homeworkIDs).
		Update("due_date", newDueDate).Error
}

// ArchiveOldHomework 归档旧作业
func (s *HomeworkService) ArchiveOldHomework(daysOld int) error {
	cutoffDate := time.Now().AddDate(0, 0, -daysOld)

	return s.DB.Model(&models.Homework{}).
		Where("created_at < ? AND status = ?", cutoffDate, "active").
		Update("status", "archived").Error
}

// CreateStudentBranch 为学生创建个人作业分支
func (s *HomeworkService) CreateStudentBranch(homeworkID uuid.UUID, studentID uuid.UUID) error {
	// 获取作业信息
	var homework models.Homework
	if err := s.DB.Preload("Project").First(&homework, homeworkID).Error; err != nil {
		return err
	}

	// TODO: 重构学生信息获取以使用GitLab用户系统
	// 暂时跳过学生信息检查

	// 检查项目是否有GitLab项目ID
	if homework.Project.GitLabProjectID == nil {
		return fmt.Errorf("项目没有关联GitLab项目")
	}

	// TODO: 重构分支创建以使用GitLab用户系统
	// 暂时跳过分支创建
	branchName := fmt.Sprintf("homework-%s-student-%s", homework.ID.String()[:8], studentID.String()[:8])

	// TODO: 实现GitLab分支创建
	// _, err := s.GitLabService.CreateBranch(...)
	// if err != nil {
	//     return fmt.Errorf("创建GitLab分支失败: %v", err)
	// }

	// 更新作业记录中的分支信息
	if homework.GitLabBranch == "" {
		homework.GitLabBranch = branchName
		s.DB.Save(&homework)
	}

	return nil
}

// GetHomeworkBranches 获取作业的所有学生分支
func (s *HomeworkService) GetHomeworkBranches(homeworkID uuid.UUID, accessToken string) ([]map[string]interface{}, error) {
	// 获取作业信息
	var homework models.Homework
	if err := s.DB.Preload("Project").First(&homework, homeworkID).Error; err != nil {
		return nil, err
	}

	if homework.Project.GitLabProjectID == nil {
		return nil, fmt.Errorf("项目没有关联GitLab项目")
	}

	// 获取项目的所有分支 - TODO: 实现GetProjectBranches方法
	// branches, err := s.GitLabService.GetProjectBranches(accessToken, *homework.Project.GitLabProjectID)
	// if err != nil {
	// 	return nil, err
	// }
	branches := []map[string]interface{}{} // 临时返回空数组

	// 过滤作业相关的分支
	var homeworkBranches []map[string]interface{}
	homeworkPrefix := fmt.Sprintf("homework-%s", homework.ID.String()[:8])

	for _, branch := range branches {
		if branchName, ok := branch["name"].(string); ok {
			if strings.HasPrefix(branchName, homeworkPrefix) {
				// 从分支名称解析学生用户名
				parts := strings.Split(branchName, "-")
				if len(parts) >= 3 {
					username := strings.Join(parts[2:], "-")

					// TODO: 重构学生信息获取以使用GitLab用户系统
					// 暂时只显示用户名
					branch["student"] = map[string]interface{}{
						"username": username,
					}
				}
				homeworkBranches = append(homeworkBranches, branch)
			}
		}
	}

	return homeworkBranches, nil
}

// SubmitHomeworkToBranch 提交作业到个人分支
func (s *HomeworkService) SubmitHomeworkToBranch(homeworkID uuid.UUID, studentID uuid.UUID, content string, files []string) (*models.Submission, error) {
	// 首先尝试创建分支（如果不存在）
	if err := s.CreateStudentBranch(homeworkID, studentID); err != nil {
		// 分支可能已存在，继续执行
	}

	// 获取作业信息
	var homework models.Homework
	if err := s.DB.Preload("Project").First(&homework, homeworkID).Error; err != nil {
		return nil, err
	}

	// TODO: 重构学生信息获取以使用GitLab用户系统
	// 暂时使用学生ID生成分支名称
	branchName := fmt.Sprintf("homework-%s-student-%s", homework.ID.String()[:8], studentID.String()[:8])

	// 创建提交记录
	submission := &models.Submission{
		HomeworkID:      homeworkID,
		StudentID:       studentID,
		Content:         content,
		Status:          models.SubmissionStatusSubmitted,
		GitLabCommitSHA: "", // 这里应该从实际的GitLab提交获取
		GitLabBranch:    branchName,
	}

	if err := s.DB.Create(submission).Error; err != nil {
		return nil, err
	}

	// 预加载关联数据
	s.DB.Preload("Homework").Preload("Student").First(submission, submission.ID)

	return submission, nil
}

// GetStudentBranchInfo 获取学生分支信息
func (s *HomeworkService) GetStudentBranchInfo(homeworkID uuid.UUID, studentID uuid.UUID) (map[string]interface{}, error) {
	// 获取作业信息
	var homework models.Homework
	if err := s.DB.Preload("Project").First(&homework, homeworkID).Error; err != nil {
		return nil, err
	}

	// 获取学生信息
	// TODO: 重构学生信息获取以使用GitLab用户系统
	// 暂时使用学生ID生成分支名称
	branchName := fmt.Sprintf("homework-%s-student-%s", homework.ID.String()[:8], studentID.String()[:8])

	if homework.Project.GitLabProjectID == nil {
		return nil, fmt.Errorf("项目没有关联GitLab项目")
	}

	// 获取分支信息 - TODO: 实现GetBranchInfo方法
	// branchInfo, err := s.GitLabService.GetBranchInfo(
	// 	student.AccessToken,
	// 	*homework.Project.GitLabProjectID,
	// 	branchName,
	// )
	// if err != nil {
	// 	return nil, err
	// }
	branchInfo := map[string]interface{}{
		"name":      branchName,
		"protected": false,
	} // 临时返回简单分支信息

	// 获取提交历史 - TODO: 实现GetBranchCommits方法
	// commits, err := s.GitLabService.GetBranchCommits(
	// 	student.AccessToken,
	// 	*homework.Project.GitLabProjectID,
	// 	branchName,
	// )
	// if err != nil {
	// 	commits = []map[string]interface{}{} // 如果获取失败，返回空数组
	// }
	commits := []map[string]interface{}{} // 临时返回空提交历史

	return map[string]interface{}{
		"branch_name": branchName,
		"branch_info": branchInfo,
		"commits":     commits,
		"project_url": homework.Project.GitLabURL,
	}, nil
}
