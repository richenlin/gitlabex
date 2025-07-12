package services

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"gitlabex/internal/models"

	"github.com/xanzy/go-gitlab"

	"gorm.io/gorm"
)

// ClassService 班级管理服务 - 集成GitLab Group功能
type ClassService struct {
	db                *gorm.DB
	permissionService *PermissionService
	gitlabService     *GitLabService
	// TODO: 添加 notificationService 当实现后
}

// NewClassService 创建班级管理服务
func NewClassService(db *gorm.DB, permissionService *PermissionService, gitlabService *GitLabService) *ClassService {
	return &ClassService{
		db:                db,
		permissionService: permissionService,
		gitlabService:     gitlabService,
	}
}

// CreateClassRequest 创建班级请求
type CreateClassRequest struct {
	Name             string `json:"name" binding:"required"`
	Description      string `json:"description"`
	EnableGitLabSync bool   `json:"enable_gitlab_sync"` // 是否启用GitLab同步
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

// CreateClass 创建班级（老师权限）- 同时创建GitLab Group
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
		SyncStatus:  "pending",
	}

	// 开始数据库事务
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 创建班级记录
		if err := tx.Create(class).Error; err != nil {
			return fmt.Errorf("failed to create class: %w", err)
		}

		// 如果启用GitLab同步，创建GitLab Group
		if req.EnableGitLabSync && s.gitlabService != nil {
			if err := s.createGitLabGroup(tx, class); err != nil {
				// 设置同步错误状态，但不阻止班级创建
				class.SetSyncError(err)
				if updateErr := tx.Save(class).Error; updateErr != nil {
					return fmt.Errorf("failed to update class sync status: %w", updateErr)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 预加载关联数据
	if err := s.db.Preload("Teacher").First(class, class.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load class: %w", err)
	}

	return class, nil
}

// createGitLabGroup 创建GitLab Group
func (s *ClassService) createGitLabGroup(tx *gorm.DB, class *models.Class) error {
	// 生成Group路径（使用班级代码确保唯一性）
	groupPath := fmt.Sprintf("class-%s", class.Code)
	groupName := fmt.Sprintf("%s (%s)", class.Name, class.Code)

	// 创建GitLab Group
	group, err := s.gitlabService.CreateGroup(groupName, groupPath, class.Description, nil)
	if err != nil {
		return fmt.Errorf("failed to create GitLab group: %w", err)
	}

	// 更新班级的GitLab信息
	class.SetGitLabGroup(group.ID, group.Path, group.WebURL)

	// 获取教师用户信息，将教师添加为Group Owner
	var teacher models.User
	if err := tx.First(&teacher, class.TeacherID).Error; err != nil {
		return fmt.Errorf("failed to find teacher: %w", err)
	}

	// 添加教师到GitLab Group
	if teacher.GitLabID > 0 {
		_, _, err = s.gitlabService.client.GroupMembers.AddGroupMember(group.ID, &gitlab.AddGroupMemberOptions{
			UserID:      gitlab.Int(teacher.GitLabID),
			AccessLevel: gitlab.AccessLevel(gitlab.OwnerPermissions),
		})
		if err != nil {
			return fmt.Errorf("failed to add teacher to GitLab group: %w", err)
		}
	}

	return nil
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

	// 如果班级已同步到GitLab，同步更新GitLab Group
	if class.IsGitLabSynced() && s.gitlabService != nil {
		if err := s.updateGitLabGroup(&class, req); err != nil {
			// 记录错误但不阻止本地更新
			class.SetSyncError(err)
		}
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

// updateGitLabGroup 更新GitLab Group信息
func (s *ClassService) updateGitLabGroup(class *models.Class, req *UpdateClassRequest) error {
	groupID := class.GetGitLabGroupID()
	if groupID == 0 {
		return fmt.Errorf("GitLab group not found")
	}

	// 构建更新参数
	updateOpts := &gitlab.UpdateGroupOptions{}
	if req.Name != "" {
		updateOpts.Name = gitlab.String(fmt.Sprintf("%s (%s)", req.Name, class.Code))
	}
	if req.Description != "" {
		updateOpts.Description = gitlab.String(req.Description)
	}

	// 更新GitLab Group
	_, _, err := s.gitlabService.client.Groups.UpdateGroup(groupID, updateOpts)
	return err
}

// DeleteClass 删除班级 - 同时删除GitLab Group
func (s *ClassService) DeleteClass(classID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取班级信息
		var class models.Class
		if err := tx.First(&class, classID).Error; err != nil {
			return fmt.Errorf("class not found: %w", err)
		}

		// 如果班级已同步到GitLab，删除GitLab Group
		if class.IsGitLabSynced() && s.gitlabService != nil {
			if err := s.deleteGitLabGroup(&class); err != nil {
				// 记录错误但继续删除本地数据
				fmt.Printf("Warning: Failed to delete GitLab group: %v\n", err)
			}
		}

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

// deleteGitLabGroup 删除GitLab Group
func (s *ClassService) deleteGitLabGroup(class *models.Class) error {
	groupID := class.GetGitLabGroupID()
	if groupID == 0 {
		return nil
	}

	_, err := s.gitlabService.client.Groups.DeleteGroup(groupID)
	return err
}

// JoinClass 学生加入班级 - 同时加入GitLab Group
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
			existingMember.GitLabSyncStatus = "pending" // 重置同步状态
			if err := s.db.Save(&existingMember).Error; err != nil {
				return nil, fmt.Errorf("failed to reactivate membership: %w", err)
			}

			// 同步到GitLab Group
			if class.IsGitLabSynced() {
				go s.syncMemberToGitLab(&class, &existingMember)
			}
		}
		return &class, nil
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}

	// 创建新的成员关系
	member := &models.ClassMember{
		ClassID:          class.ID,
		StudentID:        studentID,
		Status:           "active",
		JoinedAt:         time.Now(),
		GitLabRole:       "reporter", // 学生默认角色
		GitLabSyncStatus: "pending",
	}

	if err := s.db.Create(member).Error; err != nil {
		return nil, fmt.Errorf("failed to join class: %w", err)
	}

	// 异步同步到GitLab Group
	if class.IsGitLabSynced() {
		go s.syncMemberToGitLab(&class, member)
	}

	// 重新加载班级数据
	if err := s.db.Preload("Teacher").First(&class, class.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload class: %w", err)
	}

	return &class, nil
}

// syncMemberToGitLab 同步成员到GitLab Group
func (s *ClassService) syncMemberToGitLab(class *models.Class, member *models.ClassMember) {
	if s.gitlabService == nil {
		return
	}

	// 获取学生用户信息
	var student models.User
	if err := s.db.First(&student, member.StudentID).Error; err != nil {
		member.SetGitLabSyncError(fmt.Errorf("failed to find student: %w", err))
		s.db.Save(member)
		return
	}

	// 检查学生是否有GitLab ID
	if student.GitLabID == 0 {
		member.SetGitLabSyncError(fmt.Errorf("student has no GitLab ID"))
		s.db.Save(member)
		return
	}

	// 添加成员到GitLab Group
	groupID := class.GetGitLabGroupID()
	accessLevel := s.mapRoleToAccessLevel(member.GitLabRole)

	gitlabMember, _, err := s.gitlabService.client.GroupMembers.AddGroupMember(groupID, &gitlab.AddGroupMemberOptions{
		UserID:      gitlab.Int(student.GitLabID),
		AccessLevel: gitlab.AccessLevel(accessLevel),
	})

	if err != nil {
		member.SetGitLabSyncError(err)
	} else {
		member.SetGitLabMember(gitlabMember.ID, member.GitLabRole)
	}

	s.db.Save(member)
}

// mapRoleToAccessLevel 映射角色到GitLab访问级别
func (s *ClassService) mapRoleToAccessLevel(role string) gitlab.AccessLevelValue {
	switch role {
	case "guest":
		return gitlab.GuestPermissions
	case "reporter":
		return gitlab.ReporterPermissions
	case "developer":
		return gitlab.DeveloperPermissions
	case "maintainer":
		return gitlab.MaintainerPermissions
	case "owner":
		return gitlab.OwnerPermissions
	default:
		return gitlab.ReporterPermissions // 默认学生权限
	}
}

// AddStudentToClass 老师添加学生到班级 - 同时添加到GitLab Group
func (s *ClassService) AddStudentToClass(classID uint, studentID uint) error {
	// 获取班级信息
	var class models.Class
	if err := s.db.First(&class, classID).Error; err != nil {
		return fmt.Errorf("class not found: %w", err)
	}

	// 检查学生是否已经在班级中
	var existingMember models.ClassMember
	err := s.db.Where("class_id = ? AND student_id = ?", classID, studentID).First(&existingMember).Error
	if err == nil {
		if existingMember.Status == "inactive" {
			existingMember.Status = "active"
			existingMember.JoinedAt = time.Now()
			existingMember.GitLabSyncStatus = "pending"

			if err := s.db.Save(&existingMember).Error; err != nil {
				return err
			}

			// 同步到GitLab
			if class.IsGitLabSynced() {
				go s.syncMemberToGitLab(&class, &existingMember)
			}
		}
		return nil // 已经是活跃成员
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check membership: %w", err)
	}

	// 创建新的成员关系
	member := &models.ClassMember{
		ClassID:          classID,
		StudentID:        studentID,
		Status:           "active",
		JoinedAt:         time.Now(),
		GitLabRole:       "reporter",
		GitLabSyncStatus: "pending",
	}

	if err := s.db.Create(member).Error; err != nil {
		return fmt.Errorf("failed to add student to class: %w", err)
	}

	// 异步同步到GitLab
	if class.IsGitLabSynced() {
		go s.syncMemberToGitLab(&class, member)
	}

	return nil
}

// RemoveStudentFromClass 从班级移除学生 - 同时从GitLab Group移除
func (s *ClassService) RemoveStudentFromClass(classID uint, studentID uint) error {
	var member models.ClassMember
	if err := s.db.Where("class_id = ? AND student_id = ?", classID, studentID).First(&member).Error; err != nil {
		return fmt.Errorf("student not found in class: %w", err)
	}

	// 获取班级信息
	var class models.Class
	if err := s.db.First(&class, member.ClassID).Error; err != nil {
		return fmt.Errorf("class not found: %w", err)
	}

	// 如果已同步到GitLab，从GitLab Group移除
	if class.IsGitLabSynced() && member.IsGitLabSynced() {
		go s.removeMemberFromGitLab(&class, &member)
	}

	// 更新本地状态
	member.Status = "inactive"
	member.GitLabSyncStatus = "removed"
	if err := s.db.Save(&member).Error; err != nil {
		return fmt.Errorf("failed to remove student from class: %w", err)
	}

	return nil
}

// removeMemberFromGitLab 从GitLab Group移除成员
func (s *ClassService) removeMemberFromGitLab(class *models.Class, member *models.ClassMember) {
	if s.gitlabService == nil || !member.IsGitLabSynced() {
		return
	}

	groupID := class.GetGitLabGroupID()
	memberID := *member.GitLabMemberID

	_, err := s.gitlabService.client.GroupMembers.RemoveGroupMember(groupID, memberID, &gitlab.RemoveGroupMemberOptions{})
	if err != nil {
		// 记录错误但不阻止本地操作
		member.SetGitLabSyncError(err)
		s.db.Save(member)
	}
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

// GetClassStats 获取班级统计信息 - 包括GitLab统计
func (s *ClassService) GetClassStats(classID uint) (*models.ClassStats, error) {
	var class models.Class
	if err := s.db.First(&class, classID).Error; err != nil {
		return nil, fmt.Errorf("class not found: %w", err)
	}

	stats := &models.ClassStats{
		GitLabGroupID: class.GetGitLabGroupID(),
	}

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

	// 如果班级已同步到GitLab，获取GitLab统计
	if class.IsGitLabSynced() && s.gitlabService != nil {
		s.addGitLabStats(stats, &class)
	}

	return stats, nil
}

// addGitLabStats 添加GitLab统计信息
func (s *ClassService) addGitLabStats(stats *models.ClassStats, class *models.Class) {
	groupID := class.GetGitLabGroupID()

	// 获取Group项目数量
	if projects, _, err := s.gitlabService.client.Groups.ListGroupProjects(groupID, &gitlab.ListGroupProjectsOptions{}); err == nil {
		stats.GitLabProjectCount = len(projects)

		// 统计Issues和MR数量
		for _, project := range projects {
			if issues, _, err := s.gitlabService.client.Issues.ListProjectIssues(project.ID, &gitlab.ListProjectIssuesOptions{}); err == nil {
				stats.GitLabIssueCount += len(issues)
			}
			if mrs, _, err := s.gitlabService.client.MergeRequests.ListProjectMergeRequests(project.ID, &gitlab.ListProjectMergeRequestsOptions{}); err == nil {
				stats.GitLabMRCount += len(mrs)
			}
		}
	}
}

// SyncClassToGitLab 手动同步班级到GitLab Group
func (s *ClassService) SyncClassToGitLab(classID uint) error {
	var class models.Class
	if err := s.db.First(&class, classID).Error; err != nil {
		return fmt.Errorf("class not found: %w", err)
	}

	// 如果还未同步，创建GitLab Group
	if !class.IsGitLabSynced() {
		return s.db.Transaction(func(tx *gorm.DB) error {
			return s.createGitLabGroup(tx, &class)
		})
	}

	// 如果已同步，同步成员信息
	return s.syncClassMembers(&class)
}

// syncClassMembers 同步班级成员到GitLab Group
func (s *ClassService) syncClassMembers(class *models.Class) error {
	var members []models.ClassMember
	if err := s.db.Where("class_id = ? AND status = 'active'", class.ID).Find(&members).Error; err != nil {
		return fmt.Errorf("failed to get class members: %w", err)
	}

	for _, member := range members {
		if member.GitLabSyncStatus != "synced" {
			go s.syncMemberToGitLab(class, &member)
		}
	}

	return nil
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
