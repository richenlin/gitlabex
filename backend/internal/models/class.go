package models

import (
	"time"
)

// Class 班级模型 - 集成GitLab Group功能
type Class struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"not null" json:"name"`
	Description string `json:"description"`
	Code        string `gorm:"unique;not null" json:"code"` // 班级代码，用于学生加入
	TeacherID   uint   `gorm:"not null" json:"teacher_id"`  // 创建班级的老师ID
	Teacher     User   `gorm:"foreignKey:TeacherID" json:"teacher,omitempty"`
	Active      bool   `gorm:"default:true" json:"is_active"`

	// GitLab集成字段
	GitLabGroupID   *int       `gorm:"column:gitlab_group_id;unique" json:"gitlab_group_id,omitempty"` // 对应的GitLab Group ID
	GitLabGroupPath string     `gorm:"column:gitlab_group_path" json:"gitlab_group_path,omitempty"`    // GitLab Group路径
	GitLabGroupURL  string     `gorm:"column:gitlab_group_url" json:"gitlab_group_url,omitempty"`      // GitLab Group访问URL
	LastSyncAt      *time.Time `gorm:"column:last_sync_at" json:"last_sync_at,omitempty"`              // 最后同步时间
	SyncStatus      string     `gorm:"column:sync_status;default:'pending'" json:"sync_status"`        // 同步状态：pending, synced, failed
	SyncError       string     `gorm:"column:sync_error;type:text" json:"sync_error,omitempty"`        // 同步错误信息

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联关系
	Members  []ClassMember `gorm:"foreignKey:ClassID" json:"members,omitempty"`
	Students []User        `gorm:"many2many:class_members;" json:"students,omitempty"`
}

// TableName 指定表名
func (Class) TableName() string {
	return "classes"
}

// IsGitLabSynced 检查是否已同步到GitLab
func (c *Class) IsGitLabSynced() bool {
	return c.GitLabGroupID != nil && c.SyncStatus == "synced"
}

// GetGitLabGroupID 获取GitLab Group ID
func (c *Class) GetGitLabGroupID() int {
	if c.GitLabGroupID == nil {
		return 0
	}
	return *c.GitLabGroupID
}

// SetGitLabGroup 设置GitLab Group信息
func (c *Class) SetGitLabGroup(groupID int, groupPath, groupURL string) {
	c.GitLabGroupID = &groupID
	c.GitLabGroupPath = groupPath
	c.GitLabGroupURL = groupURL
	c.SyncStatus = "synced"
	c.SyncError = ""
	now := time.Now()
	c.LastSyncAt = &now
}

// SetSyncError 设置同步错误
func (c *Class) SetSyncError(err error) {
	c.SyncStatus = "failed"
	c.SyncError = err.Error()
	now := time.Now()
	c.LastSyncAt = &now
}

// ClassMember 班级成员关系 - 扩展GitLab权限映射
type ClassMember struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ClassID   uint      `gorm:"not null" json:"class_id"`
	StudentID uint      `gorm:"not null" json:"student_id"`
	JoinedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"joined_at"`
	Status    string    `gorm:"default:'active'" json:"status"` // active, inactive

	// GitLab集成字段
	GitLabRole       string     `gorm:"column:gitlab_role;default:'reporter'" json:"gitlab_role"`              // GitLab角色：guest, reporter, developer, maintainer, owner
	GitLabMemberID   *int       `gorm:"column:gitlab_member_id" json:"gitlab_member_id,omitempty"`             // GitLab成员ID
	GitLabSyncStatus string     `gorm:"column:gitlab_sync_status;default:'pending'" json:"gitlab_sync_status"` // 同步状态
	GitLabLastSyncAt *time.Time `gorm:"column:gitlab_last_sync_at" json:"gitlab_last_sync_at,omitempty"`       // GitLab最后同步时间
	GitLabSyncError  string     `gorm:"column:gitlab_sync_error;type:text" json:"gitlab_sync_error,omitempty"` // GitLab同步错误

	// 关联关系
	Class   Class `gorm:"foreignKey:ClassID" json:"class,omitempty"`
	Student User  `gorm:"foreignKey:StudentID" json:"student,omitempty"`
}

// TableName 指定表名
func (ClassMember) TableName() string {
	return "class_members"
}

// IsGitLabSynced 检查是否已同步到GitLab
func (cm *ClassMember) IsGitLabSynced() bool {
	return cm.GitLabMemberID != nil && cm.GitLabSyncStatus == "synced"
}

// SetGitLabMember 设置GitLab成员信息
func (cm *ClassMember) SetGitLabMember(memberID int, role string) {
	cm.GitLabMemberID = &memberID
	cm.GitLabRole = role
	cm.GitLabSyncStatus = "synced"
	cm.GitLabSyncError = ""
	now := time.Now()
	cm.GitLabLastSyncAt = &now
}

// SetGitLabSyncError 设置GitLab同步错误
func (cm *ClassMember) SetGitLabSyncError(err error) {
	cm.GitLabSyncStatus = "failed"
	cm.GitLabSyncError = err.Error()
	now := time.Now()
	cm.GitLabLastSyncAt = &now
}

// ClassStats 班级统计信息 - 扩展GitLab统计
type ClassStats struct {
	StudentCount       int `json:"student_count"`
	ProjectCount       int `json:"project_count"`
	ActiveProjects     int `json:"active_projects"`
	GitLabGroupID      int `json:"gitlab_group_id,omitempty"`
	GitLabProjectCount int `json:"gitlab_project_count,omitempty"`
	GitLabIssueCount   int `json:"gitlab_issue_count,omitempty"`
	GitLabMRCount      int `json:"gitlab_mr_count,omitempty"`
}
