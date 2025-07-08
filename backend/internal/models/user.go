package models

import (
	"time"

	"gorm.io/gorm"
)

// 教育角色枚举 - 基于GitLab权限映射
type EducationRole int

const (
	EduRoleGuest     EducationRole = 10 // GitLab Guest -> 访客
	EduRoleStudent   EducationRole = 20 // GitLab Reporter -> 学生
	EduRoleAssistant EducationRole = 30 // GitLab Developer -> 助教
	EduRoleTeacher   EducationRole = 40 // GitLab Maintainer -> 教师
	EduRoleAdmin     EducationRole = 50 // GitLab Owner -> 管理员
)

// String 返回角色的字符串表示
func (r EducationRole) String() string {
	switch r {
	case EduRoleGuest:
		return "guest"
	case EduRoleStudent:
		return "student"
	case EduRoleAssistant:
		return "assistant"
	case EduRoleTeacher:
		return "teacher"
	case EduRoleAdmin:
		return "admin"
	default:
		return "unknown"
	}
}

// User 极简用户模型 - 只存储GitLab用户映射信息
type User struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	GitLabID   int       `gorm:"column:gitlab_id;unique;not null" json:"gitlab_id"`
	Username   string    `gorm:"unique;not null" json:"username"`
	Email      string    `gorm:"unique;not null" json:"email"`
	Name       string    `gorm:"not null" json:"name"`
	Avatar     string    `json:"avatar"`
	Role       int       `gorm:"default:2" json:"role"` // 1:访客, 2:学生, 3:助教, 4:教师, 5:管理员
	Active     bool      `gorm:"default:true" json:"is_active"`
	LastSyncAt time.Time `gorm:"column:last_sync_at" json:"last_sync_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// BeforeCreate GORM钩子 - 创建前
func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.LastSyncAt = time.Now()
	return nil
}

// BeforeUpdate GORM钩子 - 更新前
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.LastSyncAt = time.Now()
	return nil
}

// IsActive 检查用户是否活跃（最近同步时间在24小时内）
func (u *User) IsActive() bool {
	return time.Since(u.LastSyncAt) < 24*time.Hour
}

// GetDefaultEducationRole 获取默认教育角色（实际角色需要通过GitLab API获取）
func (u *User) GetDefaultEducationRole() EducationRole {
	// 这个方法的实际逻辑在 UserService 中实现
	// 因为需要调用GitLab API获取用户在特定Group/Project中的权限
	return EduRoleGuest
}
