package models

import (
	"time"
)

// Class 班级模型
type Class struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `json:"description"`
	Code        string    `gorm:"unique;not null" json:"code"` // 班级代码，用于学生加入
	TeacherID   uint      `gorm:"not null" json:"teacher_id"`  // 创建班级的老师ID
	Teacher     User      `gorm:"foreignKey:TeacherID" json:"teacher,omitempty"`
	Active      bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联关系
	Members  []ClassMember `gorm:"foreignKey:ClassID" json:"members,omitempty"`
	Students []User        `gorm:"many2many:class_members;" json:"students,omitempty"`
}

// TableName 指定表名
func (Class) TableName() string {
	return "classes"
}

// ClassMember 班级成员关系
type ClassMember struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ClassID   uint      `gorm:"not null" json:"class_id"`
	StudentID uint      `gorm:"not null" json:"student_id"`
	JoinedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"joined_at"`
	Status    string    `gorm:"default:'active'" json:"status"` // active, inactive

	// 关联关系
	Class   Class `gorm:"foreignKey:ClassID" json:"class,omitempty"`
	Student User  `gorm:"foreignKey:StudentID" json:"student,omitempty"`
}

// TableName 指定表名
func (ClassMember) TableName() string {
	return "class_members"
}

// ClassStats 班级统计信息
type ClassStats struct {
	StudentCount   int `json:"student_count"`
	ProjectCount   int `json:"project_count"`
	ActiveProjects int `json:"active_projects"`
}
