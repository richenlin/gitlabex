package models

import (
	"time"

	"gorm.io/gorm"
)

// Document 文档模型
type Document struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	ProjectID  int            `gorm:"not null;index" json:"project_id"`
	Slug       string         `gorm:"not null;size:255;index" json:"slug"`
	Title      string         `gorm:"not null;size:255" json:"title"`
	Content    string         `gorm:"type:text" json:"content"`
	Format     string         `gorm:"size:50;default:markdown" json:"format"`
	FileName   string         `gorm:"size:255" json:"file_name"`
	FilePath   string         `gorm:"size:500" json:"file_path"`
	FileSize   int64          `gorm:"default:0" json:"file_size"`
	CategoryID *uint          `gorm:"index" json:"category_id"`
	IsPublic   bool           `gorm:"default:true" json:"is_public"`
	ViewCount  int            `gorm:"default:0" json:"view_count"`
	CreatedBy  uint           `gorm:"not null;index" json:"created_by"`
	UpdatedBy  uint           `gorm:"index" json:"updated_by"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	Category *DocumentCategory `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"category,omitempty"`
	Creator  *User             `gorm:"foreignKey:CreatedBy;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"creator,omitempty"`
	Updater  *User             `gorm:"foreignKey:UpdatedBy;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"updater,omitempty"`
	Tags     []Tag             `gorm:"many2many:document_tags;" json:"tags,omitempty"`
	Sessions []DocumentSession `gorm:"foreignKey:DocumentID" json:"-"`
	Comments []DocumentComment `gorm:"foreignKey:DocumentID" json:"comments,omitempty"`
}

// TableName 指定表名
func (Document) TableName() string {
	return "documents"
}

// DocumentCategory 文档分类
type DocumentCategory struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null;size:255" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	ParentID    *uint     `gorm:"index" json:"parent_id"`
	Sort        int       `gorm:"default:0" json:"sort"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联关系
	Parent    *DocumentCategory  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"parent,omitempty"`
	Children  []DocumentCategory `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Documents []Document         `gorm:"foreignKey:CategoryID" json:"-"`
}

// TableName 指定表名
func (DocumentCategory) TableName() string {
	return "document_categories"
}

// Tag 标签
type Tag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"unique;not null;size:100" json:"name"`
	Color     string    `gorm:"size:7;default:#007bff" json:"color"` // 十六进制颜色
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联关系
	Documents []Document `gorm:"many2many:document_tags;" json:"-"`
}

// TableName 指定表名
func (Tag) TableName() string {
	return "tags"
}

// DocumentSession OnlyOffice文档会话
type DocumentSession struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	DocumentID uint      `gorm:"not null;index" json:"document_id"`
	UserID     uint      `gorm:"not null;index" json:"user_id"`
	Key        string    `gorm:"unique;not null;size:255" json:"key"`
	Mode       string    `gorm:"size:20;default:edit" json:"mode"` // edit, view, comment
	ExpiresAt  time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// 关联关系
	Document Document `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"document,omitempty"`
	User     User     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
}

// TableName 指定表名
func (DocumentSession) TableName() string {
	return "document_sessions"
}

// IsExpired 检查会话是否过期
func (s *DocumentSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// DocumentComment 文档评论
type DocumentComment struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	DocumentID uint           `gorm:"not null;index" json:"document_id"`
	UserID     uint           `gorm:"not null;index" json:"user_id"`
	ParentID   *uint          `gorm:"index" json:"parent_id"` // 回复评论的ID
	Content    string         `gorm:"type:text;not null" json:"content"`
	IsResolved bool           `gorm:"default:false" json:"is_resolved"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	Document Document          `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	User     User              `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
	Parent   *DocumentComment  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"parent,omitempty"`
	Replies  []DocumentComment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}

// TableName 指定表名
func (DocumentComment) TableName() string {
	return "document_comments"
}

// DocumentVersion 文档版本
type DocumentVersion struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	DocumentID uint      `gorm:"not null;index" json:"document_id"`
	Version    string    `gorm:"not null;size:100" json:"version"`
	Title      string    `gorm:"not null;size:255" json:"title"`
	Content    string    `gorm:"type:text" json:"content"`
	ChangeLog  string    `gorm:"type:text" json:"change_log"`
	CreatedBy  uint      `gorm:"not null;index" json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`

	// 关联关系
	Document Document `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Creator  User     `gorm:"foreignKey:CreatedBy;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"creator,omitempty"`
}

// TableName 指定表名
func (DocumentVersion) TableName() string {
	return "document_versions"
}

// DocumentPermission 文档权限
type DocumentPermission struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	DocumentID uint      `gorm:"not null;index" json:"document_id"`
	UserID     *uint     `gorm:"index" json:"user_id"`               // 用户ID，为空表示所有用户
	TeamID     *uint     `gorm:"index" json:"team_id"`               // 团队ID
	Permission string    `gorm:"not null;size:20" json:"permission"` // read, write, admin
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// 关联关系
	Document Document `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	User     *User    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
	Team     *Team    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"team,omitempty"`
}

// TableName 指定表名
func (DocumentPermission) TableName() string {
	return "document_permissions"
}
