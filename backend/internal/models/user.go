package models

import "time"

// UserRole 用户角色枚举
type UserRole string

const (
	RoleAdmin     UserRole = "admin"
	RoleTeacher   UserRole = "teacher"
	RoleAssistant UserRole = "assistant"
	RoleStudent   UserRole = "student"
	RoleGuest     UserRole = "guest"
)

// EducationRole 教育角色枚举 - 基于GitLab权限映射
type EducationRole int

const (
	EduRoleGuest     EducationRole = 10 // GitLab Guest -> 访客
	EduRoleStudent   EducationRole = 20 // GitLab Reporter -> 学生
	EduRoleAssistant EducationRole = 30 // GitLab Developer -> 研究员
	EduRoleTeacher   EducationRole = 40 // GitLab Maintainer -> 教师
	EduRoleAdmin     EducationRole = 50 // GitLab Owner -> 管理员
)

// ProjectRole 项目角色枚举
type ProjectRole string

const (
	ProjectRoleOwner      ProjectRole = "owner"
	ProjectRoleMaintainer ProjectRole = "maintainer"
	ProjectRoleDeveloper  ProjectRole = "developer"
	ProjectRoleReporter   ProjectRole = "reporter"
	ProjectRoleGuest      ProjectRole = "guest"
)

// User 用户模型
type User struct {
	BaseModel
	GitLabID     int64         `gorm:"uniqueIndex;not null" json:"gitlab_id"`
	Username     string        `gorm:"uniqueIndex;not null;size:50" json:"username"`
	Email        string        `gorm:"uniqueIndex;not null;size:100" json:"email"`
	Name         string        `gorm:"not null;size:100" json:"name"`
	AvatarURL    string        `gorm:"size:500" json:"avatar_url"`
	Role         UserRole      `gorm:"not null;default:student" json:"role"`
	EduRole      EducationRole `gorm:"not null;default:20" json:"edu_role"`
	IsActive     bool          `gorm:"default:true" json:"is_active"`
	LastLoginAt  *time.Time    `json:"last_login_at,omitempty"`
	AccessToken  string        `gorm:"size:500" json:"-"`
	RefreshToken string        `gorm:"size:500" json:"-"`
	TokenExpiry  *time.Time    `json:"token_expiry,omitempty"`

	// 关联关系
	CreatedProjects       []ResearchProject `gorm:"foreignKey:CreatorID" json:"created_projects,omitempty"`
	ProjectMembers        []ProjectMember   `gorm:"foreignKey:UserID" json:"project_members,omitempty"`
	Topics                []Topic           `gorm:"foreignKey:AuthorID" json:"topics,omitempty"`
	Comments              []Comment         `gorm:"foreignKey:AuthorID" json:"comments,omitempty"`
	Documents             []Document        `gorm:"foreignKey:UploaderID" json:"documents,omitempty"`
	Submissions           []Submission      `gorm:"foreignKey:StudentID" json:"submissions,omitempty"`
	GradedSubmissions     []Submission      `gorm:"foreignKey:GradedBy" json:"graded_submissions,omitempty"`
	SentNotifications     []Notification    `gorm:"foreignKey:SenderID" json:"sent_notifications,omitempty"`
	ReceivedNotifications []Notification    `gorm:"foreignKey:RecipientID" json:"received_notifications,omitempty"`
	CreatedAnnouncements  []Announcement    `gorm:"foreignKey:AuthorID" json:"created_announcements,omitempty"`
}
