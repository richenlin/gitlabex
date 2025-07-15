package services

import (
	"fmt"
	"time"

	"gitlabex/internal/models"
	"gitlabex/internal/utils"

	"gorm.io/gorm"
)

// ProjectService 课题管理服务
type ProjectServiceV2 struct {
	db                *gorm.DB
	permissionService *PermissionService
	gitlabService     *GitLabService
}

// NewProjectServiceV2 创建课题管理服务V2
func NewProjectServiceV2(db *gorm.DB, permissionService *PermissionService, gitlabService *GitLabService) *ProjectServiceV2 {
	return &ProjectServiceV2{
		db:                db,
		permissionService: permissionService,
		gitlabService:     gitlabService,
	}
}

// CreateProjectRequestV2 创建课题请求V2
type CreateProjectRequestV2 struct {
	Name        string         `json:"name" binding:"required"`
	Description string         `json:"description"`
	Type        string         `json:"type"` // graduation, research, competition, practice
	StartDate   utils.DateOnly `json:"start_date"`
	EndDate     utils.DateOnly `json:"end_date"`
	MaxMembers  int            `json:"max_members"`
	// GitLab相关字段
	WikiEnabled   bool `json:"wiki_enabled"`
	IssuesEnabled bool `json:"issues_enabled"`
	MREnabled     bool `json:"mr_enabled"`
}

// UpdateProjectRequestV2 更新课题请求V2
type UpdateProjectRequestV2 struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Type        string          `json:"type"`
	StartDate   *utils.DateOnly `json:"start_date"`
	EndDate     *utils.DateOnly `json:"end_date"`
	MaxMembers  *int            `json:"max_members"`
	Status      string          `json:"status"`
	// GitLab相关字段
	WikiEnabled   *bool `json:"wiki_enabled"`
	IssuesEnabled *bool `json:"issues_enabled"`
	MREnabled     *bool `json:"mr_enabled"`
}

