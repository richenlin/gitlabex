package models

import (
	"time"
)

// Notification 通知模型
type Notification struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     uint       `gorm:"not null" json:"user_id"`   // 接收通知的用户ID
	Title      string     `gorm:"not null" json:"title"`     // 通知标题
	Content    string     `json:"content"`                   // 通知内容
	Type       string     `gorm:"not null" json:"type"`      // 通知类型: assignment_submitted, assignment_reviewed, project_joined, etc.
	TargetType string     `json:"target_type"`               // 目标类型: assignment, project, class
	TargetID   uint       `json:"target_id"`                 // 目标ID
	Read       bool       `gorm:"default:false" json:"read"` // 是否已读
	ReadAt     *time.Time `json:"read_at"`                   // 读取时间
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	// 关联关系
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (Notification) TableName() string {
	return "notifications"
}

// NotificationTypes 通知类型常量
const (
	NotificationTypeAssignmentSubmitted = "assignment_submitted"
	NotificationTypeAssignmentReviewed  = "assignment_reviewed"
	NotificationTypeProjectJoined       = "project_joined"
	NotificationTypeClassJoined         = "class_joined"
	NotificationTypeAssignmentCreated   = "assignment_created"
	NotificationTypeProjectCreated      = "project_created"
)

// MarkAsRead 标记通知为已读
func (n *Notification) MarkAsRead() {
	n.Read = true
	now := time.Now()
	n.ReadAt = &now
}
