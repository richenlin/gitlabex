package services

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"gitlabex/internal/models"

	"gorm.io/gorm"
)

// ClassService 班级管理服务
type ClassService struct {
	db                *gorm.DB
	permissionService *PermissionService
	// TODO: 添加 notificationService 当实现后
}

// NewClassService 创建班级管理服务
func NewClassService(db *gorm.DB, permissionService *PermissionService) *ClassService {
	return &ClassService{
		db:                db,
		permissionService: permissionService,
	}
}

// CreateClassRequest 创建班级请求
type CreateClassRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// UpdateClassRequest 更新班级请求
type UpdateClassRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Active      *bool  `json:"active"`
}

// JoinClassRequest 加入班级请求
type JoinClassRequest struct {
	Code string `json:"code" binding:"required"`
}

// CreateClass 创建班级（老师权限）
func (s *ClassService) CreateClass(teacherID uint, req *CreateClassRequest) (*models.Class, error) {
	// 生成唯一的班级代码
	code, err := s.generateClassCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate class code: %w", err)
	}

	class := &models.Class{
		Name:        req.Name,
		Description: req.Description,
		Code:        code,
		TeacherID:   teacherID,
		Active:      true,
	}

	if err := s.db.Create(class).Error; err != nil {
		return nil, fmt.Errorf("failed to create class: %w", err)
	}

	// 预加载关联数据
	if err := s.db.Preload("Teacher").First(class, class.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load class: %w", err)
	}

	return class, nil
}

// GetClassByID 根据ID获取班级
func (s *ClassService) GetClassByID(classID uint) (*models.Class, error) {
	var class models.Class
	err := s.db.Preload("Teacher").
		Preload("Members").
		Preload("Students").
		First(&class, classID).Error

	if err != nil {
		return nil, fmt.Errorf("class not found: %w", err)
	}

	return &class, nil
}

// GetClassesByTeacher 获取老师创建的班级列表
func (s *ClassService) GetClassesByTeacher(teacherID uint) ([]models.Class, error) {
	var classes []models.Class
	err := s.db.Preload("Teacher").
		Where("teacher_id = ?", teacherID).
		Order("created_at DESC").
		Find(&classes).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get classes: %w", err)
	}

	return classes, nil
}

// GetClassesByStudent 获取学生加入的班级列表
func (s *ClassService) GetClassesByStudent(studentID uint) ([]models.Class, error) {
	var classes []models.Class
	err := s.db.Preload("Teacher").
		Joins("JOIN class_members ON classes.id = class_members.class_id").
		Where("class_members.student_id = ? AND class_members.status = 'active'", studentID).
		Order("class_members.joined_at DESC").
		Find(&classes).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get classes: %w", err)
	}

	return classes, nil
}

// GetAllClasses 获取所有班级（管理员权限）
func (s *ClassService) GetAllClasses(page, pageSize int) ([]models.Class, int64, error) {
	var classes []models.Class
	var total int64

	// 计算总数
	if err := s.db.Model(&models.Class{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count classes: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := s.db.Preload("Teacher").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&classes).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get classes: %w", err)
	}

	return classes, total, nil
}

// UpdateClass 更新班级信息
func (s *ClassService) UpdateClass(classID uint, req *UpdateClassRequest) (*models.Class, error) {
	var class models.Class
	if err := s.db.First(&class, classID).Error; err != nil {
		return nil, fmt.Errorf("class not found: %w", err)
	}

	// 更新字段
	if req.Name != "" {
		class.Name = req.Name
	}
	if req.Description != "" {
		class.Description = req.Description
	}
	if req.Active != nil {
		class.Active = *req.Active
	}

	if err := s.db.Save(&class).Error; err != nil {
		return nil, fmt.Errorf("failed to update class: %w", err)
	}

	// 重新加载数据
	if err := s.db.Preload("Teacher").First(&class, class.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload class: %w", err)
	}

	return &class, nil
}

// DeleteClass 删除班级
func (s *ClassService) DeleteClass(classID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除班级成员关系
		if err := tx.Where("class_id = ?", classID).Delete(&models.ClassMember{}).Error; err != nil {
			return fmt.Errorf("failed to delete class members: %w", err)
		}

		// 删除班级
		if err := tx.Delete(&models.Class{}, classID).Error; err != nil {
			return fmt.Errorf("failed to delete class: %w", err)
		}

		return nil
	})
}

