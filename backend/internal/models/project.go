package models

import (
	"time"
)

// Project 课题模型
type Project struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `json:"description"`
	Code        string    `gorm:"unique;not null" json:"code"`    // 课题代码，用于学生加入
	TeacherID   uint      `gorm:"not null" json:"teacher_id"`     // 创建课题的老师ID
	ClassID     uint      `json:"class_id"`                       // 所属班级ID（可选）
	Status      string    `gorm:"default:'active'" json:"status"` // active, completed, archived
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// GitLab 相关字段
	GitLabProjectID int    `json:"gitlab_project_id"`                    // GitLab项目ID
	GitLabURL       string `json:"gitlab_url"`                           // GitLab项目URL
	RepositoryURL   string `json:"repository_url"`                       // 仓库URL
	DefaultBranch   string `gorm:"default:'main'" json:"default_branch"` // 默认分支
	ReadmeContent   string `json:"readme_content"`                       // README内容（课题介绍）
	WikiEnabled     bool   `gorm:"default:true" json:"wiki_enabled"`     // 是否启用Wiki
	IssuesEnabled   bool   `gorm:"default:true" json:"issues_enabled"`   // 是否启用Issues
	MREnabled       bool   `gorm:"default:true" json:"mr_enabled"`       // 是否启用合并请求

	// 关联关系
	Teacher     User            `gorm:"foreignKey:TeacherID" json:"teacher,omitempty"`
	Class       Class           `gorm:"foreignKey:ClassID" json:"class,omitempty"`
	Members     []ProjectMember `gorm:"foreignKey:ProjectID" json:"members,omitempty"`
	Students    []User          `gorm:"many2many:project_members;" json:"students,omitempty"`
	Assignments []Assignment    `gorm:"foreignKey:ProjectID" json:"assignments,omitempty"`
}

// TableName 指定表名
func (Project) TableName() string {
	return "projects"
}

// ProjectMember 课题成员关系
type ProjectMember struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID uint      `gorm:"not null" json:"project_id"`
	StudentID uint      `gorm:"not null" json:"student_id"`
	Role      string    `gorm:"default:'member'" json:"role"` // member, leader
	JoinedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"joined_at"`
	Status    string    `gorm:"default:'active'" json:"status"` // active, inactive

	// GitLab 相关字段
	GitLabAccessLevel int        `json:"gitlab_access_level"` // GitLab访问级别
	PersonalBranch    string     `json:"personal_branch"`     // 个人分支名
	PersonalBranchURL string     `json:"personal_branch_url"` // 个人分支URL
	LastCommitHash    string     `json:"last_commit_hash"`    // 最后提交的哈希
	LastCommitMessage string     `json:"last_commit_message"` // 最后提交的消息
	LastCommitTime    *time.Time `json:"last_commit_time"`    // 最后提交时间

	// 关联关系
	Project Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Student User    `gorm:"foreignKey:StudentID" json:"student,omitempty"`
}

// TableName 指定表名
func (ProjectMember) TableName() string {
	return "project_members"
}

// ProjectStats 课题统计信息
type ProjectStats struct {
	MemberCount          int `json:"member_count"`
	AssignmentCount      int `json:"assignment_count"`
	CompletedAssignments int `json:"completed_assignments"`
	PendingAssignments   int `json:"pending_assignments"`

	// GitLab 相关统计
	TotalCommits       int `json:"total_commits"`
	TotalIssues        int `json:"total_issues"`
	OpenIssues         int `json:"open_issues"`
	TotalMergeRequests int `json:"total_merge_requests"`
	OpenMergeRequests  int `json:"open_merge_requests"`
	WikiPages          int `json:"wiki_pages"`
	ActiveBranches     int `json:"active_branches"`
}
