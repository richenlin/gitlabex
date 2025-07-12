package services

import (
	"fmt"
	"strings"

	"gitlabex/internal/models"

	"gorm.io/gorm"
)

// UserService 用户管理服务
type UserService struct {
	db                *gorm.DB
	permissionService *PermissionService
}

// NewUserService 创建用户管理服务
func NewUserService(db *gorm.DB, permissionService *PermissionService) *UserService {
	return &UserService{
		db:                db,
		permissionService: permissionService,
	}
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Avatar string `json:"avatar"`
	Role   *int   `json:"role"`
	Active *bool  `json:"active"`
}

// UserSearchRequest 用户搜索请求
type UserSearchRequest struct {
	Keyword  string `json:"keyword"`
	Role     *int   `json:"role"`
	Active   *bool  `json:"active"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

// UserStats 用户统计
type UserStats struct {
	TotalUsers    int `json:"total_users"`
	ActiveUsers   int `json:"active_users"`
	InactiveUsers int `json:"inactive_users"`
	AdminCount    int `json:"admin_count"`
	TeacherCount  int `json:"teacher_count"`
	StudentCount  int `json:"student_count"`
	GuestCount    int `json:"guest_count"`
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

// GetUserByUsername 根据用户名获取用户
func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

// GetAllUsers 获取所有用户（管理员权限）
func (s *UserService) GetAllUsers(page, pageSize int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// 计算总数
	if err := s.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := s.db.Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&users).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get users: %w", err)
	}

	return users, total, nil
}

// SearchUsers 搜索用户
func (s *UserService) SearchUsers(req *UserSearchRequest) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := s.db.Model(&models.User{})

	// 关键词搜索
	if req.Keyword != "" {
		keyword := "%" + strings.ToLower(req.Keyword) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(username) LIKE ?",
			keyword, keyword, keyword)
	}

	// 角色过滤
	if req.Role != nil {
		query = query.Where("role = ?", *req.Role)
	}

	// 活跃状态过滤
	if req.Active != nil {
		query = query.Where("active = ?", *req.Active)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// 分页查询
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize
	err := query.Order("created_at DESC").
		Limit(req.PageSize).
		Offset(offset).
		Find(&users).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to search users: %w", err)
	}

	return users, total, nil
}

// GetUsersByRole 根据角色获取用户
func (s *UserService) GetUsersByRole(role int) ([]models.User, error) {
	var users []models.User
	err := s.db.Where("role = ? AND active = true", role).
		Order("name ASC").
		Find(&users).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get users by role: %w", err)
	}

	return users, nil
}

// GetTeachers 获取所有老师
func (s *UserService) GetTeachers() ([]models.User, error) {
	return s.GetUsersByRole(RoleTeacher)
}

// GetStudents 获取所有学生
func (s *UserService) GetStudents() ([]models.User, error) {
	return s.GetUsersByRole(RoleStudent)
}

// GetAdmins 获取所有管理员
func (s *UserService) GetAdmins() ([]models.User, error) {
	return s.GetUsersByRole(RoleAdmin)
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(userID uint, req *UpdateUserRequest) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 更新字段
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		// 检查邮箱是否已存在
		var existingUser models.User
		if err := s.db.Where("email = ? AND id != ?", req.Email, userID).First(&existingUser).Error; err == nil {
			return nil, fmt.Errorf("email already exists")
		}
		user.Email = req.Email
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Role != nil {
		if !IsValidRole(*req.Role) {
			return nil, fmt.Errorf("invalid role: %d", *req.Role)
		}
		user.Role = *req.Role
	}
	if req.Active != nil {
		user.Active = *req.Active
	}

	if err := s.db.Save(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &user, nil
}

// UpdateUserRole 更新用户角色
func (s *UserService) UpdateUserRole(userID uint, role int) error {
	if !IsValidRole(role) {
		return fmt.Errorf("invalid role: %d", role)
	}

	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	user.Role = role
	if err := s.db.Save(&user).Error; err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}

	return nil
}

// DeactivateUser 停用用户
func (s *UserService) DeactivateUser(userID uint) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	user.Active = false
	if err := s.db.Save(&user).Error; err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	return nil
}

// ActivateUser 激活用户
func (s *UserService) ActivateUser(userID uint) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	user.Active = true
	if err := s.db.Save(&user).Error; err != nil {
		return fmt.Errorf("failed to activate user: %w", err)
	}

	return nil
}

// DeleteUser 删除用户（软删除）
func (s *UserService) DeleteUser(userID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 检查用户是否存在
		var user models.User
		if err := tx.First(&user, userID).Error; err != nil {
			return fmt.Errorf("user not found: %w", err)
		}

		// 移除用户的所有关联关系
		// 移除班级成员关系
		if err := tx.Where("student_id = ?", userID).Delete(&models.ClassMember{}).Error; err != nil {
			return fmt.Errorf("failed to remove class memberships: %w", err)
		}

		// 移除课题成员关系
		if err := tx.Where("student_id = ?", userID).Delete(&models.ProjectMember{}).Error; err != nil {
			return fmt.Errorf("failed to remove project memberships: %w", err)
		}

		// 移除用户的通知
		if err := tx.Where("user_id = ?", userID).Delete(&models.Notification{}).Error; err != nil {
			return fmt.Errorf("failed to remove notifications: %w", err)
		}

		// 软删除用户
		if err := tx.Delete(&user).Error; err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}

		return nil
	})
}

// GetUserStats 获取用户统计信息
func (s *UserService) GetUserStats() (*UserStats, error) {
	stats := &UserStats{}

	// 总用户数
	var totalUsers int64
	if err := s.db.Model(&models.User{}).Count(&totalUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count total users: %w", err)
	}
	stats.TotalUsers = int(totalUsers)

	// 活跃用户数
	var activeUsers int64
	if err := s.db.Model(&models.User{}).Where("active = true").Count(&activeUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count active users: %w", err)
	}
	stats.ActiveUsers = int(activeUsers)

	// 非活跃用户数
	stats.InactiveUsers = stats.TotalUsers - stats.ActiveUsers

	// 按角色统计
	var adminCount int64
	if err := s.db.Model(&models.User{}).Where("role = ?", RoleAdmin).Count(&adminCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count admins: %w", err)
	}
	stats.AdminCount = int(adminCount)

	var teacherCount int64
	if err := s.db.Model(&models.User{}).Where("role = ?", RoleTeacher).Count(&teacherCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count teachers: %w", err)
	}
	stats.TeacherCount = int(teacherCount)

	var studentCount int64
	if err := s.db.Model(&models.User{}).Where("role = ?", RoleStudent).Count(&studentCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count students: %w", err)
	}
	stats.StudentCount = int(studentCount)

	var guestCount int64
	if err := s.db.Model(&models.User{}).Where("role = ?", RoleGuest).Count(&guestCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count guests: %w", err)
	}
	stats.GuestCount = int(guestCount)

	return stats, nil
}

// GetUserProfile 获取用户完整档案
func (s *UserService) GetUserProfile(userID uint) (*UserProfile, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	profile := &UserProfile{
		User: user,
	}

	// 如果是学生，获取参加的班级和课题
	if user.Role == RoleStudent {
		// 获取参加的班级
		var classes []models.Class
		if err := s.db.Preload("Teacher").
			Joins("JOIN class_members ON classes.id = class_members.class_id").
			Where("class_members.student_id = ? AND class_members.status = 'active'", userID).
			Find(&classes).Error; err == nil {
			profile.Classes = classes
		}

		// 获取参加的课题
		var projects []models.Project
		if err := s.db.Preload("Teacher").
			Joins("JOIN project_members ON projects.id = project_members.project_id").
			Where("project_members.student_id = ? AND project_members.status = 'active'", userID).
			Find(&projects).Error; err == nil {
			profile.Projects = projects
		}

		// 获取作业提交统计
		var submissionCount int64
		if err := s.db.Model(&models.AssignmentSubmission{}).
			Where("student_id = ?", userID).Count(&submissionCount).Error; err == nil {
			profile.SubmissionCount = int(submissionCount)
		}

		// 获取已评审的作业数
		var reviewedCount int64
		if err := s.db.Model(&models.AssignmentSubmission{}).
			Where("student_id = ? AND status = 'reviewed'", userID).Count(&reviewedCount).Error; err == nil {
			profile.ReviewedCount = int(reviewedCount)
		}

		// 计算平均分
		if profile.ReviewedCount > 0 {
			var averageScore float64
			if err := s.db.Model(&models.AssignmentSubmission{}).
				Where("student_id = ? AND status = 'reviewed'", userID).
				Select("AVG(score)").Scan(&averageScore).Error; err == nil {
				profile.AverageScore = averageScore
			}
		}
	}

	// 如果是老师，获取创建的班级和课题
	if user.Role == RoleTeacher {
		// 获取创建的班级
		var classes []models.Class
		if err := s.db.Where("teacher_id = ?", userID).Find(&classes).Error; err == nil {
			profile.Classes = classes
		}

		// 获取创建的课题
		var projects []models.Project
		if err := s.db.Where("teacher_id = ?", userID).Find(&projects).Error; err == nil {
			profile.Projects = projects
		}

		// 获取创建的作业数
		var assignmentCount int64
		if err := s.db.Model(&models.Assignment{}).
			Where("teacher_id = ?", userID).Count(&assignmentCount).Error; err == nil {
			profile.AssignmentCount = int(assignmentCount)
		}

		// 获取已完成的评审数
		var reviewCount int64
		if err := s.db.Model(&models.Review{}).
			Where("teacher_id = ?", userID).Count(&reviewCount).Error; err == nil {
			profile.ReviewCount = int(reviewCount)
		}
	}

	return profile, nil
}

// UserProfile 用户档案
type UserProfile struct {
	User            models.User      `json:"user"`
	Classes         []models.Class   `json:"classes,omitempty"`
	Projects        []models.Project `json:"projects,omitempty"`
	SubmissionCount int              `json:"submission_count,omitempty"`
	ReviewedCount   int              `json:"reviewed_count,omitempty"`
	AverageScore    float64          `json:"average_score,omitempty"`
	AssignmentCount int              `json:"assignment_count,omitempty"`
	ReviewCount     int              `json:"review_count,omitempty"`
}

// ValidateUserPermission 验证用户权限
func (s *UserService) ValidateUserPermission(userID uint, permission string, resourceType string, resourceID uint) (bool, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return false, err
	}

	// 管理员拥有所有权限
	if user.Role == RoleAdmin {
		return true, nil
	}

	// 根据资源类型检查权限
	switch resourceType {
	case "class":
		return s.permissionService.CanAccessClass(user, resourceID, permission), nil
	case "project":
		return s.permissionService.CanAccessProject(user, resourceID, permission), nil
	case "assignment":
		return s.permissionService.CanAccessAssignment(user, resourceID, permission), nil
	}

	return false, nil
}

// GetCurrentUser 获取当前用户（用于测试环境）
func (s *UserService) GetCurrentUser() (*models.User, error) {
	// 在生产环境中，这个方法应该从上下文中获取当前用户
	// 这里我们返回一个测试用户
	var user models.User
	if err := s.db.Where("role = ?", RoleAdmin).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}
	return &user, nil
}

// GetUserDashboard 获取用户仪表板数据
func (s *UserService) GetUserDashboard(user *models.User) (*UserDashboard, error) {
	dashboard := &UserDashboard{
		User: *user,
	}

	// 根据用户角色获取不同的仪表板数据
	switch user.Role {
	case RoleAdmin:
		// 管理员看到所有统计数据
		var totalUsers, totalClasses, totalProjects, totalAssignments int64
		s.db.Model(&models.User{}).Count(&totalUsers)
		s.db.Model(&models.Class{}).Count(&totalClasses)
		s.db.Model(&models.Project{}).Count(&totalProjects)
		s.db.Model(&models.Assignment{}).Count(&totalAssignments)

		dashboard.TotalUsers = int(totalUsers)
		dashboard.TotalClasses = int(totalClasses)
		dashboard.TotalProjects = int(totalProjects)
		dashboard.TotalAssignments = int(totalAssignments)

	case RoleTeacher:
		// 老师看到自己的班级和课题统计
		var myClasses, myProjects, myAssignments int64
		s.db.Model(&models.Class{}).Where("teacher_id = ?", user.ID).Count(&myClasses)
		s.db.Model(&models.Project{}).Where("teacher_id = ?", user.ID).Count(&myProjects)
		s.db.Model(&models.Assignment{}).Where("teacher_id = ?", user.ID).Count(&myAssignments)

		dashboard.MyClasses = int(myClasses)
		dashboard.MyProjects = int(myProjects)
		dashboard.MyAssignments = int(myAssignments)

	case RoleStudent:
		// 学生看到自己参与的统计
		var myClasses, myProjects, mySubmissions int64
		s.db.Model(&models.Class{}).
			Joins("JOIN class_members ON classes.id = class_members.class_id").
			Where("class_members.student_id = ?", user.ID).Count(&myClasses)
		s.db.Model(&models.Project{}).
			Joins("JOIN project_members ON projects.id = project_members.project_id").
			Where("project_members.student_id = ?", user.ID).Count(&myProjects)
		s.db.Model(&models.AssignmentSubmission{}).
			Where("student_id = ?", user.ID).Count(&mySubmissions)

		dashboard.MyClasses = int(myClasses)
		dashboard.MyProjects = int(myProjects)
		dashboard.MySubmissions = int(mySubmissions)
	}

	return dashboard, nil
}

// ToProfile 将用户转换为用户资料
func (s *UserService) ToProfile(user *models.User, role models.EducationRole) *UserProfile {
	profile := &UserProfile{
		User: *user,
	}

	// 根据角色添加相关统计信息
	switch user.Role {
	case RoleTeacher:
		// 获取创建的班级
		var classes []models.Class
		if err := s.db.Where("teacher_id = ?", user.ID).Find(&classes).Error; err == nil {
			profile.Classes = classes
		}

		// 获取创建的课题
		var projects []models.Project
		if err := s.db.Where("teacher_id = ?", user.ID).Find(&projects).Error; err == nil {
			profile.Projects = projects
		}

		// 获取创建的作业数
		var assignmentCount int64
		if err := s.db.Model(&models.Assignment{}).Where("teacher_id = ?", user.ID).Count(&assignmentCount).Error; err == nil {
			profile.AssignmentCount = int(assignmentCount)
		}

	case RoleStudent:
		// 获取参加的班级
		var classes []models.Class
		if err := s.db.Preload("Teacher").
			Joins("JOIN class_members ON classes.id = class_members.class_id").
			Where("class_members.student_id = ? AND class_members.status = 'active'", user.ID).
			Find(&classes).Error; err == nil {
			profile.Classes = classes
		}

		// 获取参加的课题
		var projects []models.Project
		if err := s.db.Preload("Teacher").
			Joins("JOIN project_members ON projects.id = project_members.project_id").
			Where("project_members.student_id = ? AND project_members.status = 'active'", user.ID).
			Find(&projects).Error; err == nil {
			profile.Projects = projects
		}

		// 获取作业提交统计
		var submissionCount int64
		if err := s.db.Model(&models.AssignmentSubmission{}).
			Where("student_id = ?", user.ID).Count(&submissionCount).Error; err == nil {
			profile.SubmissionCount = int(submissionCount)
		}

		// 获取已评审的作业数
		var reviewedCount int64
		if err := s.db.Model(&models.AssignmentSubmission{}).
			Where("student_id = ? AND status = 'graded'", user.ID).Count(&reviewedCount).Error; err == nil {
			profile.ReviewedCount = int(reviewedCount)
		}

		// 计算平均分
		if profile.ReviewedCount > 0 {
			var averageScore float64
			if err := s.db.Model(&models.AssignmentSubmission{}).
				Where("student_id = ? AND status = 'graded'", user.ID).
				Select("AVG(score)").Scan(&averageScore).Error; err == nil {
				profile.AverageScore = averageScore
			}
		}
	}

	return profile
}

// ListActiveUsers 获取活跃用户列表
func (s *UserService) ListActiveUsers() ([]*models.User, error) {
	var users []*models.User
	err := s.db.Where("active = true").Order("name ASC").Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list active users: %w", err)
	}
	return users, nil
}

// UpdateUserFields 更新用户信息 (兼容handler期望的签名)
func (s *UserService) UpdateUserFields(user *models.User, updates map[string]interface{}) error {
	return s.db.Model(user).Updates(updates).Error
}

// UserDashboard 用户仪表板数据
type UserDashboard struct {
	User             models.User `json:"user"`
	TotalUsers       int         `json:"total_users,omitempty"`
	TotalClasses     int         `json:"total_classes,omitempty"`
	TotalProjects    int         `json:"total_projects,omitempty"`
	TotalAssignments int         `json:"total_assignments,omitempty"`
	MyClasses        int         `json:"my_classes,omitempty"`
	MyProjects       int         `json:"my_projects,omitempty"`
	MyAssignments    int         `json:"my_assignments,omitempty"`
	MySubmissions    int         `json:"my_submissions,omitempty"`
	RecentActivities []string    `json:"recent_activities,omitempty"`
}
