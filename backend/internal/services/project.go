package services

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"gitlabex/internal/models"
	"gitlabex/internal/utils"

	"github.com/xanzy/go-gitlab"
	"gorm.io/gorm"
)

// ProjectService 课题管理服务
type ProjectService struct {
	db                *gorm.DB
	permissionService *PermissionService
	gitlabService     *GitLabService
}

// NewProjectService 创建课题管理服务
func NewProjectService(db *gorm.DB, permissionService *PermissionService, gitlabService *GitLabService) *ProjectService {
	return &ProjectService{
		db:                db,
		permissionService: permissionService,
		gitlabService:     gitlabService,
	}
}

// CreateProjectRequest 创建课题请求
type CreateProjectRequest struct {
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

// UpdateProjectRequest 更新课题请求
type UpdateProjectRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Status      string          `json:"status"`
	StartDate   *utils.DateOnly `json:"start_date"`
	EndDate     *utils.DateOnly `json:"end_date"`
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

	// 创建GitLab项目仓库
	var teacher models.User
	if err := s.db.First(&teacher, teacherID).Error; err != nil {
		return nil, fmt.Errorf("teacher not found: %w", err)
	}

	// 构建README内容
	readmeContent := fmt.Sprintf("# %s\n\n## 课题介绍\n\n%s\n\n## 使用说明\n\n1. 学生加入课题后，系统会自动为每个学生创建个人分支\n2. 学生在个人分支下的 `students/[学号]` 目录中提交作业\n3. 作业提交后，可通过合并请求进行代码审查\n4. 使用Issues进行讨论和问题跟踪\n5. 使用Wiki管理课题文档\n\n## 目录结构\n\n```\n├── README.md          # 课题说明\n├── assignments/       # 作业要求\n├── resources/         # 参考资料\n├── students/          # 学生作业目录\n│   ├── student1/     # 学生1的作业\n│   └── student2/     # 学生2的作业\n└── wiki/             # 课题文档\n```\n", req.Name, req.Description)

	// 初始化项目记录
	project := &models.Project{
		Name:          req.Name,
		Description:   req.Description,
		Code:          code,
		TeacherID:     teacherID,
		Status:        "active",
		StartDate:     req.StartDate.Time,
		EndDate:       req.EndDate.Time,
		ReadmeContent: readmeContent,
		WikiEnabled:   req.WikiEnabled,
		IssuesEnabled: req.IssuesEnabled,
		MREnabled:     req.MREnabled,
	}

	// 尝试创建GitLab项目（如果失败，仍继续创建本地项目）
	gitlabProject, err := s.gitlabService.CreateProjectRepository(req.Name, req.Description, readmeContent, teacher.GitLabID)
	if err != nil {
		// GitLab创建失败，记录警告但继续创建本地项目
		fmt.Printf("WARNING: Failed to create GitLab project: %v\n", err)
		fmt.Printf("INFO: Creating project without GitLab integration\n")
	} else {
		// GitLab创建成功，更新项目信息
		project.GitLabProjectID = gitlabProject.ID
		project.GitLabURL = gitlabProject.WebURL
		project.RepositoryURL = gitlabProject.HTTPURLToRepo
		project.DefaultBranch = gitlabProject.DefaultBranch
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
func (s *ProjectService) GetProjectsByTeacher(teacherID uint) ([]ProjectSimple, error) {
	var projects []models.Project
	err := s.db.Preload("Teacher", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, username, email, role")
	}).Select("id, name, description, code, teacher_id, status, type, start_date, end_date, max_members, total_assignments, completed_assignments, created_at, updated_at").
		Where("teacher_id = ?", teacherID).
		Order("created_at DESC").
		Find(&projects).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	// 转换为简化的结构
	simpleProjects := make([]ProjectSimple, len(projects))
	for i, project := range projects {
		simpleProjects[i] = ProjectSimple{
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

// GetProjectsByStudent 获取学生参加的课题列表
func (s *ProjectService) GetProjectsByStudent(studentID uint) ([]ProjectSimple, error) {
	var projects []models.Project
	err := s.db.Preload("Teacher", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, username, email, role")
	}).Select("projects.id, projects.name, projects.description, projects.code, projects.teacher_id, projects.status, projects.type, projects.start_date, projects.end_date, projects.max_members, projects.total_assignments, projects.completed_assignments, projects.created_at, projects.updated_at").
		Joins("JOIN project_members ON projects.id = project_members.project_id").
		Where("project_members.user_id = ? AND project_members.is_active = ?", studentID, true).
		Order("project_members.joined_at DESC").
		Find(&projects).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	// 转换为简化的结构
	simpleProjects := make([]ProjectSimple, len(projects))
	for i, project := range projects {
		simpleProjects[i] = ProjectSimple{
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

	fmt.Printf("DEBUG: Total projects count: %d\n", total)

	// 分页查询
	offset := (page - 1) * pageSize
	err := s.db.Preload("Teacher").
		Preload("Class").
		Preload("Members").
		Preload("Students").
		Preload("Assignments").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&projects).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get projects: %w", err)
	}

	fmt.Printf("DEBUG: Found %d projects\n", len(projects))

	return projects, total, nil
}

// GetAllProjectsSimple 获取所有课题的简化信息（管理员权限）
func (s *ProjectService) GetAllProjectsSimple(page, pageSize int) ([]ProjectSimple, int64, error) {
	var projects []models.Project
	var total int64

	// 计算总数
	if err := s.db.Model(&models.Project{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count projects: %w", err)
	}

	fmt.Printf("DEBUG: Total projects count: %d\n", total)

	// 分页查询，只预加载必要的关联数据
	offset := (page - 1) * pageSize
	err := s.db.Preload("Teacher", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, username, email, role")
	}).Select("id, name, description, code, teacher_id, status, type, start_date, end_date, max_members, total_assignments, completed_assignments, created_at, updated_at").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&projects).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get projects: %w", err)
	}

	fmt.Printf("DEBUG: Found %d projects\n", len(projects))

	// 转换为简化的结构
	simpleProjects := make([]ProjectSimple, len(projects))
	for i, project := range projects {
		simpleProjects[i] = ProjectSimple{
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

// ProjectSimple 简化的项目结构
type ProjectSimple struct {
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
		project.StartDate = req.StartDate.Time
	}
	if req.EndDate != nil {
		project.EndDate = req.EndDate.Time
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
	// 检查是否存在作业
	var assignmentCount int64
	if err := s.db.Model(&models.Assignment{}).Where("project_id = ?", projectID).Count(&assignmentCount).Error; err != nil {
		return fmt.Errorf("failed to check assignments: %w", err)
	}

	if assignmentCount > 0 {
		return fmt.Errorf("cannot delete project with existing assignments")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除课题成员关系
		if err := tx.Where("project_id = ?", projectID).Delete(&models.ProjectMember{}).Error; err != nil {
			return fmt.Errorf("failed to delete project members: %w", err)
		}

		// 删除课题文件
		if err := tx.Where("project_id = ?", projectID).Delete(&models.ProjectFile{}).Error; err != nil {
			return fmt.Errorf("failed to delete project files: %w", err)
		}

		// 删除编辑会话
		if err := tx.Where("project_id = ?", projectID).Delete(&models.CodeEditSession{}).Error; err != nil {
			return fmt.Errorf("failed to delete code edit sessions: %w", err)
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
		return nil, fmt.Errorf("project not found or inactive: %w", err)
	}

	// 检查学生是否已经在课题中
	var existingMember models.ProjectMember
	if err := s.db.Where("project_id = ? AND student_id = ?", project.ID, studentID).First(&existingMember).Error; err == nil {
		return nil, fmt.Errorf("already joined this project")
	}

	// 添加学生到课题
	if err := s.AddStudentToProject(project.ID, studentID, "member"); err != nil {
		return nil, fmt.Errorf("failed to add student to project: %w", err)
	}

	return &project, nil
}

// AddStudentToProject 添加学生到课题
func (s *ProjectService) AddStudentToProject(projectID uint, studentID uint, role string) error {
	// 检查学生是否已经在课题中
	var existingMember models.ProjectMember
	if err := s.db.Where("project_id = ? AND student_id = ?", projectID, studentID).First(&existingMember).Error; err == nil {
		return fmt.Errorf("student already in project")
	}

	// 获取课题信息
	var project models.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	// 获取学生信息
	var student models.User
	if err := s.db.First(&student, studentID).Error; err != nil {
		return fmt.Errorf("student not found: %w", err)
	}

	// 在GitLab中添加学生到项目
	accessLevel := gitlab.DeveloperPermissions // 学生默认为Developer权限
	if role == "leader" {
		accessLevel = gitlab.MaintainerPermissions
	}

	if err := s.gitlabService.AddStudentToProject(project.GitLabProjectID, student.GitLabID, accessLevel); err != nil {
		return fmt.Errorf("failed to add student to GitLab project: %w", err)
	}

	// 为学生创建个人分支
	branchName := fmt.Sprintf("student-%s", student.Username)
	_, err := s.gitlabService.CreateStudentBranch(project.GitLabProjectID, student.GitLabID, branchName)
	if err != nil {
		return fmt.Errorf("failed to create student branch: %w", err)
	}

	// 创建项目成员记录
	member := &models.ProjectMember{
		ProjectID: projectID,
		UserID:    studentID,
		Role:      role,
		JoinedAt:  time.Now(),
		IsActive:  true,
		// GitLab相关字段
		GitLabAccessLevel: int(accessLevel),
		PersonalBranch:    branchName,
		PersonalBranchURL: fmt.Sprintf("%s/-/tree/%s", project.GitLabURL, branchName),
	}

	if err := s.db.Create(member).Error; err != nil {
		return fmt.Errorf("failed to create project member: %w", err)
	}

	return nil
}

// RemoveStudentFromProject 从课题移除学生
func (s *ProjectService) RemoveStudentFromProject(projectID uint, studentID uint) error {
	var member models.ProjectMember
	if err := s.db.Where("project_id = ? AND student_id = ?", projectID, studentID).First(&member).Error; err != nil {
		return fmt.Errorf("student not found in project: %w", err)
	}

	member.IsActive = false
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
func (s *ProjectService) GetProjectMembers(projectID uint) ([]models.ProjectMember, error) {
	var members []models.ProjectMember
	err := s.db.Preload("User").
		Where("project_id = ? AND is_active = ?", projectID, true).
		Order("joined_at ASC").
		Find(&members).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get project members: %w", err)
	}

	return members, nil
}

// GetProjectStats 获取课题统计信息
func (s *ProjectService) GetProjectStats(projectID uint) (*models.ProjectStats, error) {
	stats := &models.ProjectStats{}

	// 成员数量
	var memberCount int64
	if err := s.db.Model(&models.ProjectMember{}).
		Where("project_id = ? AND is_active = ?", projectID, true).
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

	// 已完成作业数量（基于评分的作业提交）
	var completedCount int64
	if err := s.db.Table("assignment_submissions").
		Joins("JOIN assignments ON assignments.id = assignment_submissions.assignment_id").
		Where("assignments.project_id = ? AND assignment_submissions.status = ?", projectID, "graded").
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

// GetProjectWithGitLabStats 获取课题及其GitLab统计信息
func (s *ProjectService) GetProjectWithGitLabStats(projectID uint) (*models.Project, *ProjectStatistics, error) {
	var project models.Project
	err := s.db.Preload("Teacher").
		Preload("Class").
		Preload("Members").
		Preload("Students").
		Preload("Assignments").
		First(&project, projectID).Error

	if err != nil {
		return nil, nil, fmt.Errorf("project not found: %w", err)
	}

	// 获取GitLab统计信息
	var gitlabStats *ProjectStatistics
	if project.GitLabProjectID > 0 {
		gitlabStats, err = s.gitlabService.GetProjectStatistics(project.GitLabProjectID)
		if err != nil {
			// 如果获取统计失败，返回空统计而不是错误
			gitlabStats = &ProjectStatistics{
				ProjectID: project.GitLabProjectID,
			}
		}
	}

	return &project, gitlabStats, nil
}

// SubmitAssignmentToGitLab 提交作业到GitLab
func (s *ProjectService) SubmitAssignmentToGitLab(projectID uint, studentID uint, assignmentID uint, files map[string]string) (*models.AssignmentSubmission, error) {
	// 获取项目信息
	var project models.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// 获取学生信息
	var student models.User
	if err := s.db.First(&student, studentID).Error; err != nil {
		return nil, fmt.Errorf("student not found: %w", err)
	}

	// 获取学生的项目成员信息
	var member models.ProjectMember
	if err := s.db.Where("project_id = ? AND student_id = ?", projectID, studentID).First(&member).Error; err != nil {
		return nil, fmt.Errorf("student not in project: %w", err)
	}

	// 构建提交消息
	commitMessage := fmt.Sprintf("Submit assignment %d by %s", assignmentID, student.Username)

	// 提交文件到GitLab
	commitHash, err := s.gitlabService.SubmitAssignment(project.GitLabProjectID, member.PersonalBranch, files, commitMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to submit to GitLab: %w", err)
	}

	// 创建本地提交记录
	submission := &models.AssignmentSubmission{
		AssignmentID: assignmentID,
		StudentID:    studentID,
		Status:       "submitted",
		SubmittedAt:  time.Now(),
		// GitLab相关字段
		CommitHash:    commitHash,
		CommitMessage: commitMessage,
		CommitURL:     fmt.Sprintf("%s/-/commit/%s", project.GitLabURL, commitHash),
		BranchName:    member.PersonalBranch,
		BranchURL:     member.PersonalBranchURL,
	}

	// 构建文件列表
	var fileList []string
	for filePath := range files {
		fileList = append(fileList, filePath)
	}
	submission.FilesSubmitted = fileList

	if err := s.db.Create(submission).Error; err != nil {
		return nil, fmt.Errorf("failed to create submission record: %w", err)
	}

	// 更新成员的最后提交信息
	member.LastCommitHash = commitHash
	member.LastCommitMessage = commitMessage
	now := time.Now()
	member.LastCommitTime = &now
	if err := s.db.Save(&member).Error; err != nil {
		return nil, fmt.Errorf("failed to update member commit info: %w", err)
	}

	return submission, nil
}

// GetProjectGitLabInfo 获取课题GitLab信息
func (s *ProjectService) GetProjectGitLabInfo(projectID uint) (*GitLabProjectInfo, error) {
	var project models.Project
	if err := s.db.First(&project, projectID).Error; err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	if project.GitLabProjectID == 0 {
		return nil, fmt.Errorf("project not linked to GitLab")
	}

	// 获取GitLab统计信息
	stats, err := s.gitlabService.GetProjectStatistics(project.GitLabProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitLab statistics: %w", err)
	}

	// 获取讨论
	discussions, err := s.gitlabService.GetDiscussions(project.GitLabProjectID)
	if err != nil {
		discussions = []*gitlab.Issue{} // 如果获取失败，返回空数组
	}

	return &GitLabProjectInfo{
		ProjectID:     project.GitLabProjectID,
		WebURL:        project.GitLabURL,
		RepositoryURL: project.RepositoryURL,
		DefaultBranch: project.DefaultBranch,
		Statistics:    stats,
		Discussions:   discussions,
	}, nil
}

// ProjectStatistics 已在gitlab.go中定义，这里移除重复定义

// GitLabProjectInfo GitLab项目信息
type GitLabProjectInfo struct {
	ProjectID     int                `json:"project_id"`
	WebURL        string             `json:"web_url"`
	RepositoryURL string             `json:"repository_url"`
	DefaultBranch string             `json:"default_branch"`
	Statistics    *ProjectStatistics `json:"statistics"`
	Discussions   []*gitlab.Issue    `json:"discussions"`
}

// GetProjectAssignments 获取课题作业列表
func (s *ProjectService) GetProjectAssignments(projectID uint) ([]models.Assignment, error) {
	var assignments []models.Assignment
	err := s.db.Where("project_id = ?", projectID).
		Order("created_at DESC").
		Find(&assignments).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get project assignments: %w", err)
	}

	return assignments, nil
}
