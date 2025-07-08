package models

import (
	"time"

	"gorm.io/gorm"
)

// TopicType 话题类型枚举
type TopicType int

const (
	TopicAnnouncement TopicType = iota // 公告
	TopicProject                       // 课题
	TopicAssignment                    // 作业
	TopicDiscussion                    // 讨论
)

// String 返回话题类型的字符串表示
func (t TopicType) String() string {
	switch t {
	case TopicAnnouncement:
		return "announcement"
	case TopicProject:
		return "project"
	case TopicAssignment:
		return "assignment"
	case TopicDiscussion:
		return "discussion"
	default:
		return "unknown"
	}
}

// TopicStatus 话题状态枚举
type TopicStatus string

const (
	StatusActive   TopicStatus = "active"
	StatusClosed   TopicStatus = "closed"
	StatusArchived TopicStatus = "archived"
	StatusDraft    TopicStatus = "draft"
)

// Topic 话题模型
type Topic struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Type       TopicType      `gorm:"not null;index" json:"type"`
	Title      string         `gorm:"not null;size:255" json:"title"`
	Content    string         `gorm:"type:text" json:"content"`
	Status     TopicStatus    `gorm:"size:20;default:active;index" json:"status"`
	Priority   int            `gorm:"default:0" json:"priority"`      // 优先级，数字越大优先级越高
	IsSticky   bool           `gorm:"default:false" json:"is_sticky"` // 是否置顶
	ViewCount  int            `gorm:"default:0" json:"view_count"`
	LikeCount  int            `gorm:"default:0" json:"like_count"`
	CreatedBy  uint           `gorm:"not null;index" json:"created_by"`
	AssignedTo string         `gorm:"type:json" json:"assigned_to"` // 分配给的用户ID列表，JSON格式
	DueDate    *time.Time     `json:"due_date"`
	ProjectID  *int           `gorm:"index" json:"project_id"` // 关联的GitLab项目ID
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	Creator     User         `gorm:"foreignKey:CreatedBy;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"creator,omitempty"`
	Comments    []Comment    `gorm:"foreignKey:TopicID" json:"comments,omitempty"`
	Attachments []Attachment `gorm:"foreignKey:TopicID" json:"attachments,omitempty"`
	Assignments []Assignment `gorm:"foreignKey:TopicID" json:"assignments,omitempty"`
	Tags        []Tag        `gorm:"many2many:topic_tags;" json:"tags,omitempty"`
}

// TableName 指定表名
func (Topic) TableName() string {
	return "topics"
}

// Comment 评论模型
type Comment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	TopicID   uint           `gorm:"not null;index" json:"topic_id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	ParentID  *uint          `gorm:"index" json:"parent_id"` // 回复评论的ID
	Content   string         `gorm:"type:text;not null" json:"content"`
	LikeCount int            `gorm:"default:0" json:"like_count"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	Topic   Topic     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	User    User      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
	Parent  *Comment  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"parent,omitempty"`
	Replies []Comment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}

// TableName 指定表名
func (Comment) TableName() string {
	return "comments"
}

// Assignment 作业模型
type Assignment struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	TopicID     uint       `gorm:"not null;index" json:"topic_id"`
	StudentID   uint       `gorm:"not null;index" json:"student_id"`
	ProjectID   int        `gorm:"not null" json:"project_id"`            // GitLab项目ID
	Status      string     `gorm:"size:20;default:pending" json:"status"` // pending, submitted, graded
	SubmittedAt *time.Time `json:"submitted_at"`
	Grade       *float64   `json:"grade"`
	MaxGrade    float64    `gorm:"default:100" json:"max_grade"`
	Feedback    string     `gorm:"type:text" json:"feedback"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// 关联关系
	Topic   Topic `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"topic,omitempty"`
	Student User  `gorm:"foreignKey:StudentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"student,omitempty"`
}

// TableName 指定表名
func (Assignment) TableName() string {
	return "assignments"
}

// IsSubmitted 检查作业是否已提交
func (a *Assignment) IsSubmitted() bool {
	return a.Status == "submitted" || a.Status == "graded"
}

// IsGraded 检查作业是否已评分
func (a *Assignment) IsGraded() bool {
	return a.Status == "graded" && a.Grade != nil
}

// Attachment 附件模型
type Attachment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TopicID   uint      `gorm:"not null;index" json:"topic_id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	FileName  string    `gorm:"not null;size:255" json:"file_name"`
	FilePath  string    `gorm:"not null;size:500" json:"file_path"`
	FileSize  int64     `gorm:"not null" json:"file_size"`
	MimeType  string    `gorm:"size:100" json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`

	// 关联关系
	Topic Topic `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	User  User  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
}

// TableName 指定表名
func (Attachment) TableName() string {
	return "attachments"
}

// TopicLike 话题点赞记录
type TopicLike struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TopicID   uint      `gorm:"not null;index" json:"topic_id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`

	// 关联关系
	Topic Topic `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	User  User  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

// TableName 指定表名
func (TopicLike) TableName() string {
	return "topic_likes"
}

// CommentLike 评论点赞记录
type CommentLike struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CommentID uint      `gorm:"not null;index" json:"comment_id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`

	// 关联关系
	Comment Comment `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	User    User    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

// TableName 指定表名
func (CommentLike) TableName() string {
	return "comment_likes"
}

// TopicView 话题浏览记录
type TopicView struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	TopicID  uint      `gorm:"not null;index" json:"topic_id"`
	UserID   uint      `gorm:"not null;index" json:"user_id"`
	ViewedAt time.Time `gorm:"not null" json:"viewed_at"`

	// 关联关系
	Topic Topic `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	User  User  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

// TableName 指定表名
func (TopicView) TableName() string {
	return "topic_views"
}
