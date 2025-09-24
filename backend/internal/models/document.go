package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// DocumentType 文档类型
type DocumentType string

const (
	DocumentTypePDF      DocumentType = "pdf"
	DocumentTypeWord     DocumentType = "word"
	DocumentTypeExcel    DocumentType = "excel"
	DocumentTypePPT      DocumentType = "ppt"
	DocumentTypeText     DocumentType = "text"
	DocumentTypeMarkdown DocumentType = "markdown"
	DocumentTypeCode     DocumentType = "code"
	DocumentTypeImage    DocumentType = "image"
	DocumentTypeVideo    DocumentType = "video"
	DocumentTypeOther    DocumentType = "other"
)

// DocumentStatus 文档状态
type DocumentStatus string

const (
	DocumentStatusPending  DocumentStatus = "pending"
	DocumentStatusApproved DocumentStatus = "approved"
	DocumentStatusRejected DocumentStatus = "rejected"
)

// Document 文档模型
type Document struct {
	BaseModel
	Title          string         `gorm:"not null;size:200" json:"title"`
	Description    string         `gorm:"type:text" json:"description"`
	FilePath       string         `gorm:"not null;size:500" json:"file_path"`
	FileSize       int64          `gorm:"default:0" json:"file_size"`
	FileType       DocumentType   `gorm:"not null;default:other" json:"file_type"`
	MIMEType       string         `gorm:"size:100" json:"mime_type"`
	Status         DocumentStatus `gorm:"not null;default:pending" json:"status"`
	UploaderID     int64          `gorm:"not null" json:"uploader_id"`           // 改为int64以匹配gitlab_user_id
	ProjectID      *uuid.UUID     `gorm:"type:uuid" json:"project_id,omitempty"` // 改为可选，支持独立文档
	Category       string         `gorm:"size:100" json:"category"`
	Tags           pq.StringArray `gorm:"type:text[]" json:"tags"`
	DownloadCount  int            `gorm:"default:0" json:"download_count"`
	GitLabFilePath string         `gorm:"size:500" json:"gitlab_file_path,omitempty"`
	GitLabBranch   string         `gorm:"size:100;default:main" json:"gitlab_branch,omitempty"`
	GitLabID       string         `gorm:"size:100" json:"gitlab_id,omitempty"`
	AutoIndexed    bool           `gorm:"default:false" json:"auto_indexed,omitempty"`
	IsStandalone   bool           `gorm:"default:false" json:"is_standalone,omitempty"` // 标识独立文档
	LastSyncTime   *time.Time     `gorm:"column:last_sync_time" json:"last_sync_time,omitempty"`
	MinIOPath      string         `gorm:"column:minio_path;size:500" json:"minio_path,omitempty"` // MinIO存储路径

	// 关联关系
	// 注意：Uploader关联已移除，上传者信息从GitLab API获取
	Project *ResearchProject `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Reviews []DocumentReview `gorm:"foreignKey:DocumentID" json:"reviews,omitempty"`
}

// DocumentReview 文档审核模型
type DocumentReview struct {
	BaseModel
	DocumentID uuid.UUID      `gorm:"not null" json:"document_id"`
	ReviewerID int64          `gorm:"not null" json:"reviewer_id"` // 改为int64以匹配gitlab_user_id
	Status     DocumentStatus `gorm:"not null" json:"status"`
	Comments   string         `gorm:"type:text" json:"comments"`
	CreatedAt  time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"column:updated_at" json:"updated_at"`

	// 关联关系
	Document Document `gorm:"foreignKey:DocumentID" json:"document,omitempty"`
	// 注意：Reviewer关联已移除，审核者信息从GitLab API获取
}

// DocumentEditRequest 文档编辑请求模型（用于学生提交的修改请求）
type DocumentEditRequest struct {
	BaseModel
	DocumentID     uuid.UUID      `gorm:"not null" json:"document_id"`
	RequesterID    int64          `gorm:"not null" json:"requester_id"` // 改为int64以匹配gitlab_user_id
	Title          string         `gorm:"size:200" json:"title,omitempty"`
	Description    string         `gorm:"type:text" json:"description,omitempty"`
	Category       string         `gorm:"size:100" json:"category,omitempty"`
	Tags           pq.StringArray `gorm:"type:text[]" json:"tags,omitempty"`
	Reason         string         `gorm:"type:text" json:"reason"` // 修改原因
	Status         DocumentStatus `gorm:"not null;default:pending" json:"status"`
	ReviewerID     *int64         `gorm:"type:bigint" json:"reviewer_id,omitempty"` // 改为int64以匹配gitlab_user_id
	ReviewComments string         `gorm:"type:text" json:"review_comments,omitempty"`
	ReviewedAt     *time.Time     `json:"reviewed_at,omitempty"`

	// 关联关系
	Document Document `gorm:"foreignKey:DocumentID" json:"document,omitempty"`
	// 注意：Requester和Reviewer关联已移除，用户信息从GitLab API获取
}
