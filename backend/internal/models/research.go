package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// ResearchProject 研究课题模型
type ResearchProject struct {
	BaseModel
	Name            string     `gorm:"not null;size:200" json:"name"`
	Description     string     `gorm:"type:text" json:"description"`
	Status          string     `gorm:"not null;default:active" json:"status"` // active, archived, completed
	CreatorID       int64      `gorm:"not null" json:"creator_id"`            // GitLab用户ID
	GitLabProjectID *int64     `gorm:"uniqueIndex" json:"gitlab_project_id,omitempty"`
	GitLabURL       string     `gorm:"size:500" json:"gitlab_url,omitempty"`
	StartDate       time.Time  `json:"start_date"`
	EndDate         *time.Time `json:"end_date,omitempty"`
	IsPublic        bool       `gorm:"default:true" json:"is_public"`

	// 注意：Creator关联已移除，创建者信息从GitLab API获取
	Topics    []Topic    `gorm:"foreignKey:ProjectID" json:"topics,omitempty"`
	Homeworks []Homework `gorm:"foreignKey:ProjectID" json:"homeworks,omitempty"`
	Documents []Document `gorm:"foreignKey:ProjectID" json:"documents,omitempty"`
}

// 注意：ProjectMember模型已移除，成员管理完全使用GitLab项目成员

// Topic 话题模型
type Topic struct {
	BaseModel
	Title         string         `gorm:"not null;size:200" json:"title"`
	Content       string         `gorm:"type:text" json:"content"`
	ProjectID     *uuid.UUID     `json:"project_id,omitempty"`
	AuthorID      int64          `gorm:"not null" json:"author_id"` // GitLab用户ID
	GitLabIssueID *int64         `gorm:"uniqueIndex" json:"gitlab_issue_id,omitempty"`
	Status        string         `gorm:"not null;default:active" json:"status"`   // active, closed, archived
	Priority      string         `gorm:"not null;default:normal" json:"priority"` // low, normal, high, urgent
	Tags          pq.StringArray `gorm:"type:text[]" json:"tags"`
	ViewCount     int            `gorm:"default:0" json:"view_count"`
	LikeCount     int            `gorm:"default:0" json:"like_count"`

	// 关联关系
	Project ResearchProject `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	// 注意：Author关联已移除，作者信息从GitLab API获取
	Comments   []Comment   `gorm:"foreignKey:TopicID" json:"comments,omitempty"`
	TopicLikes []TopicLike `gorm:"foreignKey:TopicID" json:"topic_likes,omitempty"`
}

// Comment 评论模型
type Comment struct {
	BaseModel
	Content      string     `gorm:"type:text;not null" json:"content"`
	TopicID      uuid.UUID  `gorm:"not null" json:"topic_id"`
	AuthorID     int64      `gorm:"not null" json:"author_id"` // GitLab用户ID
	ParentID     *uuid.UUID `gorm:"index" json:"parent_id,omitempty"`
	GitLabNoteID *int64     `gorm:"uniqueIndex" json:"gitlab_note_id,omitempty"`

	// 关联关系
	Topic Topic `gorm:"foreignKey:TopicID" json:"topic,omitempty"`
	// 注意：Author关联已移除，作者信息从GitLab API获取
	Parent  *Comment  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Replies []Comment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}

// TopicLike 话题点赞模型
type TopicLike struct {
	BaseModel
	TopicID uuid.UUID `gorm:"not null;uniqueIndex:idx_topic_user" json:"topic_id"`
	UserID  int64     `gorm:"not null;uniqueIndex:idx_topic_user" json:"user_id"` // GitLab用户ID

	// 关联关系
	Topic Topic `gorm:"foreignKey:TopicID" json:"topic,omitempty"`
	// 注意：User关联已移除，用户信息从GitLab API获取
}
