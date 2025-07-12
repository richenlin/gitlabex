package services

import (
	"fmt"
	"time"

	"github.com/xanzy/go-gitlab"
	"gorm.io/gorm"
)

// EducationStats 教育统计数据
type EducationStats struct {
	ClassesCount            int `json:"classes_count"`
	ActiveProjectsCount     int `json:"active_projects_count"`
	PendingAssignmentsCount int `json:"pending_assignments_count"`
	DocumentsCount          int `json:"documents_count"`
}

// EducationServiceSimplified 简化版教育服务 - 专注于核心功能
type EducationServiceSimplified struct {
	gitlab *GitLabService
	db     *gorm.DB
}

// NewEducationServiceSimplified 创建简化版教育服务
func NewEducationServiceSimplified(gitlabService *GitLabService, db *gorm.DB) *EducationServiceSimplified {
	return &EducationServiceSimplified{
		gitlab: gitlabService,
		db:     db,
	}
}

// 辅助方法：将字符串切片转换为 GitLab 标签选项
func (s *EducationServiceSimplified) toLabels(labels []string) *gitlab.LabelOptions {
	labelOptions := gitlab.LabelOptions(labels)
	return &labelOptions
}

// CreateSimpleProject 创建简单的课题项目
func (s *EducationServiceSimplified) CreateSimpleProject(groupID int, title, description string) (*gitlab.Project, error) {
	// 创建项目
	project, _, err := s.gitlab.client.Projects.CreateProject(&gitlab.CreateProjectOptions{
		Name:        gitlab.String(title),
		NamespaceID: gitlab.Int(groupID),
		Description: gitlab.String(description),
		Visibility:  gitlab.Visibility(gitlab.PrivateVisibility),
		// 启用必要功能
		IssuesEnabled:        gitlab.Bool(true),
		MergeRequestsEnabled: gitlab.Bool(true),
		WikiEnabled:          gitlab.Bool(true),
	})

	return project, err
}

// CreateSimpleAssignment 创建简单的作业Issue
func (s *EducationServiceSimplified) CreateSimpleAssignment(projectID int, title, description string) (*gitlab.Issue, error) {
	// 创建作业Issue
	issue, _, err := s.gitlab.client.Issues.CreateIssue(projectID, &gitlab.CreateIssueOptions{
		Title:       gitlab.String(title),
		Description: gitlab.String(description),
		Labels:      s.toLabels([]string{"作业", "assignment"}),
	})

	return issue, err
}

// CreateSimpleAnnouncement 创建简单的公告Issue
func (s *EducationServiceSimplified) CreateSimpleAnnouncement(projectID int, title, content string) (*gitlab.Issue, error) {
	// 创建公告Issue
	issue, _, err := s.gitlab.client.Issues.CreateIssue(projectID, &gitlab.CreateIssueOptions{
		Title:       gitlab.String(title),
		Description: gitlab.String(content),
		Labels:      s.toLabels([]string{"公告", "announcement"}),
	})

	return issue, err
}

// SubmitSimpleAssignment 学生提交作业（简化版）
func (s *EducationServiceSimplified) SubmitSimpleAssignment(projectID int, issueID int, studentID int, branchName string) (*gitlab.MergeRequest, error) {
	// 创建作业提交MR
	mr, _, err := s.gitlab.client.MergeRequests.CreateMergeRequest(projectID, &gitlab.CreateMergeRequestOptions{
		Title:        gitlab.String(fmt.Sprintf("作业提交 - Issue #%d", issueID)),
		Description:  gitlab.String(fmt.Sprintf("关联作业: #%d\n\n提交人: @%d", issueID, studentID)),
		SourceBranch: gitlab.String(branchName),
		TargetBranch: gitlab.String("main"),
		AssigneeIDs:  &[]int{studentID},
		Labels:       s.toLabels([]string{"作业提交", "assignment-submission"}),
	})

	if err != nil {
		return nil, err
	}

	// 自动关联到作业Issue
	_, _, err = s.gitlab.client.Notes.CreateIssueNote(projectID, issueID, &gitlab.CreateIssueNoteOptions{
		Body: gitlab.String(fmt.Sprintf("学生提交作业: !%d", mr.IID)),
	})

	return mr, err
}

