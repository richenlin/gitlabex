package models

import (
	"time"

	"gorm.io/gorm"
)

// ExternalUser 外部系统用户映射表
// 用于存储外部系统用户与GitLab用户的映射关系
type ExternalUser struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	GitLabUserID   int64          `gorm:"not null;index:idx_gitlab_user" json:"gitlab_user_id"`
	ExternalID     string         `gorm:"not null;size:255;index:idx_external_id" json:"external_id"`
	ExternalSource string         `gorm:"not null;size:100;index:idx_external_source" json:"external_source"`
	Username       string         `gorm:"not null;size:255;index:idx_username" json:"username"`
	Email          string         `gorm:"not null;size:255;index:idx_email" json:"email"`
	Name           string         `gorm:"size:255" json:"name"`
	Role           string         `gorm:"size:50" json:"role"`
	Department     string         `gorm:"size:255" json:"department"`
	StudentID      string         `gorm:"size:100" json:"student_id"`
	TeacherID      string         `gorm:"size:100" json:"teacher_id"`
	Phone          string         `gorm:"size:50" json:"phone"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	LastSyncAt     *time.Time     `json:"last_sync_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 设置表名
func (ExternalUser) TableName() string {
	return "external_users"
}

// ExternalUserMapping 外部用户映射关系
type ExternalUserMapping struct {
	GitLabUser   *GitLabUser   `json:"gitlab_user"`
	ExternalUser *ExternalUser `json:"external_user"`
}

// SyncStatus 同步状态
type SyncStatus string

const (
	SyncStatusPending SyncStatus = "pending"
	SyncStatusSuccess SyncStatus = "success"
	SyncStatusFailed  SyncStatus = "failed"
	SyncStatusSkipped SyncStatus = "skipped"
)

// ExternalUserSyncLog 外部用户同步日志
type ExternalUserSyncLog struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	ExternalUserID uint       `gorm:"not null;index" json:"external_user_id"`
	Operation      string     `gorm:"not null;size:50" json:"operation"` // create, update, delete
	Status         SyncStatus `gorm:"not null;size:20" json:"status"`
	RequestData    string     `gorm:"type:text" json:"request_data"`
	ResponseData   string     `gorm:"type:text" json:"response_data"`
	ErrorMessage   string     `gorm:"type:text" json:"error_message"`
	APIKey         string     `gorm:"size:100" json:"api_key"` // 用于追踪是哪个API密钥发起的操作
	ClientIP       string     `gorm:"size:45" json:"client_ip"`
	CreatedAt      time.Time  `json:"created_at"`

	// 关联
	ExternalUser ExternalUser `gorm:"foreignKey:ExternalUserID" json:"external_user,omitempty"`
}

// TableName 设置表名
func (ExternalUserSyncLog) TableName() string {
	return "external_user_sync_logs"
}

// ExternalSystemStats 外部系统统计信息
type ExternalSystemStats struct {
	Source      string     `json:"source"`
	TotalUsers  int64      `json:"total_users"`
	ActiveUsers int64      `json:"active_users"`
	LastSyncAt  *time.Time `json:"last_sync_at"`
}

// ConvertToGitLabRole 将外部系统角色转换为GitLab角色
func ConvertExternalRoleToGitLab(externalRole string) GitLabRole {
	switch externalRole {
	case "admin", "administrator", "系统管理员", "管理员":
		return GitLabOwner
	case "teacher", "instructor", "professor", "教师", "老师", "教授":
		return GitLabMaintainer
	case "researcher", "research_assistant", "研究员", "研究助理":
		return GitLabDeveloper
	case "student", "undergraduate", "graduate", "学生", "本科生", "研究生":
		return GitLabReporter
	case "guest", "visitor", "游客", "访客":
		return GitLabGuest
	default:
		return GitLabGuest // 默认为访客权限
	}
}
