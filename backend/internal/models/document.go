package models

import (
	"time"
)

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
