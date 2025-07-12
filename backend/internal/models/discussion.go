package models

import (
	"time"
)

// Discussion 话题讨论模型 - 基于GitLab Issues实现
type Discussion struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Title     string `gorm:"not null" json:"title"`
	Content   string `gorm:"type:text" json:"content"`
	ProjectID uint   `gorm:"not null" json:"project_id"`
	AuthorID  uint   `gorm:"not null" json:"author_id"`

	// GitLab Issue相关字段
	GitLabIssueID  int    `gorm:"not null;unique" json:"gitlab_issue_id"`
	GitLabIssueURL string `json:"gitlab_issue_url"`

	// 讨论状态和属性
	Status   string `gorm:"not null;default:'open'" json:"status"`      // open, closed
	Priority string `gorm:"not null;default:'normal'" json:"priority"`  // low, normal, high
	Category string `gorm:"not null;default:'general'" json:"category"` // general, question, announcement
	Tags     string `json:"tags"`                                       // 标签，逗号分隔

	// 统计信息
	ViewCount  int `gorm:"default:0" json:"view_count"`
	ReplyCount int `gorm:"default:0" json:"reply_count"`
	LikeCount  int `gorm:"default:0" json:"like_count"`

	// 权限控制
	IsPublic bool `gorm:"default:true" json:"is_public"`
	IsPinned bool `gorm:"default:false" json:"is_pinned"`

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`

	// 关联关系
	Author  User    `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Project Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
}

// TableName 指定表名
func (Discussion) TableName() string {
	return "discussions"
}

// DiscussionReply 话题回复模型 - 基于GitLab Issue Notes实现
type DiscussionReply struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	DiscussionID uint   `gorm:"not null" json:"discussion_id"`
	AuthorID     uint   `gorm:"not null" json:"author_id"`
	Content      string `gorm:"type:text;not null" json:"content"`

	// GitLab Note相关字段
	GitLabNoteID  int    `gorm:"not null;unique" json:"gitlab_note_id"`
	GitLabNoteURL string `json:"gitlab_note_url"`

	// 回复属性
	ParentReplyID uint `gorm:"default:0" json:"parent_reply_id"` // 支持嵌套回复
	IsResolved    bool `gorm:"default:false" json:"is_resolved"`

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`

	// 关联关系
	Discussion  Discussion       `gorm:"foreignKey:DiscussionID" json:"discussion,omitempty"`
	Author      User             `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	ParentReply *DiscussionReply `gorm:"foreignKey:ParentReplyID" json:"parent_reply,omitempty"`
}

// TableName 指定表名
func (DiscussionReply) TableName() string {
	return "discussion_replies"
}

// DiscussionView 话题浏览记录模型
type DiscussionView struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	DiscussionID uint      `gorm:"not null" json:"discussion_id"`
	UserID       uint      `gorm:"not null" json:"user_id"`
	ViewedAt     time.Time `gorm:"not null" json:"viewed_at"`

	// 关联关系
	Discussion Discussion `gorm:"foreignKey:DiscussionID" json:"discussion,omitempty"`
	User       User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (DiscussionView) TableName() string {
	return "discussion_views"
}

// DiscussionLike 话题点赞模型
type DiscussionLike struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	DiscussionID uint      `gorm:"not null" json:"discussion_id"`
	UserID       uint      `gorm:"not null" json:"user_id"`
	CreatedAt    time.Time `json:"created_at"`

	// 关联关系
	Discussion Discussion `gorm:"foreignKey:DiscussionID" json:"discussion,omitempty"`
	User       User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (DiscussionLike) TableName() string {
	return "discussion_likes"
}

// DiscussionCreateRequest 创建话题请求
type DiscussionCreateRequest struct {
	Title     string `json:"title" binding:"required"`
	Content   string `json:"content" binding:"required"`
	ProjectID uint   `json:"project_id" binding:"required"`
	Category  string `json:"category"`
	Tags      string `json:"tags"`
	IsPublic  bool   `json:"is_public"`
}

// DiscussionUpdateRequest 更新话题请求
type DiscussionUpdateRequest struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	Category string `json:"category"`
	Tags     string `json:"tags"`
	IsPublic bool   `json:"is_public"`
}

// DiscussionReplyRequest 回复话题请求
type DiscussionReplyRequest struct {
	Content       string `json:"content" binding:"required"`
	ParentReplyID uint   `json:"parent_reply_id"`
}

// DiscussionListResponse 话题列表响应
type DiscussionListResponse struct {
	Total       int64        `json:"total"`
	Page        int          `json:"page"`
	PageSize    int          `json:"page_size"`
	Discussions []Discussion `json:"discussions"`
}

// DiscussionDetailResponse 话题详情响应
type DiscussionDetailResponse struct {
	Discussion Discussion        `json:"discussion"`
	Replies    []DiscussionReply `json:"replies"`
	IsLiked    bool              `json:"is_liked"`
	CanEdit    bool              `json:"can_edit"`
	CanDelete  bool              `json:"can_delete"`
}