// JoinClass 学生加入班级
func (s *ClassService) JoinClass(studentID uint, code string) (*models.Class, error) {
	// 查找班级
	var class models.Class
	if err := s.db.Where("code = ? AND active = true", code).First(&class).Error; err != nil {
		return nil, fmt.Errorf("invalid class code or class is inactive: %w", err)
	}

	// 检查是否已经加入
	var existingMember models.ClassMember
	err := s.db.Where("class_id = ? AND student_id = ?", class.ID, studentID).First(&existingMember).Error
	if err == nil {
		// 如果是inactive状态，重新激活
		if existingMember.Status == "inactive" {
			existingMember.Status = "active"
			existingMember.JoinedAt = time.Now()
			if err := s.db.Save(&existingMember).Error; err != nil {
				return nil, fmt.Errorf("failed to reactivate membership: %w", err)
			}
		}
		return &class, nil
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}

	// 创建新的成员关系
	member := &models.ClassMember{
		ClassID:   class.ID,
		StudentID: studentID,
		Status:    "active",
		JoinedAt:  time.Now(),
	}

	if err := s.db.Create(member).Error; err != nil {
		return nil, fmt.Errorf("failed to join class: %w", err)
	}

	// TODO: 发送通知给老师（待通知服务实现）

	// 重新加载班级数据
	if err := s.db.Preload("Teacher").First(&class, class.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload class: %w", err)
	}

	return &class, nil
}

// AddStudentToClass 老师添加学生到班级
func (s *ClassService) AddStudentToClass(classID uint, studentID uint) error {
	// 检查学生是否已经在班级中
	var existingMember models.ClassMember
	err := s.db.Where("class_id = ? AND student_id = ?", classID, studentID).First(&existingMember).Error
	if err == nil {
		if existingMember.Status == "inactive" {
			existingMember.Status = "active"
			existingMember.JoinedAt = time.Now()
			return s.db.Save(&existingMember).Error
		}
		return nil // 已经是活跃成员
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check membership: %w", err)
	}

	// 创建新的成员关系
	member := &models.ClassMember{
		ClassID:   classID,
		StudentID: studentID,
		Status:    "active",
		JoinedAt:  time.Now(),
	}

	if err := s.db.Create(member).Error; err != nil {
		return fmt.Errorf("failed to add student to class: %w", err)
	}

	// TODO: 发送通知给学生（待通知服务实现）

	return nil
}

// RemoveStudentFromClass 从班级移除学生
func (s *ClassService) RemoveStudentFromClass(classID uint, studentID uint) error {
	var member models.ClassMember
	if err := s.db.Where("class_id = ? AND student_id = ?", classID, studentID).First(&member).Error; err != nil {
		return fmt.Errorf("student not found in class: %w", err)
	}

	member.Status = "inactive"
	if err := s.db.Save(&member).Error; err != nil {
		return fmt.Errorf("failed to remove student from class: %w", err)
	}

	return nil
}

// GetClassMembers 获取班级成员列表
func (s *ClassService) GetClassMembers(classID uint) ([]models.User, error) {
	var students []models.User
	err := s.db.Joins("JOIN class_members ON users.id = class_members.student_id").
		Where("class_members.class_id = ? AND class_members.status = 'active'", classID).
		Order("class_members.joined_at DESC").
		Find(&students).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get class members: %w", err)
	}

	return students, nil
}

// GetClassStats 获取班级统计信息
func (s *ClassService) GetClassStats(classID uint) (*models.ClassStats, error) {
	stats := &models.ClassStats{}

	// 学生数量
	var studentCount int64
	if err := s.db.Model(&models.ClassMember{}).
		Where("class_id = ? AND status = 'active'", classID).
		Count(&studentCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count students: %w", err)
	}
	stats.StudentCount = int(studentCount)

	// 课题数量
	var projectCount int64
	if err := s.db.Model(&models.Project{}).
		Where("class_id = ?", classID).
		Count(&projectCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count projects: %w", err)
	}
	stats.ProjectCount = int(projectCount)

	// 活跃课题数量
	var activeProjectCount int64
	if err := s.db.Model(&models.Project{}).
		Where("class_id = ? AND status = 'active'", classID).
		Count(&activeProjectCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count active projects: %w", err)
	}
	stats.ActiveProjects = int(activeProjectCount)

	return stats, nil
}

// generateClassCode 生成唯一的班级代码
func (s *ClassService) generateClassCode() (string, error) {
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
		var existingClass models.Class
		if err := s.db.Where("code = ?", codeStr).First(&existingClass).Error; err == gorm.ErrRecordNotFound {
			return codeStr, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique class code after 10 attempts")
}
