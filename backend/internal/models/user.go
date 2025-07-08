package models

import (
	"time"

	"gorm.io/gorm"
)

// UserRole 用户角色枚举
type UserRole int

const (
	RoleStudent UserRole = iota // 学生
	RoleTeacher                 // 教师
	RoleAdmin                   // 管理员
)

// String 返回角色的字符串表示
func (r UserRole) String() string {
	switch r {
	case RoleStudent:
		return "student"
	case RoleTeacher:
		return "teacher"
	case RoleAdmin:
		return "admin"
	default:
		return "unknown"
	}
}

// User 用户模型
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	GitLabID  int            `gorm:"unique;not null" json:"gitlab_id"`
	Username  string         `gorm:"unique;not null;size:255" json:"username"`
	Email     string         `gorm:"unique;not null;size:255" json:"email"`
	Name      string         `gorm:"not null;size:255" json:"name"`
	Role      UserRole       `gorm:"not null;default:0" json:"role"`
	Avatar    string         `gorm:"size:500" json:"avatar"`
	Bio       string         `gorm:"type:text" json:"bio"`
	Location  string         `gorm:"size:255" json:"location"`
	Website   string         `gorm:"size:500" json:"website"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	LastLogin *time.Time     `json:"last_login"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	CreatedDocuments []Document `gorm:"foreignKey:CreatedBy" json:"-"`
	UpdatedDocuments []Document `gorm:"foreignKey:UpdatedBy" json:"-"`
	CreatedTopics    []Topic    `gorm:"foreignKey:CreatedBy" json:"-"`
	OwnedProjects    []Project  `gorm:"foreignKey:OwnerID" json:"-"`
	LedTeams         []Team     `gorm:"foreignKey:LeaderID" json:"-"`

	// 多对多关系
	Teams            []Team            `gorm:"many2many:team_members;" json:"-"`
	Projects         []Project         `gorm:"many2many:project_members;" json:"-"`
	DocumentSessions []DocumentSession `gorm:"foreignKey:UserID" json:"-"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// IsTeacher 检查是否是教师
func (u *User) IsTeacher() bool {
	return u.Role == RoleTeacher || u.Role == RoleAdmin
}

// IsAdmin 检查是否是管理员
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// CanEdit 检查是否有编辑权限
func (u *User) CanEdit(targetUserID uint) bool {
	return u.IsAdmin() || u.ID == targetUserID
}

// UserProfile 用户个人资料
type UserProfile struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	Avatar   string `json:"avatar"`
	Bio      string `json:"bio"`
	Location string `json:"location"`
	Website  string `json:"website"`
}

// ToProfile 转换为用户资料
func (u *User) ToProfile() *UserProfile {
	return &UserProfile{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		Name:     u.Name,
		Role:     u.Role.String(),
		Avatar:   u.Avatar,
		Bio:      u.Bio,
		Location: u.Location,
		Website:  u.Website,
	}
}

// UserSession 用户会话
type UserSession struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Token     string    `gorm:"unique;not null;size:500" json:"-"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联关系
	User User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user"`
}

// TableName 指定表名
func (UserSession) TableName() string {
	return "user_sessions"
}

// IsExpired 检查会话是否过期
func (s *UserSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// UserPreference 用户偏好设置
type UserPreference struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	UserID   uint   `gorm:"not null;uniqueIndex" json:"user_id"`
	Language string `gorm:"size:10;default:zh-CN" json:"language"`
	Timezone string `gorm:"size:50;default:Asia/Shanghai" json:"timezone"`
	Theme    string `gorm:"size:20;default:light" json:"theme"` // light, dark

	// 通知设置
	EmailNotifications bool `gorm:"default:true" json:"email_notifications"`
	PushNotifications  bool `gorm:"default:true" json:"push_notifications"`

	// 编辑器设置
	EditorFontSize int    `gorm:"default:14" json:"editor_font_size"`
	EditorTheme    string `gorm:"size:20;default:vs-dark" json:"editor_theme"`
	EditorAutoSave bool   `gorm:"default:true" json:"editor_auto_save"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联关系
	User User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

// TableName 指定表名
func (UserPreference) TableName() string {
	return "user_preferences"
}
