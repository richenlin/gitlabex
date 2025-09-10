package models

import (
	"time"

	"github.com/google/uuid"
)

// ResearchProject 研究课题模型
type ResearchProject struct {
	BaseModel
	Name            string     `gorm:"not null;size:200" json:"name"`
	Description     string     `gorm:"type:text" json:"description"`
	Status          string     `gorm:"not null;default:active" json:"status"` // active, archived, completed
	CreatorID       uuid.UUID  `gorm:"not null" json:"creator_id"`
	GitLabProjectID *int64     `gorm:"uniqueIndex" json:"gitlab_project_id,omitempty"`
	GitLabURL       string     `gorm:"size:500" json:"gitlab_url,omitempty"`
	StartDate       time.Time  `json:"start_date"`
	EndDate         *time.Time `json:"end_date,omitempty"`
	IsPublic        bool       `gorm:"default:true" json:"is_public"`

	// 关联关系
	Creator   User            `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	Members   []ProjectMember `gorm:"foreignKey:ProjectID" json:"members,omitempty"`
	Topics    []Topic         `gorm:"foreignKey:ProjectID" json:"topics,omitempty"`
	Homeworks []Homework      `gorm:"foreignKey:ProjectID" json:"homeworks,omitempty"`
	Documents []Document      `gorm:"foreignKey:ProjectID" json:"documents,omitempty"`
}

// ProjectMember 项目成员关系模型
type ProjectMember struct {
	BaseModel
	ProjectID uuid.UUID   `gorm:"not null" json:"project_id"`
	UserID    uuid.UUID   `gorm:"not null" json:"user_id"`
	Role      ProjectRole `gorm:"not null;default:reporter" json:"role"`
	JoinedAt  time.Time   `gorm:"default:CURRENT_TIMESTAMP" json:"joined_at"`

	// 关联关系
	Project ResearchProject `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	User    User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// Topic 话题模型
type Topic struct {
	BaseModel
	Title         string     `gorm:"not null;size:200" json:"title"`
	Content       string     `gorm:"type:text" json:"content"`
	ProjectID     *uuid.UUID `json:"project_id,omitempty"`
	AuthorID      uuid.UUID  `gorm:"not null" json:"author_id"`
	GitLabIssueID *int64     `gorm:"uniqueIndex" json:"gitlab_issue_id,omitempty"`
	Status        string     `gorm:"not null;default:active" json:"status"`   // active, closed, archived
	Priority      string     `gorm:"not null;default:normal" json:"priority"` // low, normal, high, urgent
	Tags          []string   `gorm:"type:text[]" json:"tags"`
	ViewCount     int        `gorm:"default:0" json:"view_count"`
	LikeCount     int        `gorm:"default:0" json:"like_count"`

	// 关联关系
	Project    ResearchProject `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Author     User            `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Comments   []Comment       `gorm:"foreignKey:TopicID" json:"comments,omitempty"`
	TopicLikes []TopicLike     `gorm:"foreignKey:TopicID" json:"topic_likes,omitempty"`
}

// Comment 评论模型
type Comment struct {
	BaseModel
	Content      string     `gorm:"type:text;not null" json:"content"`
	TopicID      uuid.UUID  `gorm:"not null" json:"topic_id"`
	AuthorID     uuid.UUID  `gorm:"not null" json:"author_id"`
	ParentID     *uuid.UUID `gorm:"index" json:"parent_id,omitempty"`
	GitLabNoteID *int64     `gorm:"uniqueIndex" json:"gitlab_note_id,omitempty"`

	// 关联关系
	Topic   Topic     `gorm:"foreignKey:TopicID" json:"topic,omitempty"`
	Author  User      `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Parent  *Comment  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Replies []Comment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}

// TopicLike 话题点赞模型
type TopicLike struct {
	BaseModel
	TopicID uuid.UUID `gorm:"not null;uniqueIndex:idx_topic_user" json:"topic_id"`
	UserID  uuid.UUID `gorm:"not null;uniqueIndex:idx_topic_user" json:"user_id"`

	// 关联关系
	Topic Topic `gorm:"foreignKey:TopicID" json:"topic,omitempty"`
	User  User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
