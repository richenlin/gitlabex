package models

import (
	"time"
)

// Document 文档模型 - 支持Wiki和文档管理
type Document struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Title     string `gorm:"not null" json:"title"`
	Content   string `gorm:"type:text" json:"content"`
	ProjectID uint   `gorm:"not null" json:"project_id"`
	AuthorID  uint   `gorm:"not null" json:"author_id"`
	Type      string `gorm:"not null;default:'wiki'" json:"type"`       // wiki, markdown, office
	Format    string `gorm:"not null;default:'markdown'" json:"format"` // markdown, html
	IsPublic  bool   `gorm:"default:true" json:"is_public"`
	ParentID  uint   `gorm:"default:0" json:"parent_id"`
	Status    string `gorm:"not null;default:'active'" json:"status"` // active, deleted, archived
	Version   int    `gorm:"default:1" json:"version"`

	// GitLab Wiki相关字段
	GitLabWikiSlug string `json:"gitlab_wiki_slug"`
	GitLabWikiURL  string `json:"gitlab_wiki_url"`

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`

	// 关联关系
	Author   User       `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Project  Project    `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Parent   *Document  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Document `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

// TableName 指定表名
func (Document) TableName() string {
	return "documents"
}

// DocumentHistory 文档历史记录
type DocumentHistory struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	DocumentID uint      `gorm:"not null" json:"document_id"`
	AuthorID   uint      `gorm:"not null" json:"author_id"`
	Content    string    `gorm:"type:text" json:"content"`
	ChangeNote string    `json:"change_note"`
	CreatedAt  time.Time `json:"created_at"`

	// 关联关系
	Document Document `gorm:"foreignKey:DocumentID" json:"document,omitempty"`
	Author   User     `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
}

// TableName 指定表名
func (DocumentHistory) TableName() string {
	return "document_histories"
}

// DocumentAttachment 文档附件模型 - 只存储OnlyOffice编辑会话信息
type DocumentAttachment struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	UserID       int        `gorm:"not null" json:"user_id"`
	ProjectID    *int       `gorm:"default:null" json:"project_id"`
	WikiPageSlug string     `gorm:"default:null" json:"wiki_page_slug"`
	FileName     string     `gorm:"not null" json:"file_name"`
	FileURL      string     `gorm:"default:null" json:"file_url"`
	FilePath     string     `gorm:"not null" json:"file_path"`
	FileType     string     `gorm:"not null" json:"file_type"` // docx, xlsx, pptx
	DocumentKey  string     `gorm:"unique;not null" json:"document_key"`
	EditMode     string     `gorm:"not null;default:'edit'" json:"edit_mode"` // edit, view
	Status       string     `gorm:"not null;default:'editing'" json:"status"` // editing, saving, closed, error
	LastEditedBy *uint      `gorm:"default:null" json:"last_edited_by"`
	LastEditedAt *time.Time `gorm:"default:null" json:"last_edited_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// 关联关系
	LastEditor *User `gorm:"foreignKey:LastEditedBy;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"last_editor,omitempty"`
}

// TableName 指定表名
func (DocumentAttachment) TableName() string {
	return "document_attachments"
}

// IsEditableType 检查是否是可编辑的文件类型
func (d *DocumentAttachment) IsEditableType() bool {
	switch d.FileType {
	case "docx", "xlsx", "pptx":
		return true
	default:
		return false
	}
}

// GetDocumentType 获取OnlyOffice文档类型
func (d *DocumentAttachment) GetDocumentType() string {
	switch d.FileType {
	case "docx":
		return "text"
	case "xlsx":
		return "spreadsheet"
	case "pptx":
		return "presentation"
	default:
		return "text"
	}
}

// MarkAsEdited 标记为已编辑
func (d *DocumentAttachment) MarkAsEdited(userID uint) {
	now := time.Now()
	d.LastEditedBy = &userID
	d.LastEditedAt = &now
}

// DocumentSummary 文档摘要信息 - 用于API响应
type DocumentSummary struct {
	ID           uint       `json:"id"`
	UserID       int        `json:"user_id"`
	ProjectID    *int       `json:"project_id"`
	Title        string     `json:"title"` // Wiki页面标题
	FileName     string     `json:"file_name"`
	FileType     string     `json:"file_type"`
	FileURL      string     `json:"file_url"`
	FilePath     string     `json:"file_path"`
	WikiSlug     string     `json:"wiki_slug"`
	EditMode     string     `json:"edit_mode"`
	Status       string     `json:"status"`
	LastEditedBy *uint      `json:"last_edited_by"`
	LastEditedAt *time.Time `json:"last_edited_at"`
	CanEdit      bool       `json:"can_edit"`
	LastEditor   *User      `json:"last_editor,omitempty"`
}

// ToSummary 转换为摘要信息
func (d *DocumentAttachment) ToSummary(title string, canEdit bool) *DocumentSummary {
	return &DocumentSummary{
		ID:           d.ID,
		UserID:       d.UserID,
		ProjectID:    d.ProjectID,
		Title:        title,
		FileName:     d.FileName,
		FileType:     d.FileType,
		FileURL:      d.FileURL,
		FilePath:     d.FilePath,
		WikiSlug:     d.WikiPageSlug,
		EditMode:     d.EditMode,
		Status:       d.Status,
		LastEditedBy: d.LastEditedBy,
		LastEditedAt: d.LastEditedAt,
		CanEdit:      canEdit,
		LastEditor:   d.LastEditor,
	}
}