// ProjectSimpleV2 简化的项目结构V2
type ProjectSimpleV2 struct {
	ID                   uint      `json:"id"`
	Name                 string    `json:"name"`
	Description          string    `json:"description"`
	Code                 string    `json:"code"`
	TeacherID            uint      `json:"teacher_id"`
	TeacherName          string    `json:"teacher_name"`
	Status               string    `json:"status"`
	Type                 string    `json:"type"`
	StartDate            time.Time `json:"start_date"`
	EndDate              time.Time `json:"end_date"`
	MaxMembers           int       `json:"max_members"`
	TotalAssignments     int       `json:"total_assignments"`
	CompletedAssignments int       `json:"completed_assignments"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// CreateProject 创建课题
func (s *ProjectServiceV2) CreateProject(teacherID uint, req *CreateProjectRequestV2) (*models.Project, error) {
	// 获取教师信息
	var teacher models.User
	if err := s.db.First(&teacher, teacherID).Error; err != nil {
		return nil, fmt.Errorf("teacher not found: %w", err)
	}

	// 生成课题代码
	code := s.generateProjectCode()

	// 创建课题记录
	project := &models.Project{
		Name:          req.Name,
		Description:   req.Description,
		Code:          code,
		TeacherID:     teacherID,
		Type:          req.Type,
		Status:        "active",
		StartDate:     req.StartDate.Time,
		EndDate:       req.EndDate.Time,
		MaxMembers:    req.MaxMembers,
		WikiEnabled:   req.WikiEnabled,
		IssuesEnabled: req.IssuesEnabled,
		MREnabled:     req.MREnabled,
	}

	// 尝试创建GitLab项目（如果失败，仍继续创建本地项目）
	readmeContent := fmt.Sprintf("# %s\n\n%s\n\n## 课题信息\n\n- **课题类型**: %s\n- **开始时间**: %s\n- **结束时间**: %s\n- **最大成员数**: %d\n\n## 开发指南\n\n请在个人分支上进行开发，完成后提交合并请求。\n",
		req.Name, req.Description, req.Type, req.StartDate.String(), req.EndDate.String(), req.MaxMembers)

	gitlabProject, err := s.gitlabService.CreateProjectRepository(req.Name, req.Description, readmeContent, teacher.GitLabID)
	if err != nil {
		// GitLab创建失败，记录警告但继续创建本地项目
		fmt.Printf("WARNING: Failed to create GitLab project: %v\n", err)
	} else {
		// GitLab创建成功，更新项目信息
		project.GitLabProjectID = gitlabProject.ID
		project.GitLabURL = gitlabProject.WebURL
		project.RepositoryURL = gitlabProject.HTTPURLToRepo
		project.DefaultBranch = gitlabProject.DefaultBranch
	}

	// 保存到数据库
	if err := s.db.Create(project).Error; err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return project, nil
}

// GetProjectsByTeacher 获取教师的课题列表
func (s *ProjectServiceV2) GetProjectsByTeacher(teacherID uint) ([]ProjectSimpleV2, error) {
	var projects []models.Project

	err := s.db.Preload("Teacher").
		Where("teacher_id = ?", teacherID).
		Select("id, name, description, code, teacher_id, status, type, start_date, end_date, max_members, total_assignments, completed_assignments, created_at, updated_at").
		Find(&projects).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get projects by teacher: %w", err)
	}

	// 转换为简化的结构
	simpleProjects := make([]ProjectSimpleV2, len(projects))
	for i, project := range projects {
		simpleProjects[i] = ProjectSimpleV2{
			ID:                   project.ID,
			Name:                 project.Name,
			Description:          project.Description,
			Code:                 project.Code,
			TeacherID:            project.TeacherID,
			TeacherName:          project.Teacher.Name,
			Status:               project.Status,
			Type:                 project.Type,
			StartDate:            project.StartDate,
			EndDate:              project.EndDate,
			MaxMembers:           project.MaxMembers,
			TotalAssignments:     project.TotalAssignments,
			CompletedAssignments: project.CompletedAssignments,
			CreatedAt:            project.CreatedAt,
			UpdatedAt:            project.UpdatedAt,
		}
	}

	return simpleProjects, nil
}

// GetProjectsByStudent 获取学生参与的课题列表
func (s *ProjectServiceV2) GetProjectsByStudent(studentID uint) ([]ProjectSimpleV2, error) {
	var projects []models.Project

	// 通过project_members表关联查询
	err := s.db.Preload("Teacher").
		Joins("JOIN project_members ON projects.id = project_members.project_id").
		Where("project_members.user_id = ? AND project_members.is_active = true", studentID).
		Select("projects.id, projects.name, projects.description, projects.code, projects.teacher_id, projects.status, projects.type, projects.start_date, projects.end_date, projects.max_members, projects.total_assignments, projects.completed_assignments, projects.created_at, projects.updated_at").
		Find(&projects).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get projects by student: %w", err)
	}

	// 转换为简化的结构
	simpleProjects := make([]ProjectSimpleV2, len(projects))
	for i, project := range projects {
		simpleProjects[i] = ProjectSimpleV2{
			ID:                   project.ID,
			Name:                 project.Name,
			Description:          project.Description,
			Code:                 project.Code,
			TeacherID:            project.TeacherID,
			TeacherName:          project.Teacher.Name,
			Status:               project.Status,
			Type:                 project.Type,
			StartDate:            project.StartDate,
			EndDate:              project.EndDate,
			MaxMembers:           project.MaxMembers,
			TotalAssignments:     project.TotalAssignments,
			CompletedAssignments: project.CompletedAssignments,
			CreatedAt:            project.CreatedAt,
			UpdatedAt:            project.UpdatedAt,
		}
	}

	return simpleProjects, nil
}

// GetAllProjects 获取所有课题（分页）
func (s *ProjectServiceV2) GetAllProjects(page, pageSize int) ([]ProjectSimpleV2, int64, error) {
	var projects []models.Project
	var total int64

	// 计算总数
	if err := s.db.Model(&models.Project{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count projects: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := s.db.Preload("Teacher").
		Select("id, name, description, code, teacher_id, status, type, start_date, end_date, max_members, total_assignments, completed_assignments, created_at, updated_at").
		Limit(pageSize).
		Offset(offset).
		Order("created_at DESC").
		Find(&projects).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get all projects: %w", err)
	}

	// 转换为简化的结构
	simpleProjects := make([]ProjectSimpleV2, len(projects))
	for i, project := range projects {
		simpleProjects[i] = ProjectSimpleV2{
			ID:                   project.ID,
			Name:                 project.Name,
			Description:          project.Description,
			Code:                 project.Code,
			TeacherID:            project.TeacherID,
			TeacherName:          project.Teacher.Name,
			Status:               project.Status,
			Type:                 project.Type,
			StartDate:            project.StartDate,
			EndDate:              project.EndDate,
			MaxMembers:           project.MaxMembers,
			TotalAssignments:     project.TotalAssignments,
			CompletedAssignments: project.CompletedAssignments,
			CreatedAt:            project.CreatedAt,
			UpdatedAt:            project.UpdatedAt,
		}
	}

	return simpleProjects, total, nil
}

// GetProjectByID 根据ID获取课题详情
func (s *ProjectServiceV2) GetProjectByID(projectID uint) (*models.Project, error) {
	var project models.Project
	err := s.db.Preload("Teacher").
		Preload("Members").
		Preload("Members.User").
		First(&project, projectID).Error

	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	return &project, nil
}

// UpdateProject 更新课题信息
func (s *ProjectServiceV2) UpdateProject(projectID uint, req *UpdateProjectRequestV2) (*models.Project, error) {
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
	if req.Type != "" {
		project.Type = req.Type
	}
	if req.StartDate != nil {
		project.StartDate = req.StartDate.Time
	}
	if req.EndDate != nil {
		project.EndDate = req.EndDate.Time
	}
	if req.MaxMembers != nil {
		project.MaxMembers = *req.MaxMembers
	}
	if req.Status != "" {
		project.Status = req.Status
	}
	if req.WikiEnabled != nil {
		project.WikiEnabled = *req.WikiEnabled
	}
	if req.IssuesEnabled != nil {
		project.IssuesEnabled = *req.IssuesEnabled
	}
	if req.MREnabled != nil {
		project.MREnabled = *req.MREnabled
	}

	// 保存更新
	if err := s.db.Save(&project).Error; err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return &project, nil
}

// DeleteProject 删除课题
func (s *ProjectServiceV2) DeleteProject(projectID uint) error {
	// 检查是否有作业
	var assignmentCount int64
	if err := s.db.Model(&models.Assignment{}).Where("project_id = ?", projectID).Count(&assignmentCount).Error; err != nil {
		return fmt.Errorf("failed to check assignments: %w", err)
	}

	if assignmentCount > 0 {
		return fmt.Errorf("cannot delete project with assignments")
	}

	// 删除项目成员
	if err := s.db.Where("project_id = ?", projectID).Delete(&models.ProjectMember{}).Error; err != nil {
		return fmt.Errorf("failed to delete project members: %w", err)
	}

	// 删除项目
	if err := s.db.Delete(&models.Project{}, projectID).Error; err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

// JoinProject 学生加入课题
func (s *ProjectServiceV2) JoinProject(studentID uint, code string) (*models.Project, error) {
	// 查找课题
	var project models.Project
	if err := s.db.Where("code = ?", code).First(&project).Error; err != nil {
		return nil, fmt.Errorf("project not found with code: %s", code)
	}

	// 检查是否已经是成员
	var existingMember models.ProjectMember
	err := s.db.Where("project_id = ? AND user_id = ?", project.ID, studentID).First(&existingMember).Error
	if err == nil {
		return nil, fmt.Errorf("already a member of this project")
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}

	// 检查成员数量限制
	var memberCount int64
	if err := s.db.Model(&models.ProjectMember{}).Where("project_id = ? AND is_active = true", project.ID).Count(&memberCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count members: %w", err)
	}

	if int(memberCount) >= project.MaxMembers {
		return nil, fmt.Errorf("project is full")
	}

	// 添加成员
	member := &models.ProjectMember{
		ProjectID: project.ID,
		UserID:    studentID,
		Role:      "student",
		IsActive:  true,
	}

	if err := s.db.Create(member).Error; err != nil {
		return nil, fmt.Errorf("failed to join project: %w", err)
	}

	return &project, nil
}

// generateProjectCode 生成课题代码
func (s *ProjectServiceV2) generateProjectCode() string {
	// 简单的代码生成逻辑：年份 + 随机数
	return fmt.Sprintf("%d%04d", time.Now().Year(), time.Now().Unix()%10000)
}