// GradeSimpleAssignment 教师批改作业（简化版）
func (s *EducationServiceSimplified) GradeSimpleAssignment(projectID int, mrID int, grade float64, feedback string) error {
	// 添加批改评论
	_, _, err := s.gitlab.client.Notes.CreateMergeRequestNote(projectID, mrID, &gitlab.CreateMergeRequestNoteOptions{
		Body: gitlab.String(fmt.Sprintf("## 作业批改\n\n**成绩**: %.1f分\n\n**反馈**: %s", grade, feedback)),
	})
	if err != nil {
		return err
	}

	// 添加成绩标签
	gradeLabel := fmt.Sprintf("成绩:%.1f", grade)
	_, _, err = s.gitlab.client.MergeRequests.UpdateMergeRequest(projectID, mrID, &gitlab.UpdateMergeRequestOptions{
		Labels: s.toLabels([]string{"作业提交", "assignment-submission", gradeLabel}),
	})

	return err
}

// GetSimpleAssignments 获取简单的作业列表
func (s *EducationServiceSimplified) GetSimpleAssignments(projectID int) ([]*gitlab.Issue, error) {
	// 获取作业Issues
	issues, _, err := s.gitlab.client.Issues.ListProjectIssues(projectID, &gitlab.ListProjectIssuesOptions{
		Labels: s.toLabels([]string{"作业", "assignment"}),
		State:  gitlab.String("opened"),
	})

	return issues, err
}

// GetSimpleSubmissions 获取简单的提交列表
func (s *EducationServiceSimplified) GetSimpleSubmissions(projectID int) ([]*gitlab.MergeRequest, error) {
	// 获取作业提交MR
	mrs, _, err := s.gitlab.client.MergeRequests.ListProjectMergeRequests(projectID, &gitlab.ListProjectMergeRequestsOptions{
		Labels: s.toLabels([]string{"作业提交", "assignment-submission"}),
		State:  gitlab.String("opened"),
	})

	return mrs, err
}

// GetSimpleAnnouncements 获取简单的公告列表
func (s *EducationServiceSimplified) GetSimpleAnnouncements(projectID int) ([]*gitlab.Issue, error) {
	// 获取公告Issues
	issues, _, err := s.gitlab.client.Issues.ListProjectIssues(projectID, &gitlab.ListProjectIssuesOptions{
		Labels: s.toLabels([]string{"公告", "announcement"}),
		State:  gitlab.String("opened"),
	})

	return issues, err
}

// GetSimpleEducationStats 获取简单的教育统计数据
func (s *EducationServiceSimplified) GetSimpleEducationStats(userID int) (*EducationStats, error) {
	// 获取用户所属的Groups
	groups, _, err := s.gitlab.client.Groups.ListGroups(&gitlab.ListGroupsOptions{
		MinAccessLevel: gitlab.AccessLevel(gitlab.ReporterPermissions),
	})
	if err != nil {
		return nil, err
	}

	// 统计数据
	stats := &EducationStats{
		ClassesCount:            len(groups),
		ActiveProjectsCount:     0,
		PendingAssignmentsCount: 0,
		DocumentsCount:          0,
	}

	// 计算项目和作业数量
	for _, group := range groups {
		projects, _, err := s.gitlab.client.Groups.ListGroupProjects(group.ID, &gitlab.ListGroupProjectsOptions{})
		if err != nil {
			continue
		}

		stats.ActiveProjectsCount += len(projects)

		// 计算作业数量
		for _, project := range projects {
			issues, _, err := s.gitlab.client.Issues.ListProjectIssues(project.ID, &gitlab.ListProjectIssuesOptions{
				Labels: s.toLabels([]string{"作业", "assignment"}),
				State:  gitlab.String("opened"),
			})
			if err != nil {
				continue
			}
			stats.PendingAssignmentsCount += len(issues)
		}
	}

	return stats, nil
}

