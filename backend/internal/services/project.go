package services

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"gitlabex/internal/models"

	"gorm.io/gorm"
)

// ProjectService 课题管理服务
type ProjectService struct {
	db                *gorm.DB
	permissionService *PermissionService
}

// NewProjectService 创建课题管理服务
func NewProjectService(db *gorm.DB, permissionService *PermissionService) *ProjectService {
	return &ProjectService{
		db:                db,
		permissionService: permissionService,
	}
}

// CreateProjectRequest 创建课题请求
type CreateProjectRequest struct {
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	ClassID     uint      `json:"class_id"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
}

// UpdateProjectRequest 更新课题请求
type UpdateProjectRequest struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
}

// JoinProjectRequest 加入课题请求
type JoinProjectRequest struct {
	Code string `json:"code" binding:"required"`
}

// CreateProject 创建课题（老师权限）
func (s *ProjectService) CreateProject(teacherID uint, req *CreateProjectRequest) (*models.Project, error) {
	// 生成唯一的课题代码
	code, err := s.generateProjectCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate project code: %w", err)
	}

	project := &models.Project{
		Name:        req.Name,
		Description: req.Description,
		Code:        code,
		TeacherID:   teacherID,
		ClassID:     req.ClassID,
		Status:      "active",
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}

	if err := s.db.Create(project).Error; err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// 预加载关联数据
	if err := s.db.Preload("Teacher").Preload("Class").First(project, project.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load project: %w", err)
	}

	return project, nil
}

// GetProjectByID 根据ID获取课题
func (s *ProjectService) GetProjectByID(projectID uint) (*models.Project, error) {
	var project models.Project
	err := s.db.Preload("Teacher").
		Preload("Class").
		Preload("Members").
		Preload("Students").
		Preload("Assignments").
		First(&project, projectID).Error

	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	return &project, nil
}

// GetProjectsByTeacher 获取老师创建的课题列表
func (s *ProjectService) GetProjectsByTeacher(teacherID uint) ([]models.Project, error) {
	var projects []models.Project
	err := s.db.Preload("Teacher").
		Preload("Class").
		Where("teacher_id = ?", teacherID).
		Order("created_at DESC").
		Find(&projects).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	return projects, nil
}

// GetProjectsByStudent 获取学生参加的课题列表
func (s *ProjectService) GetProjectsByStudent(studentID uint) ([]models.Project, error) {
	var projects []models.Project
	err := s.db.Preload("Teacher").
		Preload("Class").
		Joins("JOIN project_members ON projects.id = project_members.project_id").
		Where("project_members.student_id = ? AND project_members.status = 'active'", studentID).
		Order("project_members.joined_at DESC").
		Find(&projects).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	return projects, nil
}

// GetProjectsByClass 获取班级的课题列表
func (s *ProjectService) GetProjectsByClass(classID uint) ([]models.Project, error) {
	var projects []models.Project
	err := s.db.Preload("Teacher").
		Where("class_id = ?", classID).
		Order("created_at DESC").
		Find(&projects).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	return projects, nil
}

// GetAllProjects 获取所有课题（管理员权限）
func (s *ProjectService) GetAllProjects(page, pageSize int) ([]models.Project, int64, error) {
	var projects []models.Project
	var total int64

	// 计算总数
	if err := s.db.Model(&models.Project{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count projects: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := s.db.Preload("Teacher").
		Preload("Class").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&projects).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get projects: %w", err)
	}

	return projects, total, nil
}

// UpdateProject 更新课题信息
func (s *ProjectService) UpdateProject(projectID uint, req *UpdateProjectRequest) (*models.Project, error) {
	var project models.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// 更新字段
	if req.Name != "" {
		project.Name = req.Name
	}
	if req.Description != "" {
		project.Description = req.Description
	}
	if req.Status != "" {
		project.Status = req.Status
	}
	if req.StartDate != nil {
		project.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		project.EndDate = *req.EndDate
	}

	if err := s.db.Save(&project).Error; err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	// 重新加载数据
	if err := s.db.Preload("Teacher").Preload("Class").First(&project, project.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload project: %w", err)
	}

	return &project, nil
}

// DeleteProject 删除课题
func (s *ProjectService) DeleteProject(projectID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除课题成员关系
		if err := tx.Where("project_id = ?", projectID).Delete(&models.ProjectMember{}).Error; err != nil {
			return fmt.Errorf("failed to delete project members: %w", err)
		}

		// 删除作业提交
		if err := tx.Where("assignment_id IN (SELECT id FROM assignments WHERE project_id = ?)", projectID).Delete(&models.AssignmentSubmission{}).Error; err != nil {
			return fmt.Errorf("failed to delete assignment submissions: %w", err)
		}

		// 删除作业
		if err := tx.Where("project_id = ?", projectID).Delete(&models.Assignment{}).Error; err != nil {
			return fmt.Errorf("failed to delete assignments: %w", err)
		}

		// 删除课题
		if err := tx.Delete(&models.Project{}, projectID).Error; err != nil {
			return fmt.Errorf("failed to delete project: %w", err)
		}

		return nil
	})
}

// JoinProject 学生加入课题
func (s *ProjectService) JoinProject(studentID uint, code string) (*models.Project, error) {
	// 查找课题
	var project models.Project
	if err := s.db.Where("code = ? AND status = 'active'", code).First(&project).Error; err != nil {
		return nil, fmt.Errorf("invalid project code or project is inactive: %w", err)
	}

	// 检查是否已经加入
	var existingMember models.ProjectMember
	err := s.db.Where("project_id = ? AND student_id = ?", project.ID, studentID).First(&existingMember).Error
	if err == nil {
		// 如果是inactive状态，重新激活
		if existingMember.Status == "inactive" {
			existingMember.Status = "active"
			existingMember.JoinedAt = time.Now()
			if err := s.db.Save(&existingMember).Error; err != nil {
				return nil, fmt.Errorf("failed to reactivate membership: %w", err)
			}
		}
		return &project, nil
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}

	// 创建新的成员关系
	member := &models.ProjectMember{
		ProjectID: project.ID,
		StudentID: studentID,
		Role:      "member",
		Status:    "active",
		JoinedAt:  time.Now(),
	}

	if err := s.db.Create(member).Error; err != nil {
		return nil, fmt.Errorf("failed to join project: %w", err)
	}

	// TODO: 发送通知给老师（待通知服务实现）

	// 重新加载课题数据
	if err := s.db.Preload("Teacher").Preload("Class").First(&project, project.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload project: %w", err)
	}

	return &project, nil
}

// AddStudentToProject 老师添加学生到课题
func (s *ProjectService) AddStudentToProject(projectID uint, studentID uint, role string) error {
	if role == "" {
		role = "member"
	}

	// 检查学生是否已经在课题中
	var existingMember models.ProjectMember
	err := s.db.Where("project_id = ? AND student_id = ?", projectID, studentID).First(&existingMember).Error
	if err == nil {
		if existingMember.Status == "inactive" {
			existingMember.Status = "active"
			existingMember.Role = role
			existingMember.JoinedAt = time.Now()
			return s.db.Save(&existingMember).Error
		}
		return nil // 已经是活跃成员
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check membership: %w", err)
	}

	// 创建新的成员关系
	member := &models.ProjectMember{
		ProjectID: projectID,
		StudentID: studentID,
		Role:      role,
		Status:    "active",
		JoinedAt:  time.Now(),
	}

	if err := s.db.Create(member).Error; err != nil {
		return fmt.Errorf("failed to add student to project: %w", err)
	}

	// TODO: 发送通知给学生（待通知服务实现）

	return nil
}

// RemoveStudentFromProject 从课题移除学生
func (s *ProjectService) RemoveStudentFromProject(projectID uint, studentID uint) error {
	var member models.ProjectMember
	if err := s.db.Where("project_id = ? AND student_id = ?", projectID, studentID).First(&member).Error; err != nil {
		return fmt.Errorf("student not found in project: %w", err)
	}

	member.Status = "inactive"
	if err := s.db.Save(&member).Error; err != nil {
		return fmt.Errorf("failed to remove student from project: %w", err)
	}

	return nil
}

// UpdateStudentRole 更新学生在课题中的角色
func (s *ProjectService) UpdateStudentRole(projectID uint, studentID uint, role string) error {
	var member models.ProjectMember
	if err := s.db.Where("project_id = ? AND student_id = ? AND status = 'active'", projectID, studentID).First(&member).Error; err != nil {
		return fmt.Errorf("student not found in project: %w", err)
	}

	member.Role = role
	if err := s.db.Save(&member).Error; err != nil {
		return fmt.Errorf("failed to update student role: %w", err)
	}

	return nil
}

// GetProjectMembers 获取课题成员列表
func (s *ProjectService) GetProjectMembers(projectID uint) ([]models.User, error) {
	var students []models.User
	err := s.db.Joins("JOIN project_members ON users.id = project_members.student_id").
		Where("project_members.project_id = ? AND project_members.status = 'active'", projectID).
		Order("project_members.joined_at DESC").
		Find(&students).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get project members: %w", err)
	}

	return students, nil
}

// GetProjectStats 获取课题统计信息
func (s *ProjectService) GetProjectStats(projectID uint) (*models.ProjectStats, error) {
	stats := &models.ProjectStats{}

	// 成员数量
	var memberCount int64
	if err := s.db.Model(&models.ProjectMember{}).
		Where("project_id = ? AND status = 'active'", projectID).
		Count(&memberCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count members: %w", err)
	}
	stats.MemberCount = int(memberCount)

	// 作业数量
	var assignmentCount int64
	if err := s.db.Model(&models.Assignment{}).
		Where("project_id = ?", projectID).
		Count(&assignmentCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count assignments: %w", err)
	}
	stats.AssignmentCount = int(assignmentCount)

	// 已完成作业数量
	var completedCount int64
	if err := s.db.Model(&models.Assignment{}).
		Where("project_id = ? AND status = 'closed'", projectID).
		Count(&completedCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count completed assignments: %w", err)
	}
	stats.CompletedAssignments = int(completedCount)

	// 待完成作业数量
	stats.PendingAssignments = stats.AssignmentCount - stats.CompletedAssignments

	return stats, nil
}

// generateProjectCode 生成唯一的课题代码
func (s *ProjectService) generateProjectCode() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const codeLength = 8

	for attempts := 0; attempts < 10; attempts++ {
		code := make([]byte, codeLength)
		for i := range code {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
			if err != nil {
				return "", err
			}
			code[i] = charset[n.Int64()]
		}

		codeStr := string(code)

		// 检查代码是否已存在
		var existingProject models.Project
		if err := s.db.Where("code = ?", codeStr).First(&existingProject).Error; err == gorm.ErrRecordNotFound {
			return codeStr, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique project code after 10 attempts")
}
