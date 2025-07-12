package models

import (
	"time"
)

// Project 课题模型
type Project struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `json:"description"`
	Code        string    `gorm:"unique;not null" json:"code"`    // 课题代码，用于学生加入
	TeacherID   uint      `gorm:"not null" json:"teacher_id"`     // 创建课题的老师ID
	ClassID     uint      `json:"class_id"`                       // 所属班级ID（可选）
	Status      string    `gorm:"default:'active'" json:"status"` // active, completed, archived
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联关系
	Teacher     User            `gorm:"foreignKey:TeacherID" json:"teacher,omitempty"`
	Class       Class           `gorm:"foreignKey:ClassID" json:"class,omitempty"`
	Members     []ProjectMember `gorm:"foreignKey:ProjectID" json:"members,omitempty"`
	Students    []User          `gorm:"many2many:project_members;" json:"students,omitempty"`
	Assignments []Assignment    `gorm:"foreignKey:ProjectID" json:"assignments,omitempty"`
}

// TableName 指定表名
func (Project) TableName() string {
	return "projects"
}

// ProjectMember 课题成员关系
type ProjectMember struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID uint      `gorm:"not null" json:"project_id"`
	StudentID uint      `gorm:"not null" json:"student_id"`
	Role      string    `gorm:"default:'member'" json:"role"` // member, leader
	JoinedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"joined_at"`
	Status    string    `gorm:"default:'active'" json:"status"` // active, inactive

	// 关联关系
	Project Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Student User    `gorm:"foreignKey:StudentID" json:"student,omitempty"`
}

// TableName 指定表名
func (ProjectMember) TableName() string {
	return "project_members"
}

// ProjectStats 课题统计信息
type ProjectStats struct {
	MemberCount          int `json:"member_count"`
	AssignmentCount      int `json:"assignment_count"`
	CompletedAssignments int `json:"completed_assignments"`
	PendingAssignments   int `json:"pending_assignments"`
}