// TestEducationWorkflow 测试教育工作流
func (s *EducationServiceSimplified) TestEducationWorkflow(groupID int) (*WorkflowTestResult, error) {
	result := &WorkflowTestResult{
		Steps: []WorkflowStep{},
	}

	// 步骤1: 创建项目
	step1 := WorkflowStep{
		Name:      "创建项目",
		Status:    "pending",
		StartTime: time.Now(),
	}

	project, err := s.CreateSimpleProject(groupID, "测试项目", "这是一个测试项目")
	if err != nil {
		step1.Status = "failed"
		step1.Error = err.Error()
		step1.EndTime = time.Now()
		result.Steps = append(result.Steps, step1)
		return result, err
	}

	step1.Status = "success"
	step1.EndTime = time.Now()
	result.Steps = append(result.Steps, step1)

	// 步骤2: 创建作业
	step2 := WorkflowStep{
		Name:      "创建作业",
		Status:    "pending",
		StartTime: time.Now(),
	}

	assignment, err := s.CreateSimpleAssignment(project.ID, "测试作业", "这是一个测试作业")
	if err != nil {
		step2.Status = "failed"
		step2.Error = err.Error()
		step2.EndTime = time.Now()
		result.Steps = append(result.Steps, step2)
		return result, err
	}

	step2.Status = "success"
	step2.EndTime = time.Now()
	result.Steps = append(result.Steps, step2)

	// 步骤3: 创建公告
	step3 := WorkflowStep{
		Name:      "创建公告",
		Status:    "pending",
		StartTime: time.Now(),
	}

	_, err = s.CreateSimpleAnnouncement(project.ID, "测试公告", "这是一个测试公告")
	if err != nil {
		step3.Status = "failed"
		step3.Error = err.Error()
		step3.EndTime = time.Now()
		result.Steps = append(result.Steps, step3)
		return result, err
	}

	step3.Status = "success"
	step3.EndTime = time.Now()
	result.Steps = append(result.Steps, step3)

	// 步骤4: 模拟学生提交作业
	step4 := WorkflowStep{
		Name:      "学生提交作业",
		Status:    "pending",
		StartTime: time.Now(),
	}

	submission, err := s.SubmitSimpleAssignment(project.ID, assignment.IID, 1, "feature/student-submission")
	if err != nil {
		step4.Status = "failed"
		step4.Error = err.Error()
		step4.EndTime = time.Now()
		result.Steps = append(result.Steps, step4)
		return result, err
	}

	step4.Status = "success"
	step4.EndTime = time.Now()
	result.Steps = append(result.Steps, step4)

	// 步骤5: 教师批改作业
	step5 := WorkflowStep{
		Name:      "教师批改作业",
		Status:    "pending",
		StartTime: time.Now(),
	}

	err = s.GradeSimpleAssignment(project.ID, submission.IID, 85.5, "作业完成得不错，但还有改进空间")
	if err != nil {
		step5.Status = "failed"
		step5.Error = err.Error()
		step5.EndTime = time.Now()
		result.Steps = append(result.Steps, step5)
		return result, err
	}

	step5.Status = "success"
	step5.EndTime = time.Now()
	result.Steps = append(result.Steps, step5)

	result.Success = true
	return result, nil
}

// WorkflowTestResult 工作流测试结果
type WorkflowTestResult struct {
	Success bool           `json:"success"`
	Steps   []WorkflowStep `json:"steps"`
}

// WorkflowStep 工作流步骤
type WorkflowStep struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"` // pending, success, failed
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Error     string    `json:"error,omitempty"`
}
