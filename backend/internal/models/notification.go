package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// NotificationType 通知类型枚举
type NotificationType string

const (
	NotificationTypeAnnouncement   NotificationType = "announcement"    // 公告
	NotificationTypeProjectCreate  NotificationType = "project_create"  // 课题创建
	NotificationTypeProjectUpdate  NotificationType = "project_update"  // 课题更新
	NotificationTypeTopicCreate    NotificationType = "topic_create"    // 话题创建
	NotificationTypeTopicReply     NotificationType = "topic_reply"     // 话题回复
	NotificationTypeHomeworkCreate NotificationType = "homework_create" // 作业创建
	NotificationTypeHomeworkSubmit NotificationType = "homework_submit" // 作业提交
	NotificationTypeHomeworkGrade  NotificationType = "homework_grade"  // 作业评分
	NotificationTypeDocumentUpload NotificationType = "document_upload" // 文档上传
	NotificationTypeDocumentReview NotificationType = "document_review" // 文档审核
)

// Notification 通知模型
type Notification struct {
	BaseModel
	Type        NotificationType `gorm:"not null" json:"type"`
	Title       string           `gorm:"not null" json:"title"`
	Content     string           `gorm:"not null" json:"content"`
	RecipientID uuid.UUID        `gorm:"not null" json:"recipient_id"`
	SenderID    *uuid.UUID       `json:"sender_id,omitempty"`
	ProjectID   *uuid.UUID       `json:"project_id,omitempty"`
	TopicID     *uuid.UUID       `json:"topic_id,omitempty"`
	HomeworkID  *uuid.UUID       `json:"homework_id,omitempty"`
	DocumentID  *uuid.UUID       `json:"document_id,omitempty"`
	IsRead      bool             `gorm:"default:false" json:"is_read"`
	ReadAt      *time.Time       `json:"read_at,omitempty"`
	ActionURL   string           `json:"action_url"`

	// 关联关系
	Recipient User             `gorm:"foreignKey:RecipientID" json:"recipient,omitempty"`
	Sender    *User            `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	Project   *ResearchProject `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Topic     *Topic           `gorm:"foreignKey:TopicID" json:"topic,omitempty"`
	Homework  *Homework        `gorm:"foreignKey:HomeworkID" json:"homework,omitempty"`
	Document  *Document        `gorm:"foreignKey:DocumentID" json:"document,omitempty"`
}

// Announcement 公告模型
type Announcement struct {
	BaseModel
	Title       string         `gorm:"not null" json:"title"`
	Content     string         `gorm:"not null" json:"content"`
	AuthorID    uuid.UUID      `gorm:"not null" json:"author_id"`
	Priority    string         `gorm:"default:normal" json:"priority"` // low, normal, high, urgent
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	ValidFrom   time.Time      `json:"valid_from"`
	ValidTo     *time.Time     `json:"valid_to,omitempty"`
	TargetRoles pq.StringArray `gorm:"type:text[]" json:"target_roles"` // 目标角色

	// 关联关系
	Author User `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
}

// AssignmentTemplate 作业模板
type AssignmentTemplate struct {
	BaseModel
	Title       string         `gorm:"not null" json:"title"`
	Description string         `json:"description"`
	DueDate     *time.Time     `json:"due_date"`
	MaxGrade    int            `gorm:"default:100" json:"max_grade"`
	CreatorID   uuid.UUID      `gorm:"not null" json:"creator_id"`
	IsPublic    bool           `gorm:"default:false" json:"is_public"`
	Tags        pq.StringArray `gorm:"type:text[]" json:"tags"`

	// 关联关系
	Creator User `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
}
